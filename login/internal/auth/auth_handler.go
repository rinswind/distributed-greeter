package auth

import (
	"encoding/base64"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
)

const (
	// ContextKey under which the JWT claims will be stored in the request Context
	ContextKey = "auth.Authorization"
)

var (
	authzSplit = regexp.MustCompile("Bearer (.+)")
)

// Handler is a middleware function to authenticate an HTTP endpoint
func Handler(c *gin.Context) {
	authzHeader := c.Request.Header.Get("Authorization")

	split := authzSplit.FindStringSubmatch(authzHeader)
	if len(split) != 2 {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	token, err := base64.StdEncoding.DecodeString(split[1])
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	claims, err := DecodeToken(string(token))
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	c.Set(ContextKey, claims)
}
