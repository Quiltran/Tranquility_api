package test

import (
	"testing"
	"tranquility/services"
)

func TestValidatePassword(t *testing.T) {
	password := "1234"
	if isValid := services.VerifyPasswordRequirements(password); isValid {
		t.Fatalf("invalid password was marked as valid: %s", password)
	}

	password = "1234567890"
	if isValid := services.VerifyPasswordRequirements(password); isValid {
		t.Fatalf("invalid password was marked as valid: %s", password)
	}

	password = "Skihflkshelhgj"
	if isValid := services.VerifyPasswordRequirements(password); isValid {
		t.Fatalf("invalid password was marked as valid: %s", password)
	}

	password = "skihflkshelhgj"
	if isValid := services.VerifyPasswordRequirements(password); isValid {
		t.Fatalf("invalid password was marked as valid: %s", password)
	}

	password = "skihflkshelh gS"
	if isValid := services.VerifyPasswordRequirements(password); isValid {
		t.Fatalf("invalid password was marked as valid: %s", password)
	}

	password = "This is 4 valid P#ssword"
	if isValid := services.VerifyPasswordRequirements(password); !isValid {
		t.Fatalf("valid password was marked as invalid: %s", password)
	}
}
