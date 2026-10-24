package session_service

import (
	"crypto/sha256"
	"net/http"
	"sync"

	"carf-chorf.com/app/app/kernel"
	"github.com/gorilla/sessions"
)

const Name = "carf_chorf_session"

var (
	store *sessions.CookieStore
	once  sync.Once
)

func Init() {
	once.Do(func() {
		authKey := sha256.Sum256([]byte(kernel.Server.SessionSecret + "|auth"))
		encKey := sha256.Sum256([]byte(kernel.Server.SessionSecret + "|enc"))

		store = sessions.NewCookieStore(authKey[:], encKey[:])
		store.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   0,
			HttpOnly: true,
			Secure:   !kernel.IsDev(),
			SameSite: http.SameSiteLaxMode,
		}
	})
}

func Get(r *http.Request) (*sessions.Session, error) {
	Init()
	return store.Get(r, Name)
}
