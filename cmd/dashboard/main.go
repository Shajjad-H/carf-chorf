package main

import (
	"crypto/sha256"
	"net/http"
	"time"

	"carf-chorf.com/app/app/features/admin"
	"carf-chorf.com/app/app/features/static"
	"carf-chorf.com/app/app/kernel"
	ip_middleware "carf-chorf.com/app/app/middlewares/ip"
	"carf-chorf.com/app/app/services/admin_auth"
	"carf-chorf.com/app/app/services/database"
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

	csrfKey := sha256.Sum256([]byte(kernel.Server.SessionSecret + "|admin-csrf"))
	r.Use(csrf.Protect(csrfKey[:],
		csrf.Secure(!kernel.IsDev()),
		csrf.HttpOnly(true),
		csrf.SameSite(csrf.SameSiteLaxMode),
		csrf.Path("/"),
		csrf.CookieName("carf_chorf_admin_csrf"),
		csrf.FieldName("csrf_token"),
		csrf.MaxAge(12*60*60),
		csrf.ErrorHandler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			simple_logger.Warnf("admin csrf validation failed: %v", csrf.FailureReason(req))
			http.Error(w, "invalid or missing CSRF token", http.StatusForbidden)
		})),
	))

	handler := http.StripPrefix("/public", static.Handler("public"))
	r.Get("/public/*", handler.ServeHTTP)
	r.Head("/public/*", handler.ServeHTTP)

	limiter := admin_auth.NewFailureLimiter(5, 5*time.Minute)
	r.Get("/login", admin.LoginPage)
	r.Post("/login", admin.LoginSubmit(limiter))
	r.Post("/logout", admin.Logout)

	r.Group(func(r chi.Router) {
		r.Use(admin.RequireAuth)
		r.Get("/", admin.Index)
		r.Get("/transactions", admin.Transactions)
		r.Get("/orders", admin.Orders)
		r.Get("/daily-transactions", admin.DailyTransactions)
		r.Get("/daily-report", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/daily-transactions", http.StatusMovedPermanently)
		})
		r.Get("/users", admin.Users)
		r.Post("/users/remind", admin.RemindUsers)
	})

	simple_logger.Successf("admin dashboard running on: %s as %s", kernel.Dashboard.Address, kernel.Server.Env)
	if err := http.ListenAndServe(kernel.Dashboard.Address, r); err != nil {
		simple_logger.Fatalf("admin server: %v", err)
	}
}

func startup() {
	kernel.InitDashboard()
	renderer.InitTemplates("templates")

	if _, err := database.GetDB(); err != nil {
		simple_logger.Fatalf("%v", err)
	}
	if kernel.Database.Migrate {
		if err := database.Migrate(); err != nil {
			simple_logger.Fatalf("migration failed: %v", err)
		}
	}

	admin_auth.Init()
}
