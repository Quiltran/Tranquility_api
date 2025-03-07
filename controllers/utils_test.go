package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"tranquility/middleware"
	"tranquility/models"
)

func TestGetJsonBody(t *testing.T) {
	type Address struct {
		Street string `json:"street"`
		City   string `json:"city"`
	}
	type TestStruct struct {
		Name    string  `json:"name"`
		Age     int     `json:"age"`
		Email   string  `json:"email,omitempty"` // Example of optional field
		Address Address `json:"address"`         //Example of nested struct
	}

	tests := []struct {
		name        string
		requestBody string
		want        TestStruct
		wantErr     bool
		errMsg      string //Specific error message check
	}{
		{
			name:        "Valid JSON",
			requestBody: `{"name": "John Doe", "age": 30, "address": {"street":"123 Main St", "city":"Anytown"}}`,
			want:        TestStruct{Name: "John Doe", Age: 30, Address: Address{Street: "123 Main St", City: "Anytown"}},
			wantErr:     false,
		},
		{
			name:        "Invalid JSON",
			requestBody: `{"name": "John Doe", "age": "30"}`, // Age should be int
			want:        TestStruct{},
			wantErr:     true,
			errMsg:      "json: cannot unmarshal string into Go struct field TestStruct.age of type int",
			// errMsg:      "invalid character '}' after value", //Or a more general error check
		},
		{
			name:        "Empty JSON",
			requestBody: `{}`,
			want:        TestStruct{}, //Empty struct is valid
			wantErr:     false,
		},
		{
			name:        "Missing field",
			requestBody: `{"name": "John Doe"}`,       //Missing age
			want:        TestStruct{Name: "John Doe"}, // Zero value for int
			wantErr:     false,
		},
		{
			name:        "Optional Field",
			requestBody: `{"name": "John Doe", "age": 30, "email": "john.doe@example.com"}`,
			want:        TestStruct{Name: "John Doe", Age: 30, Email: "john.doe@example.com", Address: Address{}}, //Zero value for Address
			wantErr:     false,
		},
		{
			name:        "Not a struct",
			requestBody: `123`, //Not a struct
			want:        TestStruct{},
			wantErr:     true,
			errMsg:      "json: cannot unmarshal number into Go value of type controllers.TestStruct",
		},
		{
			name:        "Nested Struct",
			requestBody: `{"name": "John Doe", "age": 30, "address": {"street":"123 Main St", "city":"Anytown"}}`,
			want:        TestStruct{Name: "John Doe", Age: 30, Address: Address{Street: "123 Main St", City: "Anytown"}},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := bytes.NewReader([]byte(tt.requestBody))
			r, _ := http.NewRequest("POST", "/", reqBody) // Method doesn't matter for this test

			got, err := getJsonBody[TestStruct](r)

			if (err != nil) != tt.wantErr {
				t.Errorf("getJsonBody() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("getJsonBody() error message = %v, want message %v", err.Error(), tt.errMsg)
				}
				return //If we expect an error, no need to check the value
			}

			if !reflect.DeepEqual(*got, tt.want) { // Use DeepEqual for struct comparison
				t.Errorf("getJsonBody() = %v, want %v", *got, tt.want)
			}
		})
	}
}

func TestWriteJsonBody(t *testing.T) {
	tests := []struct {
		name                string
		body                interface{}
		expectedErr         string
		expectedBody        string
		expectedContentType string
	}{
		{
			name: "Valid struct",
			body: struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{Name: "Alice", Age: 30},
			expectedBody:        `{"name":"Alice","age":30}`,
			expectedContentType: "application/json",
		},
		{
			name:                "Empty struct",
			body:                struct{}{},
			expectedBody:        `{}`,
			expectedContentType: "application/json",
		},
		{
			name: "Nested struct",
			body: struct {
				Address struct {
					City    string `json:"city"`
					ZipCode string `json:"zip"`
				} `json:"address"`
			}{Address: struct {
				City    string `json:"city"`
				ZipCode string `json:"zip"`
			}{City: "New York", ZipCode: "10001"}},
			expectedBody:        `{"address":{"city":"New York","zip":"10001"}}`,
			expectedContentType: "application/json",
		},
		{
			name:        "Not a struct",
			body:        "not a struct",
			expectedErr: "tried writing body to request but it was not a struct or array: string",
		},
		{
			name:                "Slice",
			body:                []int{1, 2, 3},
			expectedBody:        "[1, 2, 3]",
			expectedContentType: "application/json",
		},
		{
			name:        "Map",
			body:        map[string]int{"a": 1},
			expectedErr: "tried writing body to request but it was not a struct or array: map",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()

			err := writeJsonBody(recorder, tt.body)

			if tt.expectedErr != "" {
				if err == nil {
					t.Errorf("expected error: %q, but got nil", tt.expectedErr)
				} else if err.Error() != tt.expectedErr {
					t.Errorf("expected error: %q, but got: %q", tt.expectedErr, err.Error())
				}
				return // If expecting error, no further checks needed
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if recorder.Header().Get("content-type") != tt.expectedContentType {
				t.Errorf("expected content-type: %q, but got: %q", tt.expectedContentType, recorder.Header().Get("content-type"))
			}

			var buf bytes.Buffer
			buf.ReadFrom(recorder.Body)
			actualBody := buf.String()

			// Use a more robust comparison for JSON to handle variations in whitespace/ordering
			var expectedJSON interface{}
			json.Unmarshal([]byte(tt.expectedBody), &expectedJSON)
			var actualJSON interface{}
			json.Unmarshal([]byte(actualBody), &actualJSON)

			if !reflect.DeepEqual(actualJSON, expectedJSON) {
				t.Errorf("expected body: %q, but got: %q", tt.expectedBody, actualBody)
			}

		})
	}
}

func TestHandleError(t *testing.T) {
	// Create a mock logger
	mockLogger := &MockLogger{}

	var requestWId http.Request = *httptest.NewRequest("GET", "/", nil).WithContext(context.WithValue(context.Background(), middleware.RequestID, "123"))
	var requestWOId http.Request = *httptest.NewRequest("GET", "/", nil)

	// Create test cases
	tests := []struct {
		name     string
		w        *httptest.ResponseRecorder
		r        *http.Request
		err      error
		claims   *models.Claims
		code     int
		logLevel string
		wantLog  string
		wantCode int
	}{
		{
			name:     "error with valid claims",
			w:        httptest.NewRecorder(),
			r:        &requestWId,
			err:      errors.New("test error"),
			claims:   &models.Claims{Username: "testUser"},
			code:     http.StatusBadRequest,
			logLevel: "ERROR",
			wantLog:  "requestId: 123: testUser encountered error: test error",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "warning with nil claims",
			w:        httptest.NewRecorder(),
			r:        &requestWId,
			err:      errors.New("test warning"),
			claims:   nil,
			code:     http.StatusBadRequest,
			logLevel: "WARNING",
			wantLog:  "requestId: 123: Anonymous encountered warning: test warning",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "different status code",
			w:        httptest.NewRecorder(),
			r:        &requestWId,
			err:      errors.New("internal error"),
			claims:   &models.Claims{Username: "testUser"},
			code:     http.StatusInternalServerError,
			logLevel: "ERROR",
			wantLog:  "requestId: 123: testUser encountered error: internal error",
			wantCode: http.StatusInternalServerError,
		},
		{
			name:     "different status code",
			w:        httptest.NewRecorder(),
			r:        &requestWOId,
			err:      errors.New("internal error"),
			claims:   &models.Claims{Username: "testUser"},
			code:     http.StatusInternalServerError,
			logLevel: "ERROR",
			wantLog:  "requestId: a request was made without a request id: testUser encountered error: internal error",
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handleError(tt.w, tt.r, mockLogger, tt.err, tt.claims, tt.code, tt.logLevel)

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
