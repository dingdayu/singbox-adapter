// Package middleware defines gin middlewares.
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/dingdayu/go-project-template/model/entity"
	"github.com/dingdayu/go-project-template/model/entity/contextkey"
	"github.com/dingdayu/go-project-template/pkg/jwt"
	"github.com/dingdayu/go-project-template/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// Authorization returns a JWT-based authentication middleware.
func Authorization() gin.HandlerFunc {
	secret := viper.GetString("jwt.secret")
	if secret == "" {
		panic("JWT secret is not configured")
	}

	return func(c *gin.Context) {
		authorization := c.Request.Header.Get("Authorization")
		if authorization == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, entity.ErrAuthForbidden)
			return
		}
		token := strings.Split(authorization, " ")
		if token[0] != "Bearer" || len(token) != 2 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, entity.ErrAuthForbidden)
			return
		}
		// validate JWT token
		user, err := jwt.ParseJwt(token[1], secret)
		if err != nil {
			logger.Logger().WarnContext(c.Request.Context(), "JWT parsing failed",
				"error", err,
				"ip", c.ClientIP(),
				"user_agent", c.Request.UserAgent())
			c.AbortWithStatusJSON(http.StatusUnauthorized, entity.ErrAuthTokenInvalid)
			return
		}

		// build enhanced request context
		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, contextkey.Email, user.Email)
		ctx = context.WithValue(ctx, contextkey.RealName, user.RealName)
		ctx = context.WithValue(ctx, contextkey.Role, user.Roles)
		ctx = context.WithValue(ctx, contextkey.UserName, user.Username)
		ctx = context.WithValue(ctx, contextkey.IP, c.ClientIP())

		// attach updated context to request
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
