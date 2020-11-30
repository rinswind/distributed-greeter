package auth

import (
	"context"
	"encoding/base64"
	"net/http"
	"regexp"
)

var (
	authzSplit = regexp.MustCompile("Bearer (.+)")
)

// Ctx is a typesafe key under which the user name is attached to the request context
type Ctx struct {
}

// Handler is a middleware function to authenticate an HTTP endpoint
func Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authzHeader := r.Header.Get("Authorization")

		split := authzSplit.FindStringSubmatch(authzHeader)
		if len(split) != 2 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		token, err := base64.StdEncoding.DecodeString(split[1])
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		user, err := GetAuth(string(token))
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), Ctx{}, user)
		req := r.Clone(ctx)

		next.ServeHTTP(w, req)
	})
}
