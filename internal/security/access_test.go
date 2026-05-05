package security

import (
	"testing"
	"time"
)

func TestAuthenticateToken(t *testing.T) {
	principal, err := AuthenticateToken("token", map[string]Principal{
		"token": {ID: "user", Roles: []string{"admin"}},
	})
	if err != nil {
		t.Fatalf("AuthenticateToken: %v", err)
	}
	if principal.ID != "user" {
		t.Fatalf("principal ID = %q, want user", principal.ID)
	}
	if _, err := AuthenticateToken("", nil); err == nil {
		t.Fatal("empty token authenticated")
	}
	if _, err := AuthenticateToken("bad", nil); err == nil {
		t.Fatal("bad token authenticated")
	}
}

func TestAuthorizer(t *testing.T) {
	authz := NewAuthorizer(map[string][]string{
		"admin": {"module:push", "config:set"},
		"user":  {"prompt:run"},
	})
	if err := authz.Authorize(Principal{ID: "u", Roles: []string{"admin"}}, "config:set"); err != nil {
		t.Fatalf("Authorize admin: %v", err)
	}
	if err := authz.Authorize(Principal{ID: "u", Roles: []string{"user"}}, "config:set"); err == nil {
		t.Fatal("unauthorized user accepted")
	}
	if err := authz.Authorize(Principal{}, "config:set"); err == nil {
		t.Fatal("unauthenticated principal accepted")
	}
}

func TestRateLimiter(t *testing.T) {
	limiter := NewRateLimiter(2, 5*time.Millisecond)
	if !limiter.Allow() || !limiter.Allow() {
		t.Fatal("initial tokens rejected")
	}
	if limiter.Allow() {
		t.Fatal("third request allowed before refill")
	}
	time.Sleep(6 * time.Millisecond)
	if !limiter.Allow() {
		t.Fatal("request rejected after refill")
	}
}
