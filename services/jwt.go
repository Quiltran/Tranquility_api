package services

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"tranquility/models"

	"github.com/golang-jwt/jwt/v5"
)

var (
	lifetime time.Duration = time.Duration(2 * time.Minute)
	issuer   string        = "Tranquility"
	audience []string      = []string{"Tranquility"}
	key      string
)

type Claims struct {
	Username string `json:"username"`
	ID       int32  `json:"id"`
	*jwt.RegisteredClaims
}

func LoadJWTSettings() error {
	lifetimeSetting := os.Getenv("JWT_LIFETIME")
	if lifetimeSetting != "" {
		l, err := strconv.ParseInt(lifetimeSetting, 10, 64)
		if err != nil {
			return fmt.Errorf("an error occurred while loading jwt lifetime: %v", err)
		}
		lifetime = time.Duration(time.Duration(l) * time.Minute)
	}

	issuerSetting := os.Getenv("JWT_ISSUER")
	if issuerSetting != "" {
		issuer = issuerSetting
	}

	audienceSetting := os.Getenv("JWT_AUDIENCE")
	if audienceSetting != "" {
		audience = strings.Split(audienceSetting, ",")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return errors.New("JWT_SECRET was not set")
	} else {
		key = jwtSecret
	}

	return nil
}

func GenerateToken(user *models.AuthUser) (string, error) {
	claims := Claims{
		user.Username,
		user.ID,
		&jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(lifetime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    issuer,
			Audience:  audience,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(key))
}

func VerifyToken(token string) (*Claims, error) {
	claims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(key), nil
	})

	if err != nil {
		return nil, fmt.Errorf("an error occurred while parsing JWT: %v", err)
	} else if claims, ok := claims.Claims.(*Claims); ok {
		return claims, nil
	} else {
		return nil, fmt.Errorf("claims provided was not valid: %v", err)
	}
}
