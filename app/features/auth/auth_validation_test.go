package auth

import "testing"

func TestNormalizeEmail(t *testing.T) {
	if got := normalizeEmail("  User@Example.COM "); got != "user@example.com" {
		t.Fatalf("normalizeEmail() = %q, want %q", got, "user@example.com")
	}
}

func TestNormalizeCode(t *testing.T) {
	if got := normalizeCode("  ab12cd34 "); got != "AB12CD34" {
		t.Fatalf("normalizeCode() = %q, want %q", got, "AB12CD34")
	}
}

func TestValidateEmail(t *testing.T) {
	valid := []string{"user@example.com", "a.b+tag@sub.example.co"}
	for _, email := range valid {
		if err := validateEmail(email); err != nil {
			t.Errorf("validateEmail(%q) error = %v", email, err)
		}
	}

	invalid := []string{"", "nope", "user@", "@example.com", "user example.com"}
	for _, email := range invalid {
		if err := validateEmail(email); err == nil {
			t.Errorf("validateEmail(%q) = nil, want error", email)
		}
	}
}

func TestSafeNext(t *testing.T) {
	valid := map[string]string{
		"/":                                 "/",
		"/checkout":                         "/checkout",
		"/checkout?token=abc&credits=2":     "/checkout?token=abc&credits=2",
		"  /checkout?token=abc&credits=4  ": "/checkout?token=abc&credits=4",
	}
	for input, want := range valid {
		if got := safeNext(input); got != want {
			t.Errorf("safeNext(%q) = %q, want %q", input, got, want)
		}
	}

	invalid := []string{
		"",
		"https://evil.example",
		"//evil.example",
		"/\\evil.example",
		"/%5cevil.example",
		"evil.example",
		"javascript:alert(1)",
	}
	for _, input := range invalid {
		if got := safeNext(input); got != "/" {
			t.Errorf("safeNext(%q) = %q, want %q", input, got, "/")
		}
	}
}

func TestValidateCode(t *testing.T) {
	for _, code := range []string{"AB12CD34", "12345678"} {
		if err := validateCode(code); err != nil {
			t.Errorf("valid code %q rejected: %v", code, err)
		}
	}

	invalid := []string{"", "AB12CD3", "AB12CD345", "ab12cd34", "AB12 CD4", "AB12-CD4"}
	for _, code := range invalid {
		if err := validateCode(code); err == nil {
			t.Errorf("validateCode(%q) = nil, want error", code)
		}
	}
}
