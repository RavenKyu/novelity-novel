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

func TestDuplicateOriginHostsRejected(t *testing.T) {
	cfg := Config{Origins: map[string]string{"studio": "http://localhost:8000", "store": "http://localhost:8000"}}
	if err := cfg.Validate([]string{"studio", "store"}); err == nil {
		t.Fatal("accepted two surfaces on the same host and port")
	}
	for _, store := range []string{"http://LOCALHOST:8000"} {
		cfg.Origins["store"] = store
		if err := cfg.Validate([]string{"studio", "store"}); err == nil {
			t.Fatalf("accepted duplicate host %s", store)
		}
	}
	for _, pair := range [][2]string{{"http://studio.test", "http://studio.test:80"}, {"https://studio.test:443", "https://STUDIO.test"}} {
		cfg.Origins["studio"], cfg.Origins["store"] = pair[0], pair[1]
		if err := cfg.Validate([]string{"studio", "store"}); err == nil {
			t.Fatalf("accepted duplicate host %v", pair)
		}
	}
	cfg.Origins["studio"], cfg.Origins["store"] = "http://localhost:8000", "http://localhost:8001"
	if err := cfg.Validate([]string{"studio", "store"}); err != nil {
		t.Fatal(err)
	}
}

func TestListenerAddress(t *testing.T) {
	t.Setenv("PORT", "8081")
	t.Setenv("BIND_HOST", "127.0.0.1")
	if got := Load().Addr; got != "127.0.0.1:8081" {
		t.Fatalf("loopback listener: %s", got)
	}
	t.Setenv("BIND_HOST", "")
	if got := Load().Addr; got != ":8081" {
		t.Fatalf("container listener: %s", got)
	}
}
