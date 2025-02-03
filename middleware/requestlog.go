package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"
	"tranquility/services"

	"github.com/google/uuid"
)

type requestIDKey struct{}

var RequestID requestIDKey

func RequestLog(next http.Handler, logger *services.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := uuid.New().String()
		logger.TRACE(fmt.Sprintf("Request received %s: %s %s from %s", requestID, r.Method, r.URL.Path, r.RemoteAddr))

		ctx := context.WithValue(r.Context(), RequestID, requestID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)

		duration := time.Since(start)
		logger.TRACE(fmt.Sprintf("Request %s completed in %s", requestID, duration.String()))
	})
}
