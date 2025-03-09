package test

import (
	"os"
	"reflect"
	"slices"
	"testing"
	"time"
	"tranquility/config"
)

func TestNewConfig(t *testing.T) {
	// Backup environment variables
	backupEnv := make(map[string]string)
	for _, env := range []string{"CONNECTION_STRING", "UPLOAD_PATH", "ALLOWED_ORIGINS", "JWT_LIFETIME", "JWT_ISSUER", "JWT_AUDIENCE", "JWT_SECRET", "JWT_PRIVATE_KEY_PATH"} {
		backupEnv[env] = os.Getenv(env)
	}
	defer func() {
		// Restore environment variables
		for env, value := range backupEnv {
			os.Setenv(env, value)
		}
	}()
	// Base config values
	os.Setenv("CONNECTION_STRING", "test_connection")
	os.Setenv("UPLOAD_PATH", "test_upload")
	os.Setenv("ALLOWED_ORIGINS", "http://localhost:3000,https://example.com")
	os.Setenv("TURNSTILE_SECRET", "1x0000000000000000000000000000000AA")

	// JWT config values
	os.Setenv("JWT_PRIVATE_KEY_PATH", "../.vscode/private_key.pem")
	os.Setenv("JWT_LIFETIME", "10")
	os.Setenv("JWT_ISSUER", "test_issuer")
	os.Setenv("JWT_AUDIENCE", "test_audience1,test_audience2")
	os.Setenv("JWT_SECRET", "test_secret")

	// Push Notification config values
	os.Setenv("VAPID_PRIVATE", "private")
	os.Setenv("VAPID_PUBLIC", "public")
	os.Setenv("PUSH_SUB", "sub")

	// WebAuthn config values
	os.Setenv("RP_DISPLAY_NAME", "test.com")
	os.Setenv("RPID", "test")
	os.Setenv("RP_ORIGINS", "http://localhost,https://example.com")

	t.Run("success", func(t *testing.T) {
		cfg, err := config.NewConfig()
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

		if cfg.JWTConfig.JWEPrivateKey == nil {
			t.Errorf("JWEPrivateKey is null")
		}
		if cfg.JWTConfig.Lifetime != 2*time.Minute {
			t.Errorf("JWT Lifetime mismatch: got %v, want %v", cfg.JWTConfig.Lifetime, 2*time.Minute)
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

		if cfg.PushNotificationConfig.VapidPrivateKey != "private" {
			t.Errorf("VapidPrivateKey mismatch: got %v, want %v", cfg.VapidPrivateKey, "private")
		}
		if cfg.VapidPublicKey != "public" {
			t.Errorf("VapidPublicKey mismatch: got %v, want %v", cfg.VapidPrivateKey, "public")
		}
		if cfg.Sub != "sub" {
			t.Errorf("PushNotification Sub mismatch: got %v, want %v", cfg.Sub, "sub")
		}

		if cfg.RPDisplayName != "test.com" {
			t.Errorf("RPDisplayName mismatch: got %v, want %v", cfg.RPDisplayName, "test.com")
		}
		if cfg.RPID != "test" {
			t.Errorf("RPID mismatch: got %v, want %v", cfg.RPID, "test")
		}
		expectedRPOrigins := []string{"http://localhost", "https://example.com"}
		if !reflect.DeepEqual(cfg.RPOrigins, expectedRPOrigins) {
			t.Errorf("RPOrigins mismatch: got %v, want %v", cfg.RPOrigins, expectedRPOrigins)
		}
	})

	t.Run("missing CONNECTION_STRING", func(t *testing.T) {
		os.Unsetenv("CONNECTION_STRING")
		_, err := config.NewConfig()
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
		_, err := config.NewConfig()
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
		_, err := config.NewConfig()
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		if err.Error() != "ALLOWED_ORIGINS was not set" {
			t.Errorf("expected error message '%s', got '%s'", "ALLOWED_ORIGINS was not set", err.Error())
		}
	})

	t.Run("missing TURNSTILE_SECRET", func(t *testing.T) {
		os.Setenv("CONNECTION_STRING", "test_connection")
		os.Setenv("UPLOAD_PATH", "test_upload")
		os.Setenv("ALLOWED_ORIGINS", "http://localhost:3000")

		turnstileSecret := os.Getenv("TURNSTILE_SECRET")
		defer os.Setenv("TURNSTILE_SECRET", turnstileSecret)

		os.Unsetenv("TURNSTILE_SECRET")

		_, err := config.NewConfig()
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		if err.Error() != "TURNSTILE_SECRET was not set" {
			t.Errorf("expected error message '%s', got '%s'", "ALLOWED_ORIGINS was not set", err.Error())
		}
	})

	t.Run("missing JWT_PRIVATE_KEY_PATH", func(t *testing.T) {
		jwtPemPath := os.Getenv("JWT_PRIVATE_KEY_PATH")
		defer os.Setenv("JWT_PRIVATE_KEY_PATH", jwtPemPath)
		os.Unsetenv("JWT_PRIVATE_KEY_PATH")

		_, err := config.NewConfig()
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		if err.Error() != "an error occurred while reading JWE private key: open : The system cannot find the file specified." {
			t.Errorf(
				"expected error message '%s', got '%s'",
				"an error occurred while reading JWE private key: open : The system cannot find the file specified.",
				err.Error(),
			)
		}
	})

	t.Run("missing JWT_SECRET", func(t *testing.T) {
		os.Setenv("CONNECTION_STRING", "test_connection")
		os.Setenv("UPLOAD_PATH", "test_upload")
		os.Setenv("ALLOWED_ORIGINS", "http://localhost:3000")

		jwtSecret := os.Getenv("JWT_SECRET")
		defer os.Setenv("JWT_SECRET", jwtSecret)
		os.Unsetenv("JWT_SECRET")

		_, err := config.NewConfig()
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		if err.Error() != "JWT_SECRET was not set" {
			t.Errorf(
				"expected error message '%s', got '%s'",
				"JWT_SECRET was not set",
				err.Error(),
			)
		}
	})

	t.Run("Empty ALLOWED_ORIGINS string", func(t *testing.T) {
		os.Setenv("CONNECTION_STRING", "test_connection")
		os.Setenv("UPLOAD_PATH", "test_upload")

		allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
		defer os.Setenv("ALLOWED_ORIGINS", allowedOrigins)
		os.Unsetenv("ALLOWED_ORIGINS")

		_, err := config.NewConfig()
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		if err.Error() != "ALLOWED_ORIGINS was not set" {
			t.Errorf("expected error message '%s', got '%s'", "ALLOWED_ORIGINS was not set", err.Error())
		}
	})

	t.Run("Empty ISSUER string", func(t *testing.T) {
		issuer := os.Getenv("JWT_ISSUER")
		defer os.Setenv("JWT_ISSUER", issuer)
		os.Unsetenv("JWT_ISSUER")

		_, err := config.NewConfig()
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		if err.Error() != "JWT_ISSUER was not set" {
			t.Errorf(
				"expected error message '%s', got '%s'",
				"JWT_ISSUER was not set",
				err.Error(),
			)
		}
	})

	t.Run("Non-numeric Lifetime", func(t *testing.T) {
		lifetime := os.Getenv("JWT_LIFETIME")
		defer os.Setenv("JWT_LIFETIME", lifetime)
		os.Setenv("JWT_LIFETIME", "1r")

		_, err := config.NewConfig()
		if err == nil {
			t.Error("expected error, got nil")
		}
		if err.Error() != "an error occurred while loading jwt lifetime: strconv.ParseInt: parsing \"1r\": invalid syntax" {
			t.Errorf(
				"expected error message '%s', got '%s'",
				"an error occurred while loading jwt lifetime: strconv.ParseInt: parsing \"1r\": invalid syntax",
				err.Error(),
			)
		}
	})

	t.Run("Empty VAPID Private", func(t *testing.T) {
		vapidPrivate := os.Getenv("VAPID_PRIVATE")
		defer os.Setenv("VAPID_PRIVATE", vapidPrivate)
		os.Unsetenv("VAPID_PRIVATE")

		_, err := config.NewConfig()
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		if err.Error() != "VAPID_PRIVATE was not set" {
			t.Errorf(
				"expected error message '%s', got '%s'",
				"VAPID_PRIVATE was not set",
				err.Error(),
			)
		}
	})

	t.Run("Empty VAPID Public", func(t *testing.T) {
		vapidPrivate := os.Getenv("VAPID_PUBLIC")
		defer os.Setenv("VAPID_PUBLIC", vapidPrivate)
		os.Unsetenv("VAPID_PUBLIC")

		_, err := config.NewConfig()
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		if err.Error() != "VAPID_PUBLIC was not set" {
			t.Errorf(
				"expected error message '%s', got '%s'",
				"VAPID_PUBLIC was not set",
				err.Error(),
			)
		}
	})

	t.Run("Empty Push Sub", func(t *testing.T) {
		vapidPrivate := os.Getenv("PUSH_SUB")
		defer os.Setenv("PUSH_SUB", vapidPrivate)
		os.Unsetenv("PUSH_SUB")

		_, err := config.NewConfig()
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		if err.Error() != "PUSH_SUB was not set" {
			t.Errorf(
				"expected error message '%s', got '%s'",
				"PUSH_SUB was not set",
				err.Error(),
			)
		}
	})

	t.Run("Empty RP Display Name", func(t *testing.T) {
		vapidPrivate := os.Getenv("RP_DISPLAY_NAME")
		defer os.Setenv("RP_DISPLAY_NAME", vapidPrivate)
		os.Unsetenv("RP_DISPLAY_NAME")

		_, err := config.NewConfig()
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		if err.Error() != "RP_DISPLAY_NAME was not set" {
			t.Errorf(
				"expected error message '%s', got '%s'",
				"RP_DISPLAY_NAME was not set",
				err.Error(),
			)
		}
	})

	t.Run("Empty RP ID", func(t *testing.T) {
		vapidPrivate := os.Getenv("RPID")
		defer os.Setenv("RPID", vapidPrivate)
		os.Unsetenv("RPID")

		_, err := config.NewConfig()
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		if err.Error() != "RPID was not set" {
			t.Errorf(
				"expected error message '%s', got '%s'",
				"RPID was not set",
				err.Error(),
			)
		}
	})

	t.Run("Empty RP Origins", func(t *testing.T) {
		vapidPrivate := os.Getenv("RP_ORIGINS")
		defer os.Setenv("RP_ORIGINS", vapidPrivate)
		os.Unsetenv("RP_ORIGINS")

		_, err := config.NewConfig()
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		if err.Error() != "invalid RP_ORIGINS were dectected: []" {
			t.Errorf(
				"expected error message '%s', got '%s'",
				"invalid RP_ORIGINS were dectected: []",
				err.Error(),
			)
		}
	})
}
