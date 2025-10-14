// Package controller defines HTTP controllers.
package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Hello is the home/health-check endpoint.
func Hello(c *gin.Context) {
	c.String(http.StatusOK, "hello, word.")
}

// Ping responds with pong for health checking.
func Ping(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}
