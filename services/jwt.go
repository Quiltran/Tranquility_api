package services

import (
	"fmt"
	"slices"
	"time"
	"tranquility/config"
	"tranquility/models"

	"github.com/go-jose/go-jose/v4"
	"github.com/golang-jwt/jwt/v5"
)

type JWTHandler struct {
	*config.JWTConfig
}

func NewJWTHandler(config *config.JWTConfig) *JWTHandler {
	return &JWTHandler{config}
}

func (j *JWTHandler) encryptToken(token string) (string, error) {
	encryptor, err := jose.NewEncrypter(jose.A128GCM, jose.Recipient{Algorithm: jose.RSA_OAEP, Key: &j.JWEPrivateKey.PublicKey}, nil)
	if err != nil {
		return "", fmt.Errorf("an error occurred while creating encrypter: %v", err)
	}

	jweObject, err := encryptor.Encrypt([]byte(token))
	if err != nil {
		return "", fmt.Errorf("an error occurred while encrypting JWT: %v", err)
	}

	compact, err := jweObject.CompactSerialize()
	if err != nil {
		return "", fmt.Errorf("an error occurred while serializing JWE: %v", err)
	}

	return compact, nil
}

func (j *JWTHandler) decryptToken(encrypted string) (string, error) {
	parsedCompact, err := jose.ParseEncrypted(encrypted, []jose.KeyAlgorithm{jose.RSA_OAEP}, []jose.ContentEncryption{jose.A128GCM})
	if err != nil {
		return "", fmt.Errorf("an error occurred while parsing JWE: %v", err)
	}

	tokenBytes, err := parsedCompact.Decrypt(&j.JWEPrivateKey)
	if err != nil {
		return "", fmt.Errorf("an error occurred while decrypting JWE: %v", err)
	}

	return string(tokenBytes), nil
}

func (j *JWTHandler) GenerateToken(user *models.AuthUser) (string, error) {
	claims := models.Claims{
		Username: user.Username,
		ID:       user.ID,
		RegisteredClaims: &jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.Lifetime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    j.Issuer,
			Audience:  j.Audience,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.Key))
}

func (j *JWTHandler) ParseToken(token string) (*models.Claims, error) {
	jwtToken, err := jwt.ParseWithClaims(
		token,
		&models.Claims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(j.Key), nil
		},
		jwt.WithoutClaimsValidation(),
	)
	if err != nil {
		return nil, err
	}

	claims, ok := jwtToken.Claims.(*models.Claims)
	if !ok {
		return nil, fmt.Errorf("claims provided was not valid: %+v", err)
	}

	if claims.Issuer != j.Issuer {
		return nil, fmt.Errorf("invalid issuer was provided through token")
	}

	testAud := claims.Audience
	slices.Sort(testAud)
	for i := range testAud {
		if testAud[i] != j.Audience[i] {
			return nil, fmt.Errorf("invalid audience field")
		}
	}

	return claims, nil
}

func (j *JWTHandler) VerifyToken(token string) (*models.Claims, error) {
	claims, err := jwt.ParseWithClaims(
		token,
		&models.Claims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(j.Key), nil
		})

	if err != nil {
		return nil, fmt.Errorf("an error occurred while parsing JWT: %v", err)
	} else if claims, ok := claims.Claims.(*models.Claims); ok {
		return claims, nil
	} else {
		return nil, fmt.Errorf("claims provided was not valid: %v", err)
	}
}
