package config

import "testing"

func TestSurfaceSelection(t *testing.T) {
	for _, value := range []string{"studio,store", "studio", "store", "admin"} {
		if _, err := ParseSurfaces(value); err != nil {
			t.Fatalf("%q: %v", value, err)
		}
	}
	for _, value := range []string{"", "studio,", "studio,admin", "admin,store", "unknown", "store,store"} {
		if _, err := ParseSurfaces(value); err == nil {
			t.Fatalf("accepted %q", value)
		}
	}
}
func TestOriginValidation(t *testing.T) {
	cfg := Config{Origins: map[string]string{"studio": "https://studio.novelity.kr", "store": "https://novelity.kr"}, SessionDomain: "novelity.kr"}
	if err := cfg.Validate([]string{"studio", "store"}); err != nil {
		t.Fatal(err)
	}
	for _, raw := range []string{"https://evil.test", "http://studio.novelity.kr", "https://studio.novelity.kr/path", "https://user@studio.novelity.kr", "https://studio.novelity.kr?x=1"} {
		cfg.Origins["studio"] = raw
		if err := cfg.Validate([]string{"studio", "store"}); err == nil {
			t.Fatalf("accepted %s", raw)
		}
	}
}
