package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	auth_service "carf-chorf.com/app/app/services/auth"
	flashes_service "carf-chorf.com/app/app/services/flashes"
	"carf-chorf.com/app/pkg/renderer"
	"carf-chorf.com/app/pkg/simple_logger"
	"github.com/flosch/pongo2/v7"
)

const (
	loginTemplate      = "auth/login.html"
	getStartedTemplate = "auth/getstarted.html"
	otpTemplate        = "auth/otp.html"
)

func LoginPage(w http.ResponseWriter, r *http.Request) {
	renderEmailForm(w, r, loginTemplate, "", http.StatusOK, "")
}

func LoginSubmit(w http.ResponseWriter, r *http.Request) {
	email := normalizeEmail(r.PostFormValue("email"))
	if err := validateEmail(email); err != nil {
		renderEmailForm(w, r, loginTemplate, email, http.StatusUnprocessableEntity, err.Error())
		return
	}

	user, err := findUserByEmail(email)
	if err != nil {
		simple_logger.Errorf("auth: find user: %v", err)
		renderEmailForm(w, r, loginTemplate, email, http.StatusInternalServerError, "Something went wrong. Please try again.")
		return
	}

	_ = auth_service.ClearChallenge(w, r)

	send := user != nil && user.IsActive
	if err := beginChallenge(w, r, IntentLogin, email, send); err != nil {
		simple_logger.Errorf("auth: begin login challenge: %v", err)
		renderEmailForm(w, r, loginTemplate, email, http.StatusInternalServerError, "We could not send the code. Please try again.")
		return
	}

	http.Redirect(w, r, "/otp", http.StatusSeeOther)
}

func GetStartedPage(w http.ResponseWriter, r *http.Request) {
	renderEmailForm(w, r, getStartedTemplate, "", http.StatusOK, "")
}

func GetStartedSubmit(w http.ResponseWriter, r *http.Request) {
	email := normalizeEmail(r.PostFormValue("email"))
	if err := validateEmail(email); err != nil {
		renderEmailForm(w, r, getStartedTemplate, email, http.StatusUnprocessableEntity, err.Error())
		return
	}

	_ = auth_service.ClearChallenge(w, r)

	if err := beginChallenge(w, r, IntentRegister, email, true); err != nil {
		simple_logger.Errorf("auth: begin register challenge: %v", err)
		renderEmailForm(w, r, getStartedTemplate, email, http.StatusInternalServerError, "We could not send the code. Please try again.")
		return
	}

	http.Redirect(w, r, "/otp", http.StatusSeeOther)
}

func OTPPage(w http.ResponseWriter, r *http.Request) {
	challenge, ok := auth_service.Challenge(r)
	if !ok {
		flashes_service.AddFlash(w, r, "error", "Your verification code expired. Please request a new one.")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	renderOTP(w, r, challenge, http.StatusOK, "")
}

func ResendOTP(w http.ResponseWriter, r *http.Request) {
	challenge, ok := auth_service.Challenge(r)
	if !ok {
		flashes_service.AddFlash(w, r, "error", "Your verification code expired. Please request a new one.")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err := resendChallenge(w, r, challenge); err != nil {
		if errors.Is(err, ErrResendCooldown) {
			flashes_service.AddFlash(w, r, "error", "Please wait a minute before requesting another code.")
		} else {
			simple_logger.Errorf("auth: resend challenge: %v", err)
			flashes_service.AddFlash(w, r, "error", "We could not send a new code. Please try again.")
		}
		http.Redirect(w, r, "/otp", http.StatusSeeOther)
		return
	}

	flashes_service.AddFlash(w, r, "success", "If an account exists with that email, a new code has been sent.")
	http.Redirect(w, r, "/otp", http.StatusSeeOther)
}

func Verify(w http.ResponseWriter, r *http.Request) {
	challenge, ok := auth_service.Challenge(r)
	if !ok {
		flashes_service.AddFlash(w, r, "error", "Your verification code expired. Please request a new one.")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	code := normalizeCode(r.PostFormValue("code"))
	if err := validateCode(code); err != nil {
		flashes_service.AddFlash(w, r, "error", err.Error())
		http.Redirect(w, r, "/otp", http.StatusSeeOther)
		return
	}

	switch err := VerifyCode(challenge, code); {
	case errors.Is(err, ErrCodeExpired):
		flashes_service.AddFlash(w, r, "error", "Your verification code expired. Please request a new one.")
		http.Redirect(w, r, "/otp", http.StatusSeeOther)
		return
	case errors.Is(err, ErrInvalidCode):
		challenge.Attempts--
		_ = auth_service.UpdateChallenge(w, r, challenge)
		if challenge.Attempts <= 0 {
			flashes_service.AddFlash(w, r, "error", "Too many incorrect attempts. Please request a new code.")
		} else {
			flashes_service.AddFlash(w, r, "error", fmt.Sprintf("Incorrect code. %d attempts remaining.", challenge.Attempts))
		}
		http.Redirect(w, r, "/otp", http.StatusSeeOther)
		return
	case err != nil:
		flashes_service.AddFlash(w, r, "error", "Your verification code is no longer valid. Please request a new one.")
		http.Redirect(w, r, "/otp", http.StatusSeeOther)
		return
	}

	user, err := findUserByEmail(challenge.Email)
	if err != nil {
		simple_logger.Errorf("auth: find user after verification: %v", err)
		flashes_service.AddFlash(w, r, "error", "Something went wrong. Please try again.")
		http.Redirect(w, r, "/otp", http.StatusSeeOther)
		return
	}

	if user == nil {
		if challenge.Intent != IntentRegister {
			_ = auth_service.ClearChallenge(w, r)
			flashes_service.AddFlash(w, r, "error", "We could not verify that code. Please try again.")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		user, err = createUser(challenge.Email)
		if err != nil {
			simple_logger.Errorf("auth: create user: %v", err)
			flashes_service.AddFlash(w, r, "error", "Something went wrong. Please try again.")
			http.Redirect(w, r, "/otp", http.StatusSeeOther)
			return
		}
	}

	if !user.IsActive {
		_ = auth_service.ClearChallenge(w, r)
		flashes_service.AddFlash(w, r, "error", "This account is not active.")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	state := auth_service.State{UserID: user.ID, Email: user.Email}
	if err := auth_service.SetAuthenticated(w, r, state); err != nil {
		simple_logger.Errorf("auth: set session: %v", err)
		flashes_service.AddFlash(w, r, "error", "Something went wrong. Please try again.")
		http.Redirect(w, r, "/otp", http.StatusSeeOther)
		return
	}

	message := "Welcome back!"
	if challenge.Intent == IntentRegister {
		message = "Welcome! Your account is ready."
	}
	flashes_service.AddFlash(w, r, "success", message)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	if err := auth_service.Logout(w, r); err != nil {
		simple_logger.Errorf("auth: logout: %v", err)
	}

	flashes_service.AddFlash(w, r, "success", "You have been logged out.")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func renderEmailForm(w http.ResponseWriter, r *http.Request, template, email string, status int, errMsg string) {
	ctx := pongo2.Context{
		"email": email,
	}

	if errMsg != "" {
		ctx["errors"] = []string{errMsg}
	}

	renderer.Pongo(w, r, template, status, ctx)
}

func renderOTP(w http.ResponseWriter, r *http.Request, challenge auth_service.OTPChallenge, status int, errMsg string) {
	now := time.Now().Unix()
	expired := now > challenge.ExpiresAt
	exhausted := challenge.Attempts <= 0

	ctx := pongo2.Context{
		"masked_email":  maskEmail(challenge.Email),
		"attempts_left": challenge.Attempts,
		"expired":       expired,
		"exhausted":     exhausted,
		"can_verify":    !expired && !exhausted,
	}

	if errMsg != "" {
		ctx["errors"] = []string{errMsg}
	}

	renderer.Pongo(w, r, otpTemplate, status, ctx)
}

func maskEmail(email string) string {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 || parts[0] == "" {
		return email
	}

	local := parts[0]
	if len(local) <= 2 {
		return local[:1] + "***@" + parts[1]
	}
	return local[:2] + "***@" + parts[1]
}
