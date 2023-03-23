package auth

import (
	"github.com/golang-jwt/jwt/v4"
)

// Claims resolves permissions for a requester.
type Claims interface {
	// Subject returns the token subject if present (usually a user or service ID)
	Subject() string
	// Can resolves system-level permissions.
	Can(p Permission) bool
	// CanInstance resolves instance-level permissions.
	CanInstance(instanceID string, p Permission) bool
	// Email returns the email of the user
	GetEmail() string
	// UserGroup returns the group user belongs to
	GetUserGroups() []string
}

// jwtClaims implements Claims and resolve permissions based on a JWT payload.
type jwtClaims struct {
	jwt.RegisteredClaims
	System    []Permission            `json:"sys,omitempty"`
	Instances map[string][]Permission `json:"ins,omitempty"`
	Email     string                  `json:"email"`
	Groups    []string                `json:"group_names"`
}

func (c *jwtClaims) Subject() string {
	return c.RegisteredClaims.Subject
}

func (c *jwtClaims) Can(p Permission) bool {
	for _, p2 := range c.System {
		if p2 == p {
			return true
		}
	}
	return false
}

func (c *jwtClaims) CanInstance(instanceID string, p Permission) bool {
	for _, p2 := range c.Instances[instanceID] {
		if p2 == p {
			return true
		}
	}
	return c.Can(p)
}

func (c *jwtClaims) GetEmail() string {
	return c.Email
}

func (c *jwtClaims) GetUserGroups() []string {
	return c.Groups
}

// openClaims implements Claims and allows all actions.
// It is used for servers with auth disabled.
type openClaims struct{}

func (c openClaims) Subject() string {
	return ""
}

func (c openClaims) Can(p Permission) bool {
	return true
}

func (c openClaims) CanInstance(instanceID string, p Permission) bool {
	return true
}

func (c openClaims) GetEmail() string {
	return ""
}

func (c openClaims) GetUserGroups() []string {
	return nil
}

// anonClaims imeplements Claims with no permissions.
// It is used for unauthorized requests when auth is enabled.
type anonClaims struct{}

func (c anonClaims) Subject() string {
	return ""
}

func (c anonClaims) Can(p Permission) bool {
	return false
}

func (c anonClaims) CanInstance(instanceID string, p Permission) bool {
	return false
}

func (c anonClaims) GetEmail() string {
	return ""
}

func (c anonClaims) GetUserGroup() []string {
	return nil
}
