package checkout

import (
	"errors"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

var (
	ErrGuestCredits    = errors.New("login required for multi-credit packages")
	ErrVehicleRequired = errors.New("a vehicle is required for guest checkout")
	ErrVehicleNotFound = errors.New("vehicle record not found")
	ErrInvalidEmail    = errors.New("enter a valid email address")
	ErrEmailMismatch   = errors.New("email addresses do not match")
	ErrTermsRequired   = errors.New("please accept the terms and conditions")
	ErrInvalidPackage  = errors.New("invalid credit package")
	ErrInactiveUser    = errors.New("account is not active")
)

// PurchaseRequest is the package form payload.
type PurchaseRequest struct {
	Credits       uint
	Email         string
	ConfirmEmail  string
	Token         string
	TermsAccepted bool
}

// normalize trims and lowercases the email and token fields.
func (p *PurchaseRequest) normalize() {
	p.Email = strings.ToLower(strings.TrimSpace(p.Email))
	p.ConfirmEmail = strings.ToLower(strings.TrimSpace(p.ConfirmEmail))
	p.Token = strings.TrimSpace(p.Token)
}

// validateGuest enforces the guest checkout rules: only one credit, tied to
// a vehicle, with an accepted email confirmation and terms.
func (p *PurchaseRequest) validateGuest() error {
	if p.Credits != 1 {
		return ErrGuestCredits
	}
	if p.Token == "" {
		return ErrVehicleRequired
	}
	if err := validation.Validate(p.Email,
		validation.Required.Error(ErrInvalidEmail.Error()),
		is.EmailFormat.Error(ErrInvalidEmail.Error()),
	); err != nil {
		return ErrInvalidEmail
	}
	if p.Email != p.ConfirmEmail {
		return ErrEmailMismatch
	}
	if !p.TermsAccepted {
		return ErrTermsRequired
	}
	return nil
}

// validateAuthenticated ensures the purchase is confirmed by accepting the
// terms. Email comes from the authenticated user.
func (p *PurchaseRequest) validateAuthenticated() error {
	if !p.TermsAccepted {
		return ErrTermsRequired
	}
	return nil
}
