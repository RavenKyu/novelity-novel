package auth

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"unicode"

	"github.com/labstack/echo/v5"
)

// requestSpace matches only configured hosts, never user-controlled forwarded headers.
func (s *Service) requestSpace(c *echo.Context) (string, string, error) {
	for _, space := range s.cfg.Surfaces {
		origin := s.cfg.Origins[space]
		u, err := url.Parse(origin)
		if err == nil && strings.EqualFold(c.Request().Host, u.Host) {
			return space, origin, nil
		}
	}
	return "", "", fmt.Errorf("unknown browser origin")
}

func (s *Service) validReturnTo(raw, origin string) (string, error) {
	if raw == "" {
		return "/", nil
	}
	decoded, err := url.PathUnescape(raw)
	if err != nil {
		return "", err
	}
	for _, value := range []string{raw, decoded} {
		if strings.Contains(value, "\\") || strings.IndexFunc(value, func(r rune) bool { return unicode.IsControl(r) || unicode.IsSpace(r) }) >= 0 {
			return "", fmt.Errorf("unsafe returnTo")
		}
	}
	u, err := url.Parse(raw)
	if err != nil || u.User != nil || u.Opaque != "" {
		return "", fmt.Errorf("invalid returnTo")
	}
	if !u.IsAbs() {
		if !strings.HasPrefix(raw, "/") || strings.HasPrefix(raw, "//") || u.Host != "" || strings.HasPrefix(decoded, "//") {
			return "", fmt.Errorf("invalid relative returnTo")
		}
		return raw, nil
	}
	allowed := false
	for _, space := range s.cfg.Surfaces {
		candidate := s.cfg.Origins[space]
		if u.Scheme+"://"+u.Host == candidate {
			allowed = true
			break
		}
	}
	if !allowed {
		return "", fmt.Errorf("returnTo origin is not allowed")
	}
	// origin is the initiating, trusted origin; absolute returns never select the OAuth callback.
	_ = origin
	return u.String(), nil
}

func (s *Service) loginError(c *echo.Context, code, returnTo string) error {
	values := url.Values{"error": {code}}
	if returnTo != "" && returnTo != "/" {
		values.Set("returnTo", returnTo)
	}
	return c.Redirect(http.StatusFound, "/login?"+values.Encode())
}

// SameOrigin protects cookie-authenticated state changes, including same-site sibling hosts.
// Non-browser clients with no Origin/Referer are allowed unless Fetch Metadata says cross-origin.
func (s *Service) SameOrigin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		if method := c.Request().Method; method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions {
			return next(c)
		}
		_, origin, err := s.requestSpace(c)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_origin"})
		}
		supplied := c.Request().Header.Get("Origin")
		if supplied == "" {
			if referer := c.Request().Header.Get("Referer"); referer != "" {
				u, err := url.Parse(referer)
				if err != nil {
					return c.NoContent(http.StatusForbidden)
				}
				supplied = u.Scheme + "://" + u.Host
			} else if site := c.Request().Header.Get("Sec-Fetch-Site"); site == "cross-site" || site == "same-site" {
				return c.NoContent(http.StatusForbidden)
			}
		}
		if supplied != "" && supplied != origin {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "invalid_origin"})
		}
		return next(c)
	}
}
