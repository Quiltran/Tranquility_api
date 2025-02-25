package middleware

import (
	"context"
	"fmt"
	"net/http"
	"tranquility/services"
)

type claimsKey struct{}

var ClaimsContextKey claimsKey

func ValidateJWT(next http.Handler, logger services.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			logger.WARNING("auth head was missing on required route")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		token := authHeader[len("Bearer "):]

		claims, err := services.VerifyToken(token)
		if err != nil {
			logger.ERROR(fmt.Sprintf("an error occurred while verifying auth token: %v", err))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func ParseJWT(next http.Handler, logger services.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			logger.WARNING("auth head was missing on required route")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		token := authHeader[len("Bearer "):]

		claims, err := services.ParseToken(token)
		if err != nil {
			logger.ERROR(fmt.Sprintf("an error occurred while verifying auth token: %v", err))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
