package middleware

import (
	"context"
	"fmt"
	"net/http"
	"tranquility/services"
)

type claimsKey struct{}

var ClaimsContextKey claimsKey

// ValidateJWT middleware is used to completely verify the JWT.
//
// It will return a 401 if any of the following are found:
//  1. The JWT is expired
//  2. The audience provided does not match
//  3. The issuer provided does not match
//  4. The signature is invalid
func ValidateJWT(next http.Handler, logger services.Logger, jwtHandler *services.JWTHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			logger.WARNING("auth head was missing on required route")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if len(authHeader) < len("Bearer ") {
			logger.WARNING("auth header is not the correct size")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		token := authHeader[len("Bearer "):]

		claims, err := jwtHandler.VerifyToken(token)
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

func ParseJWT(next http.Handler, logger services.Logger, jwtHandler *services.JWTHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			logger.WARNING("auth head was missing on required route")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		token := authHeader[len("Bearer "):]

		claims, err := jwtHandler.ParseToken(token)
		if err != nil {
			logger.ERROR(fmt.Sprintf("an error occurred while parsing auth token: %v", err))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
