// Package admin_auth manages the dashboard's isolated admin session.
package admin_auth

import (
	"crypto/sha256"
	"net/http"
	"sync"

	"carf-chorf.com/app/app/kernel"
	"github.com/gorilla/sessions"
)

const (
	cookieName = "carf_chorf_admin"
	sessionKey = "admin"
)

var (
	store *sessions.CookieStore
	once  sync.Once
)

func Init() {
	once.Do(func() {
		authKey := sha256.Sum256([]byte(kernel.Server.SessionSecret + "|admin-auth"))
		encKey := sha256.Sum256([]byte(kernel.Server.SessionSecret + "|admin-enc"))
		store = sessions.NewCookieStore(authKey[:], encKey[:])
		store.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   12 * 60 * 60,
			HttpOnly: true,
			Secure:   !kernel.IsDev(),
			SameSite: http.SameSiteLaxMode,
		}
	})
}

func session(r *http.Request) (*sessions.Session, error) {
	Init()
	return store.Get(r, cookieName)
}

// IsAuthenticated reports whether the admin session is marked authenticated.
func IsAuthenticated(r *http.Request) bool {
	sess, err := session(r)
	if err != nil {
		return false
	}
	value, ok := sess.Values[sessionKey].(bool)
	return ok && value
}

// SetAuthenticated marks the admin session as authenticated.
func SetAuthenticated(w http.ResponseWriter, r *http.Request) error {
	sess, err := session(r)
	if err != nil {
		return err
	}
	sess.Values[sessionKey] = true
	return sess.Save(r, w)
}

// Clear removes the admin session.
func Clear(w http.ResponseWriter, r *http.Request) error {
	sess, err := session(r)
	if err != nil {
		return err
	}
	delete(sess.Values, sessionKey)
	return sess.Save(r, w)
}
