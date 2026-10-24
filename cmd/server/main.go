package main

import (
	"crypto/sha256"
	"net/http"

	"carf-chorf.com/app/app/kernel"
	ip_middleware "carf-chorf.com/app/app/middlewares/ip"
	"carf-chorf.com/app/app/routes"
	"carf-chorf.com/app/app/services/database"
	session_service "carf-chorf.com/app/app/services/session"
	"carf-chorf.com/app/pkg/renderer"
	"carf-chorf.com/app/pkg/simple_logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/csrf"
)

func main() {
	startup()

	r := chi.NewRouter()

	r.Use(middleware.ClientIPFromHeader("X-Forwarded-For"))
	r.Use(ip_middleware.WriteRemoteAddr)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	if kernel.IsDev() {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				next.ServeHTTP(w, csrf.PlaintextHTTPRequest(req))
			})
		})
	}

	csrfKey := sha256.Sum256([]byte(kernel.Server.SessionSecret + "|csrf"))
	r.Use(csrf.Protect(csrfKey[:],
		csrf.Secure(!kernel.IsDev()),
		csrf.HttpOnly(true),
		csrf.SameSite(csrf.SameSiteLaxMode),
		csrf.Path("/"),
		csrf.CookieName("carf_chorf_csrf"),
		csrf.FieldName("csrf_token"),
		csrf.MaxAge(12*60*60),
		csrf.ErrorHandler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			simple_logger.Warnf("csrf validation failed: %v", csrf.FailureReason(req))
			http.Error(w, "invalid or missing CSRF token", http.StatusForbidden)
		})),
	))

	routes.ServerRoutes(r)

	simple_logger.Successf("server running on: %s as %s", kernel.Server.Address, kernel.Server.Env)
	http.ListenAndServe(kernel.Server.Address, r)

}

func startup() {
	kernel.Init()

	renderer.InitTemplates("templates")

	if _, err := database.GetDB(); err != nil {
		simple_logger.Fatalf("%v", err)
	}

	if kernel.Database.Migrate {
		if err := database.Migrate(); err != nil {
			simple_logger.Fatalf("migration failed: %v", err)
		}
	}

	session_service.Init()
}
