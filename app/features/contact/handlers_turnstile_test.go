package contact

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"carf-chorf.com/app/pkg/renderer"
)

func TestContactSubmitRequiresTurnstile(t *testing.T) {
	renderer.InitTemplates("../../../templates")

	form := url.Values{
		"email":   {"user@example.com"},
		"phone":   {"555-123-4567"},
		"message": {"Hello, I need help with my order."},
	}
	req := httptest.NewRequest(http.MethodPost, "/contact", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	Submit(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnprocessableEntity)
	}
	if !strings.Contains(rec.Body.String(), "verification challenge") {
		t.Fatal("response is missing the Turnstile error message")
	}
}
