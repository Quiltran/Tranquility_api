package config

import (
	"os"
	"reflect"
	"slices"
	"testing"
	"time"
)

func TestNewConfig(t *testing.T) {
	// Backup environment variables
	backupEnv := make(map[string]string)
	for _, env := range []string{"CONNECTION_STRING", "UPLOAD_PATH", "ALLOWED_ORIGINS", "JWT_LIFETIME", "JWT_ISSUER", "JWT_AUDIENCE", "JWT_SECRET"} {
		backupEnv[env] = os.Getenv(env)
	}
	defer func() {
		// Restore environment variables
		for env, value := range backupEnv {
			os.Setenv(env, value)
		}
	}()

	t.Run("success", func(t *testing.T) {
		os.Setenv("CONNECTION_STRING", "test_connection")
		os.Setenv("UPLOAD_PATH", "test_upload")
		os.Setenv("ALLOWED_ORIGINS", "http://localhost:3000,https://example.com")
		os.Setenv("JWT_LIFETIME", "10")
		os.Setenv("JWT_ISSUER", "test_issuer")
		os.Setenv("JWT_AUDIENCE", "test_audience1,test_audience2")
		os.Setenv("JWT_SECRET", "test_secret")

		cfg, err := NewConfig()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expectedOrigins := []string{"http://localhost:3000", "https://example.com"}
		expectedAudience := []string{"test_audience1", "test_audience2"}
		slices.Sort(expectedAudience)

		if cfg.ConnectionString != "test_connection" {
			t.Errorf("ConnectionString mismatch: got %v, want %v", cfg.ConnectionString, "test_connection")
		}
		if cfg.UploadPath != "test_upload" {
			t.Errorf("UploadPath mismatch: got %v, want %v", cfg.UploadPath, "test_upload")
		}
		if !reflect.DeepEqual(cfg.AllowedOrigins, expectedOrigins) {
			t.Errorf("AllowedOrigins mismatch: got %v, want %v", cfg.AllowedOrigins, expectedOrigins)
		}
		if cfg.JWTConfig.Lifetime != 10*time.Minute {
			t.Errorf("JWT Lifetime mismatch: got %v, want %v", cfg.JWTConfig.Lifetime, 10*time.Minute)
		}
		if cfg.JWTConfig.Issuer != "test_issuer" {
			t.Errorf("JWT Issuer mismatch: got %v, want %v", cfg.JWTConfig.Issuer, "test_issuer")
		}
		if !reflect.DeepEqual(cfg.JWTConfig.Audience, expectedAudience) {
			t.Errorf("JWT Audience mismatch: got %v, want %v", cfg.JWTConfig.Audience, expectedAudience)
		}
		if cfg.JWTConfig.Key != "test_secret" {
			t.Errorf("JWT Secret mismatch: got %v, want %v", cfg.JWTConfig.Key, "test_secret")
		}
	})

	t.Run("missing CONNECTION_STRING", func(t *testing.T) {
		os.Unsetenv("CONNECTION_STRING")
		_, err := NewConfig()
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		if err.Error() != "CONNECTION_STRING was not set" {
			t.Errorf("expected error message '%s', got '%s'", "CONNECTION_STRING was not set", err.Error())
		}
	})

	t.Run("missing UPLOAD_PATH", func(t *testing.T) {
		os.Setenv("CONNECTION_STRING", "test_connection")
		os.Unsetenv("UPLOAD_PATH")
		_, err := NewConfig()
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		if err.Error() != "UPLOAD_PATH was not set" {
			t.Errorf("expected error message '%s', got '%s'", "UPLOAD_PATH was not set", err.Error())
		}
	})

	t.Run("missing ALLOWED_ORIGINS", func(t *testing.T) {
		os.Setenv("CONNECTION_STRING", "test_connection")
		os.Setenv("UPLOAD_PATH", "test_upload")
		os.Unsetenv("ALLOWED_ORIGINS")
		_, err := NewConfig()
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		if err.Error() != "ALLOWED_ORIGINS was not set" {
			t.Errorf("expected error message '%s', got '%s'", "ALLOWED_ORIGINS was not set", err.Error())
		}
	})

	t.Run("invalid JWT_LIFETIME", func(t *testing.T) {
		os.Setenv("CONNECTION_STRING", "test_connection")
		os.Setenv("UPLOAD_PATH", "test_upload")
		os.Setenv("ALLOWED_ORIGINS", "http://localhost:3000")
		os.Setenv("JWT_LIFETIME", "invalid")
		os.Setenv("JWT_SECRET", "test_secret")

		defer func() {
			if r := recover(); r == nil {
				t.Errorf("The code did not panic")
			}
		}()
		NewConfig()
	})

	t.Run("missing JWT_SECRET", func(t *testing.T) {
		os.Setenv("CONNECTION_STRING", "test_connection")
		os.Setenv("UPLOAD_PATH", "test_upload")
		os.Setenv("ALLOWED_ORIGINS", "http://localhost:3000")
		os.Unsetenv("JWT_SECRET")

		defer func() {
			if r := recover(); r == nil {
				t.Errorf("The code did not panic")
			}
		}()
		NewConfig()
	})

	t.Run("Empty ALLOWED_ORIGINS string", func(t *testing.T) {
		os.Setenv("CONNECTION_STRING", "test_connection")
		os.Setenv("UPLOAD_PATH", "test_upload")
		os.Setenv("ALLOWED_ORIGINS", "")
		_, err := NewConfig()
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		if err.Error() != "ALLOWED_ORIGINS was not set" {
			t.Errorf("expected error message '%s', got '%s'", "ALLOWED_ORIGINS was not set", err.Error())
		}
	})
}
