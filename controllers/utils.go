package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"tranquility/middleware"
	"tranquility/models"
	"tranquility/services"
)

func getClaims(r *http.Request) (*models.Claims, error) {
	claims, ok := r.Context().Value(middleware.ClaimsContextKey).(*models.Claims)
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

func getJsonBody[T any](r *http.Request) (*T, error) {
	var body T

	v := reflect.ValueOf(body)
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("tried getting body from request but it was not a struct")
	}

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		return nil, err
	}

	return &body, nil
}

func writeJsonBody[T any](w http.ResponseWriter, body T) error {
	v := reflect.ValueOf(body)

	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return fmt.Errorf("nil pointer provided to response body")
		}

		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Array:
	case reflect.Struct:
	case reflect.Slice:
		break
	default:
		return fmt.Errorf("tried writing body to request but it was not a struct or array: %v", v.Kind())
	}

	w.Header().Add("content-type", "application/json")
	return json.NewEncoder(w).Encode(body)
}

func handleError(w http.ResponseWriter, r *http.Request, logger services.Logger, err error, claims *models.Claims, code int, logLevel string, message ...string) {
	var requestId string
	requestId, requestErr := getRequestID(r)
	if requestErr != nil {
		requestId = requestErr.Error()
	}
	if claims == nil {
		claims = &models.Claims{
			Username: "Anonymous",
		}
	}
	switch strings.ToLower(logLevel) {
	case "error":
		logger.ERROR(fmt.Sprintf("requestId: %s: %s encountered error: %v", requestId, claims.Username, err))
	case "warning":
		logger.ERROR(fmt.Sprintf("requestId: %s: %s encountered warning: %v", requestId, claims.Username, err))
	}

	responseText := http.StatusText(code)
	if len(message) > 0 {
		responseText = message[0]
	}
	http.Error(w, responseText, code)
}
