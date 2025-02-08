package controllers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"tranquility/services"
)

func TestHandleError(t *testing.T) {
	// Create a mock logger
	mockLogger := &MockLogger{}

	// Create test cases
	tests := []struct {
		name     string
		w        *httptest.ResponseRecorder
		err      error
		claims   *services.Claims
		code     int
		logLevel string
		wantLog  string
		wantCode int
	}{
		{
			name:     "error with valid claims",
			w:        httptest.NewRecorder(),
			err:      errors.New("test error"),
			claims:   &services.Claims{Username: "testUser"},
			code:     http.StatusBadRequest,
			logLevel: "ERROR",
			wantLog:  "testUser encountered error: test error",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "warning with nil claims",
			w:        httptest.NewRecorder(),
			err:      errors.New("test warning"),
			claims:   nil,
			code:     http.StatusBadRequest,
			logLevel: "WARNING",
			wantLog:  "Anonymous encountered warning: test warning",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "different status code",
			w:        httptest.NewRecorder(),
			err:      errors.New("internal error"),
			claims:   &services.Claims{Username: "testUser"},
			code:     http.StatusInternalServerError,
			logLevel: "ERROR",
			wantLog:  "testUser encountered error: internal error",
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handleError(tt.w, mockLogger, tt.err, tt.claims, tt.code, tt.logLevel)

			// Check status code
			if tt.w.Code != tt.wantCode {
				t.Errorf("handleError() status code = %v, want %v", tt.w.Code, tt.wantCode)
			}

			// Check response body contains status text
			if !strings.Contains(tt.w.Body.String(), http.StatusText(tt.wantCode)) {
				t.Errorf("handleError() body = %v, want to contain %v", tt.w.Body.String(), http.StatusText(tt.wantCode))
			}

			// Check if logger was called with correct message
			if mockLogger.LastMessage != tt.wantLog {
				t.Errorf("handleError() log message = %v, want %v", mockLogger.LastMessage, tt.wantLog)
			}
		})
	}
}

// Mock logger implementation
type MockLogger struct {
	LastMessage string
}

func (m *MockLogger) ERROR(msg string) {
	m.LastMessage = msg
}

func (m *MockLogger) WARNING(msg string) {
	m.LastMessage = msg
}

func (m *MockLogger) INFO(msg string) {
	m.LastMessage = msg
}

func (m *MockLogger) TRACE(msg string) {
	m.LastMessage = msg
}
