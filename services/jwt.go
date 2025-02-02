package services

import "tranquility/models"

func GenerateToken(user *models.AuthUser) (string, error) {
	return "jwt", nil
}
