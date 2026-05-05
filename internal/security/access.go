package security

import (
	"sync"
	"time"
)

// Principal is an authenticated caller.
type Principal struct {
	ID    string
	Roles []string
}

// AuthenticateToken validates token against a static token table.
func AuthenticateToken(token string, tokens map[string]Principal) (Principal, error) {
	if token == "" {
		return Principal{}, NewSecurityError("missing_token", "authentication token is required")
	}
	principal, ok := tokens[token]
	if !ok {
		return Principal{}, NewSecurityError("invalid_token", "authentication token is invalid")
	}
	return principal, nil
}

// Authorizer checks whether principals have permissions.
type Authorizer struct {
	rolePermissions map[string]map[string]bool
}

// NewAuthorizer creates an authorizer from role to permission names.
func NewAuthorizer(roles map[string][]string) *Authorizer {
	a := &Authorizer{rolePermissions: make(map[string]map[string]bool)}
	for role, permissions := range roles {
		a.rolePermissions[role] = make(map[string]bool)
		for _, permission := range permissions {
			a.rolePermissions[role][permission] = true
		}
	}
	return a
}

// Authorize returns nil when principal has permission.
func (a *Authorizer) Authorize(principal Principal, permission string) error {
	if principal.ID == "" {
		return NewSecurityError("unauthenticated", "principal is not authenticated")
	}
	for _, role := range principal.Roles {
		if a.rolePermissions[role][permission] {
			return nil
		}
	}
	return NewSecurityError("unauthorized", "principal is not authorized")
}

// RateLimiter is a small token-bucket limiter.
type RateLimiter struct {
	mu       sync.Mutex
	capacity int
	tokens   int
	interval time.Duration
	last     time.Time
}

// NewRateLimiter creates a token-bucket rate limiter.
func NewRateLimiter(capacity int, interval time.Duration) *RateLimiter {
	if capacity <= 0 {
		capacity = 1
	}
	if interval <= 0 {
		interval = time.Second
	}
	return &RateLimiter{
		capacity: capacity,
		tokens:   capacity,
		interval: interval,
		last:     time.Now(),
	}
}

// Allow reports whether one event may proceed.
func (l *RateLimiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	if now.Sub(l.last) >= l.interval {
		l.tokens = l.capacity
		l.last = now
	}
	if l.tokens == 0 {
		return false
	}
	l.tokens--
	return true
}
