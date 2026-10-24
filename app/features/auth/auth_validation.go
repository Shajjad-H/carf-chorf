package auth

import (
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func normalizeCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

func validateEmail(email string) error {
	return validation.Validate(email,
		validation.Required.Error("Email is required"),
		is.EmailFormat.Error("Enter a valid email address"),
	)
}

func validateCode(code string) error {
	return validation.Validate(code,
		validation.Required.Error("Code is required"),
		validation.Match(codeRegex).Error("Enter the 8-character code"),
	)
}
