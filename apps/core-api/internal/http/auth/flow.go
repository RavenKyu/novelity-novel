package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
)

type loginFlow struct {
	State     string `json:"state"`
	Nonce     string `json:"nonce"`
	Verifier  string `json:"verifier"`
	ReturnTo  string `json:"returnTo"`
	Origin    string `json:"origin"`
	ExpiresAt int64  `json:"expiresAt"`
}

func (s *Service) flowName(space string) string {
	if space == "admin" {
		return "nn_admin_oauth"
	}
	return flowCookie
}

func (s *Service) setFlow(c *echo.Context, space string, flow loginFlow) error {
	payload, _ := json.Marshal(flow)
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, s.flowKey)
	mac.Write([]byte(encoded))
	value := encoded + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	cookie := &http.Cookie{Name: s.flowName(space), Value: value, Path: flowPath, MaxAge: int(flowTTL.Seconds()), HttpOnly: true, Secure: strings.HasPrefix(flow.Origin, "https://"), SameSite: http.SameSiteLaxMode}
	if len(cookie.String()) > 4096 {
		return fmt.Errorf("login flow cookie is too large")
	}
	c.SetCookie(cookie)
	return nil
}

// takeFlow clears the host-only cookie and verifies integrity, expiry and initiating origin.
func (s *Service) takeFlow(c *echo.Context, space, origin string) (loginFlow, bool) {
	ck, err := c.Cookie(s.flowName(space))
	c.SetCookie(&http.Cookie{Name: s.flowName(space), Path: flowPath, MaxAge: -1, HttpOnly: true, Secure: strings.HasPrefix(origin, "https://"), SameSite: http.SameSiteLaxMode})
	if err != nil {
		return loginFlow{}, false
	}
	parts := strings.Split(ck.Value, ".")
	if len(parts) != 2 {
		return loginFlow{}, false
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return loginFlow{}, false
	}
	mac := hmac.New(sha256.New, s.flowKey)
	mac.Write([]byte(parts[0]))
	if !hmac.Equal(signature, mac.Sum(nil)) {
		return loginFlow{}, false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return loginFlow{}, false
	}
	var flow loginFlow
	if json.Unmarshal(payload, &flow) != nil || flow.State == "" || flow.Nonce == "" || flow.Verifier == "" || flow.Origin != origin || flow.ExpiresAt <= s.now().Unix() {
		return loginFlow{}, false
	}
	if _, err := s.validReturnTo(flow.ReturnTo); err != nil {
		return loginFlow{}, false
	}
	return flow, true
}
