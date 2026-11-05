package auth

import (
	"net/url"
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

// safeNext only allows local paths so the post-login redirect cannot leave the
// site. Invalid or external values fall back to the dashboard.
func safeNext(next string) string {
	next = strings.TrimSpace(next)
	if next == "" || !strings.HasPrefix(next, "/") || strings.HasPrefix(next, "//") {
		return "/dashboard"
	}

	u, err := url.Parse(next)
	if err != nil || u.Scheme != "" || u.Host != "" {
		return "/dashboard"
	}

	// Decode percent-encoded slashes/backslashes as well so values such as
	// /%2f%2fevil.example or /%255cevil.example cannot become open redirects.
	path := u.Path
	for i := 0; i < 2; i++ {
		decoded, err := url.PathUnescape(path)
		if err != nil || decoded == path {
			break
		}
		path = decoded
	}
	if strings.HasPrefix(path, "//") || strings.ContainsRune(path, '\\') {
		return "/dashboard"
	}

	return next
}
