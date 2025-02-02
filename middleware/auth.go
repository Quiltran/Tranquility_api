package middleware

import (
	"context"
	"net/http"
	"tranquility/services"
)

type claims string

var (
	ClaimContext = claims("claims")
)

func ValidateJWT(next http.Handler, logger *services.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("authentication")
		if authHeader == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		token := authHeader[len("Bearer "):]

		claims, err := services.VerifyToken(token)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ClaimContext, claims)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
