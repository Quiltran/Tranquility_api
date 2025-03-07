package services

import (
	"fmt"
	"slices"
	"time"
	"tranquility/config"
	"tranquility/models"

	"github.com/golang-jwt/jwt/v5"
)

type JWTHandler struct {
	*config.JWTConfig
}

func NewJWTHandler(config *config.JWTConfig) *JWTHandler {
	return &JWTHandler{config}
}

func (j *JWTHandler) GenerateToken(user *models.AuthUser) (string, error) {
	claims := models.Claims{
		user.Username,
		user.ID,
		&jwt.RegisteredClaims{
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
