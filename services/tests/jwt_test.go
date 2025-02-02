package services_test

import (
	"os"
	"testing"
	"tranquility/models"
	"tranquility/services"
)

func TestGenerateToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "testing")
	services.LoadJWTSettings()

	user := models.AuthUser{
		ID:       1,
		Username: "Steven",
	}

	token, err := services.GenerateToken(&user)
	if err != nil {
		t.Fatalf("generating a token returned an error: %v", err)
	}

	if token == "" {
		t.Fatal("an empty string was returned while generating token")
	}
	t.Logf("generated token: %s", token)
}

func TestValidateToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "testing")
	services.LoadJWTSettings()

	user := models.AuthUser{
		ID:       1,
		Username: "Steven",
	}
	claims := services.Claims{
		Username: "Steven",
		ID:       1,
	}
	token, err := services.GenerateToken(&user)
	if err != nil {
		t.Fatalf("generating token returned an error: %v", err)
	}

	parsedToken, err := services.VerifyToken(token)
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
