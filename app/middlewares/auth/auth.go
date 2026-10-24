package auth_middleware

import (
	"net/http"

	auth_service "carf-chorf.com/app/app/services/auth"
)

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !auth_service.IsAuthenticated(r) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func CurrentUserID(r *http.Request) (uint, bool) {
	state, ok := auth_service.Current(r)
	if !ok {
		return 0, false
	}
	return state.UserID, true
}
