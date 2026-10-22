package ip_middleware

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

// middleware that reads the real IP from context (set by ClientIPFromHeader) and writes it into r.RemoteAddr so chi's Logger picks it up.
func WriteRemoteAddr(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ip := middleware.GetClientIP(r.Context()); ip != "" {
			r.RemoteAddr = ip
		}
		next.ServeHTTP(w, r)
	})
}
