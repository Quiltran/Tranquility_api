package services_test

import (
	"testing"
	"time"
	"tranquility/config"
	"tranquility/models"
	"tranquility/services"
)

func TestGenerateToken(t *testing.T) {
	user := models.AuthUser{
		ID:       1,
		Username: "Steven",
	}

	config := config.JWTConfig{
		Key: "testing",
	}
	jwtHandler := services.NewJWTHandler(&config)

	token, err := jwtHandler.GenerateToken(&user)
	if err != nil {
		t.Fatalf("generating a token returned an error: %v", err)
	}

	if token == "" {
		t.Fatal("an empty string was returned while generating token")
	}
	t.Logf("generated token: %s", token)
}

func TestValidateToken(t *testing.T) {
	config := config.JWTConfig{
		Lifetime: time.Duration(2 * time.Minute),
		Key:      "testing",
	}
	jwtHandler := services.NewJWTHandler(&config)

	user := models.AuthUser{
		ID:       1,
		Username: "Steven",
	}
	claims := services.Claims{
		Username: "Steven",
		ID:       1,
	}
	token, err := jwtHandler.GenerateToken(&user)
	if err != nil {
		t.Fatalf("generating token returned an error: %v", err)
	}

	parsedToken, err := jwtHandler.VerifyToken(token)
	if err != nil {
		t.Fatalf("verifying token has returned an error: %v", err)
	}

	if claims.Username != parsedToken.Username {
		t.Fatalf("parsed token username does not match original")
	}
	if claims.ID != parsedToken.ID {
		t.Fatalf("parsed token id does not match original id")
	}
}
