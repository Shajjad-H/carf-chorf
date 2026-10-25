package auth

import (
	"strings"
	"testing"
	"time"
)

func TestGenerateCode(t *testing.T) {
	for i := 0; i < 100; i++ {
		code, err := GenerateCode()
		if err != nil {
			t.Fatalf("GenerateCode() error = %v", err)
		}
		if len(code) != CodeLength {
			t.Fatalf("code length = %d, want %d", len(code), CodeLength)
		}
		if !IsValidCodeFormat(code) {
			t.Fatalf("code %q does not match [A-Z0-9]{8}", code)
		}
		for _, r := range code {
			if !strings.ContainsRune(codeAlphabet, r) {
				t.Fatalf("code %q contains invalid character %q", code, r)
			}
		}
	}
}

func TestHashCode(t *testing.T) {
	a := HashCode("ABC12345")
	b := HashCode("ABC12345")
	if a != b {
		t.Fatalf("hash is not deterministic: %q != %q", a, b)
	}
	if a == HashCode("XYZ98765") {
		t.Fatal("different codes produced the same hash")
	}
}

func TestNewChallenge(t *testing.T) {
	before := time.Now().Unix()
	challenge := NewChallenge(IntentRegister, "user@example.com", "ABC12345", "/checkout?credits=2", true)

	if challenge.Send != true {
		t.Fatal("Send flag not stored")
	}
	if challenge.Intent != IntentRegister {
		t.Fatalf("Intent = %q, want %q", challenge.Intent, IntentRegister)
	}
	if challenge.Next != "/checkout?credits=2" {
		t.Fatalf("Next = %q", challenge.Next)
	}
	if challenge.Attempts != MaxAttempts {
		t.Fatalf("Attempts = %d, want %d", challenge.Attempts, MaxAttempts)
	}
	if challenge.ExpiresAt < before+int64(OTPTTL/time.Second) {
		t.Fatal("ExpiresAt was not set to the full TTL")
	}
}

func TestVerifyCode(t *testing.T) {
	valid := NewChallenge(IntentLogin, "user@example.com", "ABC12345", "", true)
	if err := VerifyCode(valid, "ABC12345"); err != nil {
		t.Fatalf("valid code rejected: %v", err)
	}

	wrong := NewChallenge(IntentLogin, "user@example.com", "ABC12345", "", true)
	if err := VerifyCode(wrong, "XYZ98765"); err != ErrInvalidCode {
		t.Fatalf("wrong code error = %v, want ErrInvalidCode", err)
	}

	expired := NewChallenge(IntentLogin, "user@example.com", "ABC12345", "", true)
	expired.ExpiresAt = time.Now().Add(-time.Minute).Unix()
	if err := VerifyCode(expired, "ABC12345"); err != ErrCodeExpired {
		t.Fatalf("expired code error = %v, want ErrCodeExpired", err)
	}

	exhausted := NewChallenge(IntentLogin, "user@example.com", "ABC12345", "", true)
	exhausted.Attempts = 0
	if err := VerifyCode(exhausted, "ABC12345"); err != ErrTooManyAttempts {
		t.Fatalf("exhausted code error = %v, want ErrTooManyAttempts", err)
	}
}
