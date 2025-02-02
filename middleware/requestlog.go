package middleware

import (
	"fmt"
	"net/http"
	"time"
	"tranquility/services"

	"github.com/google/uuid"
)

func RequestLog(next http.Handler, logger *services.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := uuid.New().String()
		logger.INFO(fmt.Sprintf("Request received %s: %s %s from %s", requestID, r.Method, r.URL.Path, r.RemoteAddr))

		next.ServeHTTP(w, r)

		duration := time.Since(start)
		logger.INFO(fmt.Sprintf("Request %s completed in: %s", requestID, duration.String()))
	})
}
