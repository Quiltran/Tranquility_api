package services

import (
	"encoding/base64"
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

	tokenBytes, err := parsedCompact.Decrypt(j.JWEPrivateKey)
	if err != nil {
		return "", fmt.Errorf("an error occurred while decrypting JWE: %v", err)
	}

	return string(tokenBytes), nil
}

func (j *JWTHandler) GenerateToken(user *models.AuthUser) (string, error) {
	encodedUserHandle := base64.StdEncoding.EncodeToString(user.UserHandle)

	claims := models.Claims{
		Username:   user.Username,
		ID:         user.ID,
		UserHandle: encodedUserHandle,
		RegisteredClaims: &jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.Lifetime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    j.Issuer,
			Audience:  j.Audience,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedString, err := token.SignedString([]byte(j.Key))
	if err != nil {
		return "", fmt.Errorf("an error occurred while signing jwt: %v", err)
	}
	encrypted, err := j.encryptToken(signedString)
	if err != nil {
		return "", fmt.Errorf("an error occurred while encrypting jwt: %v", err)
	}

	return encrypted, nil
}

func (j *JWTHandler) ParseToken(token string) (*models.Claims, error) {
	decryptedToken, err := j.decryptToken(token)
	if err != nil {
		return nil, fmt.Errorf("an error occurred while decrypting token: %v", err)
	}
	// We do not create a list of options because we are disabling all validation then doing manual validation.
	jwtToken, err := jwt.ParseWithClaims(
		decryptedToken,
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
	decryptedToken, err := j.decryptToken(token)
	if err != nil {
		return nil, fmt.Errorf("an error occurred while decrypting token while verifying: %v", err)
	}

	parserOptions := make([]jwt.ParserOption, 0)
	for i := range j.Audience {
		parserOptions = append(parserOptions, jwt.WithAudience(j.Audience[i]))
	}
	parserOptions = append(parserOptions, jwt.WithIssuer(j.Issuer))

	claims, err := jwt.ParseWithClaims(
		decryptedToken,
		&models.Claims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(j.Key), nil
		},
		parserOptions...,
	)

	if err != nil {
		return nil, fmt.Errorf("an error occurred while parsing JWT: %v", err)
	} else if claims, ok := claims.Claims.(*models.Claims); ok {
		return claims, nil
	} else {
		return nil, fmt.Errorf("claims provided was not valid: %v", err)
	}
}
