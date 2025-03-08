package test

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"testing"
	"time"
	"tranquility/config"
	"tranquility/models"
	"tranquility/services"
)

var (
	jwtConfig config.JWTConfig
)

func init() {
	jwePem, err := os.ReadFile("../.vscode/private_key.pem")
	if err != nil {
		panic(fmt.Errorf("an error occurred while reading JWE private key: %v", err))
	}
	block, _ := pem.Decode(jwePem)
	if block == nil {
		panic(fmt.Errorf("an error occurred while decoding JWE private key"))
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		panic(fmt.Errorf("an error occurred while parsing JWE private key: %v", err))
	}
	rsaKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		panic(fmt.Errorf("an error occurred while converting JWE private key to RSA: %v", err))
	}

	jwtConfig = config.JWTConfig{
		JWEPrivateKey: rsaKey,
		Lifetime:      time.Duration(2 * time.Minute),
		Issuer:        "TestIssuer",
		Audience:      []string{"http://localhost", "https://example.com"},
		Key:           "secret_key",
	}

}

func TestGenerateToken(t *testing.T) {
	user := models.AuthUser{
		ID:         1,
		Username:   "Steven",
		UserHandle: []byte{1, 2, 3, 4, 5},
	}

	jwtHandler := services.NewJWTHandler(&jwtConfig)

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
	jwtHandler := services.NewJWTHandler(&jwtConfig)

	user := models.AuthUser{
		ID:         1,
		Username:   "Steven",
		UserHandle: []byte("THIS IS THE HANDLE"),
	}
	claims := models.Claims{
		Username:   "Steven",
		ID:         1,
		UserHandle: base64.StdEncoding.EncodeToString([]byte("THIS IS THE HANDLE")),
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
	if claims.UserHandle != parsedToken.UserHandle {
		t.Fatalf("parsed user handle does not match original user handle")
	}
}
