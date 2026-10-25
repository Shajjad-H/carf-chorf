package auth

import (
	"net/http"
	"time"

	auth_service "carf-chorf.com/app/app/services/auth"
	"carf-chorf.com/app/app/services/mail_service"
)

func beginChallenge(w http.ResponseWriter, r *http.Request, intent, email, next string, send bool) error {
	code, err := GenerateCode()
	if err != nil {
		return err
	}

	if send {
		if err := mail_service.SendOTP(email, code); err != nil {
			return err
		}
	}

	return auth_service.SetChallenge(w, r, NewChallenge(intent, email, code, next, send))
}

func resendChallenge(w http.ResponseWriter, r *http.Request, challenge auth_service.OTPChallenge) error {
	if time.Now().Unix()-challenge.SentAt < int64(ResendCooldown/time.Second) {
		return ErrResendCooldown
	}

	code, err := GenerateCode()
	if err != nil {
		return err
	}

	if challenge.Send {
		if err := mail_service.SendOTP(challenge.Email, code); err != nil {
			return err
		}
	}

	return auth_service.SetChallenge(w, r, NewChallenge(challenge.Intent, challenge.Email, code, challenge.Next, challenge.Send))
}
