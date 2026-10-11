package identity

import "testing"

func TestEffectiveRoles(t *testing.T) {
	user := User{ID: "u1", Email: "writer@example.com", Roles: []Role{Author, Admin}}
	allowed := EffectiveRoles(user, []string{" WRITER@example.com "})
	if !allowed.HasRole(Reader) || !allowed.HasRole(Author) || !allowed.HasRole(Admin) {
		t.Fatalf("allowed roles: %v", allowed.Roles)
	}
	revoked := EffectiveRoles(user, nil)
	if revoked.HasRole(Admin) || !revoked.HasRole(Author) {
		t.Fatalf("revoked roles: %v", revoked.Roles)
	}
	legacy := EffectiveRoles(User{ID: "legacy"}, nil)
	if len(legacy.Roles) != 1 || !legacy.HasRole(Reader) {
		t.Fatalf("legacy roles: %v", legacy.Roles)
	}
}
