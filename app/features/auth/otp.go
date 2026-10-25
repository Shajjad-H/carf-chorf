package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"math/big"
	"regexp"
	"time"

	auth_service "carf-chorf.com/app/app/services/auth"
)

const (
	CodeLength     = 8
	OTPTTL         = 10 * time.Minute
	MaxAttempts    = 5
	ResendCooldown = 60 * time.Second

	IntentLogin    = "login"
	IntentRegister = "register"
)

const codeAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var (
	ErrCodeExpired     = errors.New("verification code expired")
	ErrTooManyAttempts = errors.New("too many attempts")
	ErrInvalidCode     = errors.New("invalid verification code")
	ErrResendCooldown  = errors.New("please wait before requesting another code")

	codeRegex = regexp.MustCompile(`^[A-Z0-9]{8}$`)
)

func GenerateCode() (string, error) {
	code := make([]byte, CodeLength)
	max := big.NewInt(int64(len(codeAlphabet)))

	for i := range code {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		code[i] = codeAlphabet[n.Int64()]
	}

	return string(code), nil
}

func HashCode(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}

func NewChallenge(intent, email, code, next string, send bool) auth_service.OTPChallenge {
	now := time.Now()
	return auth_service.OTPChallenge{
		Email:     email,
		CodeHash:  HashCode(code),
		Intent:    intent,
		Next:      next,
		ExpiresAt: now.Add(OTPTTL).Unix(),
		SentAt:    now.Unix(),
		Attempts:  MaxAttempts,
		Send:      send,
	}
}

func VerifyCode(challenge auth_service.OTPChallenge, input string) error {
	if time.Now().Unix() > challenge.ExpiresAt {
		return ErrCodeExpired
	}
	if challenge.Attempts <= 0 {
		return ErrTooManyAttempts
	}
	if subtle.ConstantTimeCompare([]byte(HashCode(input)), []byte(challenge.CodeHash)) != 1 {
		return ErrInvalidCode
	}
	return nil
}

func IsValidCodeFormat(code string) bool {
	return codeRegex.MatchString(code)
}
