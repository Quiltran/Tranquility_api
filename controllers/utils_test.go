package controllers

import (
	"bytes"
	"compress/gzip"
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

type TestStruct struct {
	StringField  string            `json:"string_field"`
	IntField     int               `json:"int_field"`
	BoolField    bool              `json:"bool_field"`
	SliceField   []string          `json:"slice_field"`
	MapField     map[string]int    `json:"map_field"`
	NestedStruct *NestedTestStruct `json:"nested_struct,omitempty"`
}

type NestedTestStruct struct {
	NestedString string `json:"nested_string"`
}

func TestWriteJsonBody(t *testing.T) {
	testCases := []struct {
		name    string
		body    interface{}
		wantErr bool
		want    interface{}
	}{
		{
			name: "valid struct",
			body: TestStruct{
				StringField:  "test",
				IntField:     123,
				BoolField:    true,
				SliceField:   []string{"a", "b"},
				MapField:     map[string]int{"x": 1, "y": 2},
				NestedStruct: &NestedTestStruct{NestedString: "nested"},
			},
			wantErr: false,
		},
		{
			name: "valid struct pointer",
			body: &TestStruct{
				StringField:  "test",
				IntField:     123,
				BoolField:    true,
				SliceField:   []string{"a", "b"},
				MapField:     map[string]int{"x": 1, "y": 2},
				NestedStruct: &NestedTestStruct{NestedString: "nested"},
			},
			wantErr: false,
		},
		{
			name:    "nil struct pointer",
			body:    (*TestStruct)(nil),
			wantErr: true,
		},
		{
			name:    "valid slice",
			body:    []int{1, 2, 3},
			wantErr: false,
		},
		{
			name:    "valid array",
			body:    [3]int{1, 2, 3},
			wantErr: false,
		},
		{
			name:    "invalid type int",
			body:    123,
			wantErr: true,
		},
		{
			name:    "invalid type string",
			body:    "test",
			wantErr: true,
		},
		{
			name:    "empty struct",
			body:    TestStruct{},
			wantErr: false,
		},
		{
			name:    "empty slice",
			body:    []string{},
			wantErr: false,
		},
		{
			name:    "empty map",
			body:    map[string]int{},
			wantErr: true,
		},
		{
			name: "struct with nil nested struct",
			body: TestStruct{
				StringField: "test",
				IntField:    123,
				BoolField:   true,
				SliceField:  []string{"a", "b"},
				MapField:    map[string]int{"x": 1, "y": 2},
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			err := writeJsonBody(recorder, tc.body)

			if (err != nil) != tc.wantErr {
				t.Errorf("writeJsonBody() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if !tc.wantErr {
				if recorder.Header().Get("content-type") != "application/json" {
					t.Errorf("content-type header not set correctly")
				}
				if recorder.Header().Get("Content-Encoding") != "gzip" {
					t.Errorf("Content-Encoding header not set correctly")
				}

				gzipReader, err := gzip.NewReader(recorder.Body)
				if err != nil {
					t.Fatalf("failed to create gzip reader: %v", err)
				}
				defer gzipReader.Close()

				var buf bytes.Buffer
				_, err = buf.ReadFrom(gzipReader)
				if err != nil {
					t.Fatalf("failed to read from gzip reader: %v", err)
				}

				var decoded interface{}
				err = json.Unmarshal(buf.Bytes(), &decoded)
				if err != nil {
					t.Fatalf("failed to unmarshal json: %v", err)
				}

				if tc.want != nil {
					if !reflect.DeepEqual(decoded, tc.want) {
						t.Errorf("decoded json does not match expected: got %v, want %v", decoded, tc.want)
					}
				} else {
					//If want is nil, just check that unmarshaling was successful.
				}

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
