package auth

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"carf-chorf.com/app/pkg/renderer"
)

func TestAuthFormsRequireTurnstile(t *testing.T) {
	renderer.InitTemplates("../../../templates")

	tests := []struct {
		name    string
		path    string
		handler http.HandlerFunc
	}{
		{name: "login", path: "/login", handler: LoginSubmit},
		{name: "get started", path: "/getstarted", handler: GetStartedSubmit},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			form := url.Values{"email": {"user@example.com"}}
			req := httptest.NewRequest(http.MethodPost, tc.path, strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rec := httptest.NewRecorder()

			tc.handler(rec, req)

			if rec.Code != http.StatusUnprocessableEntity {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnprocessableEntity)
			}
			if !strings.Contains(rec.Body.String(), "verification challenge") {
				t.Fatal("response is missing the Turnstile error message")
			}
		})
	}
}
