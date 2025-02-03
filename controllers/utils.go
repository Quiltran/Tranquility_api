package controllers

import (
	"fmt"
	"net/http"
	"tranquility/middleware"
	"tranquility/services"
)

func getClaims(r *http.Request) (*services.Claims, error) {
	claims, ok := r.Context().Value(middleware.ClaimsContextKey).(*services.Claims)
	if !ok {
		return nil, fmt.Errorf("a request was made without valid claims to refresh auth tokens")
	}

	return claims, nil
}

func getRequestID(r *http.Request) (string, error) {
	requestId, ok := r.Context().Value(middleware.RequestID).(string)
	if !ok {
		return "", fmt.Errorf("a request was made without a request id")
	}

	return requestId, nil
}
