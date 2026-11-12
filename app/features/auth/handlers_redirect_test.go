package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"carf-chorf.com/app/app/kernel"
	auth_service "carf-chorf.com/app/app/services/auth"
	session_service "carf-chorf.com/app/app/services/session"
)

func TestLoginPageRedirectsAuthenticatedUser(t *testing.T) {
	previous := kernel.Server.SessionSecret
	kernel.Server.SessionSecret = "01234567890123456789012345678901"
	t.Cleanup(func() { kernel.Server.SessionSecret = previous })

	session_service.Init()

	setup := httptest.NewRequest(http.MethodGet, "/login", nil)
	setupRec := httptest.NewRecorder()
	if err := auth_service.SetAuthenticated(setupRec, setup, auth_service.State{
		UserID: 1,
		Email:  "user@example.com",
	}); err != nil {
		t.Fatalf("set authenticated session: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/login?next=%2Fcheckout", nil)
	for _, cookie := range setupRec.Result().Cookies() {
		req.AddCookie(cookie)
	}
	rec := httptest.NewRecorder()

	LoginPage(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	if location := rec.Header().Get("Location"); location != "/checkout" {
		t.Fatalf("Location = %q, want /checkout", location)
	}
}
