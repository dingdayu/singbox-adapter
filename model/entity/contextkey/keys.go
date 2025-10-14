// Package contextkey defines keys used in context.
package contextkey

// UserContext is a key type used to pass user info via context.
type UserContext string

var (
	// Email is the user email
	Email UserContext = "email"
	// RealName is the user's real name
	RealName UserContext = "real_name"
	// UserName is the username (login)
	UserName UserContext = "user_name"
	// Role is the user's roles
	Role UserContext = "role"
	// IP is the client IP address
	IP UserContext = "ip"
)
