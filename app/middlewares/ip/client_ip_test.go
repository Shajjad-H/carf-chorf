package ip_middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResolveClientIPIgnoresHeadersFromUntrustedPeer(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "203.0.113.10:1234"
	req.Header.Set("X-Forwarded-For", "198.51.100.99")

	var got string
	handler := ResolveClientIP("")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.RemoteAddr
	}))
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if got != "203.0.113.10:1234" {
		t.Fatalf("RemoteAddr = %q, want the direct peer address", got)
	}
}

func TestResolveClientIPUsesForwardedForFromTrustedProxy(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	req.Header.Set("X-Forwarded-For", "198.51.100.7, 10.0.0.1")

	var got string
	handler := ResolveClientIP("127.0.0.1/32")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.RemoteAddr
	}))
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if got != "198.51.100.7" {
		t.Fatalf("RemoteAddr = %q, want the first forwarded IP", got)
	}
}

func TestResolveClientIPPrefersCloudflareHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.2:1234"
	req.Header.Set("CF-Connecting-IP", "203.0.113.8")
	req.Header.Set("X-Forwarded-For", "198.51.100.7")

	var got string
	handler := ResolveClientIP("10.0.0.0/8")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.RemoteAddr
	}))
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if got != "203.0.113.8" {
		t.Fatalf("RemoteAddr = %q, want CF-Connecting-IP", got)
	}
}
