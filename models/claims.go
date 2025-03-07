package models

import "github.com/golang-jwt/jwt/v5"

type Claims struct {
	Username string `json:"username"`
	ID       int32  `json:"id"`
	*jwt.RegisteredClaims
}
