package checkout

import (
	"errors"
	"testing"
)

func TestValidateGuest(t *testing.T) {
	valid := PurchaseRequest{
		Credits:       3,
		Email:         "User@Example.com",
		ConfirmEmail:  "user@example.com",
		Token:         "token",
		TermsAccepted: true,
	}
	valid.normalize()
	if err := valid.validateGuest(); err != nil {
		t.Fatalf("valid guest purchase rejected: %v", err)
	}

	tests := []struct {
		name string
		req  PurchaseRequest
		want error
	}{
		{
			name: "multi credit allowed",
			req:  PurchaseRequest{Credits: 5, Email: "user@example.com", ConfirmEmail: "user@example.com", Token: "token", TermsAccepted: true},
		},
		{
			name: "missing vehicle",
			req:  PurchaseRequest{Credits: 3, Email: "user@example.com", ConfirmEmail: "user@example.com", TermsAccepted: true},
			want: ErrVehicleRequired,
		},
		{
			name: "invalid email",
			req:  PurchaseRequest{Credits: 3, Email: "nope", ConfirmEmail: "nope", Token: "token", TermsAccepted: true},
			want: ErrInvalidEmail,
		},
		{
			name: "email mismatch",
			req:  PurchaseRequest{Credits: 3, Email: "user@example.com", ConfirmEmail: "other@example.com", Token: "token", TermsAccepted: true},
			want: ErrEmailMismatch,
		},
		{
			name: "terms not accepted",
			req:  PurchaseRequest{Credits: 3, Email: "user@example.com", ConfirmEmail: "user@example.com", Token: "token"},
			want: ErrTermsRequired,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.req.normalize()
			if err := tc.req.validateGuest(); !errors.Is(err, tc.want) {
				t.Fatalf("ValidateGuest() error = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestValidateAuthenticated(t *testing.T) {
	valid := PurchaseRequest{Credits: 5, TermsAccepted: true}
	if err := valid.validateAuthenticated(); err != nil {
		t.Fatalf("valid authenticated purchase rejected: %v", err)
	}

	if err := (&PurchaseRequest{Credits: 3}).validateAuthenticated(); !errors.Is(err, ErrTermsRequired) {
		t.Fatalf("terms error = %v", err)
	}
}
