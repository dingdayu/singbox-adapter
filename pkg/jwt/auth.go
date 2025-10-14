// Package jwt provides JWT issue and parse utilities.
package jwt

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.opentelemetry.io/otel/attribute"
)

// User defines custom JWT claims.
type User struct {
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	RealName string   `json:"real_name,omitempty"`
	Email    string   `json:"email,omitempty"`

	jwt.RegisteredClaims
}

// GenerateJWT generates a JWT token string.
func GenerateJWT(ctx context.Context, username, realName, mail string, roles []string, secretKey string) (string, error) {
	_, span := tracer.Start(ctx, "GenerateJWT")
	defer span.End()

	span.SetAttributes(
		attribute.String("username", username),
		attribute.String("email", mail),
		attribute.StringSlice("roles", roles))

	claims := User{
		Email:    mail,
		Username: username,
		RealName: realName,
		Roles:    roles, // Add roles field
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "SingBox-Adapter", // issuer
			Subject:   username,
			Audience:  jwt.ClaimStrings{"vben"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)), // expires in 7 days
			IssuedAt:  jwt.NewNumericDate(time.Now()),                         // issued at
			NotBefore: jwt.NewNumericDate(time.Now()),                         // not before
		},
	}
	// sign with HS256
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t.Header["kid"] = "v1"
	s, err := t.SignedString([]byte(secretKey))

	return s, err
}

// ParseJwt parses the JWT token and returns claims.
func ParseJwt(tokenstring, secretKey string) (*User, error) {
	t, err := jwt.ParseWithClaims(tokenstring, &User{}, func(_ *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if claims, ok := t.Claims.(*User); ok && t.Valid {
		return claims, nil
	}
	return nil, err
}
