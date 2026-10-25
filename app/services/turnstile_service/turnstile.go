// Package turnstile_service verifies Cloudflare Turnstile challenge responses.
package turnstile_service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"carf-chorf.com/app/app/kernel"
	"carf-chorf.com/app/pkg/http_client"
)

// SiteverifyURL is Cloudflare's server-side verification endpoint.
const SiteverifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

var (
	// ErrMissingToken means the form did not include a Turnstile response.
	ErrMissingToken = errors.New("turnstile: missing token")
	// ErrVerificationFailed means Cloudflare rejected the token.
	ErrVerificationFailed = errors.New("turnstile: verification failed")

	// Overridable in tests to avoid calling Cloudflare.
	siteverifyURL = SiteverifyURL
	client        = http_client.Shared
)

// verifyResponse mirrors the fields Cloudflare's siteverify endpoint returns.
type verifyResponse struct {
	Success    bool     `json:"success"`
	ErrorCodes []string `json:"error-codes"`
}

// Verify checks a Turnstile response token with Cloudflare. remoteIP is optional
// but lets Cloudflare validate that the token was solved from the same address.
func Verify(ctx context.Context, token, remoteIP string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return ErrMissingToken
	}

	// Cloudflare expects application/x-www-form-urlencoded fields, not JSON.
	form := url.Values{}
	form.Set("secret", kernel.Turnstile.SecretKey)
	form.Set("response", token)
	if remoteIP != "" {
		form.Set("remoteip", remoteIP)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, siteverifyURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result verifyResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("turnstile: decode response: %w", err)
	}

	if !result.Success {
		// Error codes are safe to log and make verification failures diagnosable.
		if len(result.ErrorCodes) > 0 {
			return fmt.Errorf("%w: %s", ErrVerificationFailed, strings.Join(result.ErrorCodes, ", "))
		}
		return ErrVerificationFailed
	}

	return nil
}
