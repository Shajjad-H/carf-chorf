package turnstile_service

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"carf-chorf.com/app/app/kernel"
)

func TestVerifySuccess(t *testing.T) {
	prevSecret := kernel.Turnstile.SecretKey
	kernel.Turnstile.SecretKey = "test-secret"
	defer func() { kernel.Turnstile.SecretKey = prevSecret }()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm() error = %v", err)
		}
		if got := r.PostFormValue("secret"); got != "test-secret" {
			t.Fatalf("secret = %q", got)
		}
		if got := r.PostFormValue("response"); got != "test-token" {
			t.Fatalf("response = %q", got)
		}
		if got := r.PostFormValue("remoteip"); got != "192.0.2.1" {
			t.Fatalf("remoteip = %q", got)
		}
		_, _ = io.WriteString(w, `{"success":true}`)
	}))
	defer server.Close()

	prevURL, prevClient := siteverifyURL, client
	defer func() { siteverifyURL, client = prevURL, prevClient }()
	siteverifyURL = server.URL
	client = server.Client()

	if err := Verify(context.Background(), "test-token", "192.0.2.1"); err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
}

func TestVerifyFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"success":false,"error-codes":["invalid-input-response"]}`)
	}))
	defer server.Close()

	prevURL, prevClient := siteverifyURL, client
	defer func() { siteverifyURL, client = prevURL, prevClient }()
	siteverifyURL = server.URL
	client = server.Client()

	err := Verify(context.Background(), "bad-token", "")
	if !errors.Is(err, ErrVerificationFailed) {
		t.Fatalf("expected ErrVerificationFailed, got %v", err)
	}
}

func TestVerifyMissingToken(t *testing.T) {
	if err := Verify(context.Background(), "  ", ""); !errors.Is(err, ErrMissingToken) {
		t.Fatalf("expected ErrMissingToken, got %v", err)
	}
}
