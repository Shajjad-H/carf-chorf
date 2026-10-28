package routes

import (
	"net"
	"net/http"
	"time"

	"carf-chorf.com/app/app/features/auth"
	"carf-chorf.com/app/app/features/checkout"
	"carf-chorf.com/app/app/features/contact"
	"carf-chorf.com/app/app/features/home"
	"carf-chorf.com/app/app/features/pages"
	"carf-chorf.com/app/app/features/payment"
	"carf-chorf.com/app/app/features/static"
	"carf-chorf.com/app/app/features/vehicle"
	"carf-chorf.com/app/pkg/print_routes"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
)

// ServerRoutes registers the public site routes.
func ServerRoutes(router chi.Router) {
	handler := http.StripPrefix("/public", static.Handler("public"))
	router.Get("/public/*", handler.ServeHTTP)
	router.Head("/public/*", handler.ServeHTTP)

	router.Get("/", home.Index)

	// A lookup can spend provider credits, so cap it per IP before the handler
	// runs Turnstile verification and the carbpi request.
	router.With(httprate.LimitBy(3, 10*time.Minute, lookupRateKey,
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Too many lookups. Please try again in a few minutes.", http.StatusTooManyRequests)
		}),
	)).Post("/lookup", vehicle.Lookup)

	router.Get("/checkout", checkout.Index)
	router.Post("/checkout", checkout.Create)
	router.Get("/payment", payment.Show)

	router.Get("/contact", contact.Show)
	router.Post("/contact", contact.Submit)
	router.Get("/pricing", pages.Pricing)
	router.Get("/terms", pages.Terms)
	router.Get("/cgv", pages.CGV)

	router.Get("/login", auth.LoginPage)
	router.Post("/login", auth.LoginSubmit)
	router.Get("/getstarted", auth.GetStartedPage)
	router.Post("/getstarted", auth.GetStartedSubmit)
	router.Get("/otp", auth.OTPPage)
	router.Post("/otp", auth.ResendOTP)
	router.Post("/verify", auth.Verify)
	router.Post("/logout", auth.Logout)

	print_routes.PrintRoutes(router)
}

// lookupRateKey buckets requests by the client IP resolved by the global
// ClientIP middleware. It falls back to RemoteAddr so rate limiting still works
// in tests and when the middleware is not installed.
func lookupRateKey(r *http.Request) (string, error) {
	ip := middleware.GetClientIP(r.Context())
	if ip == "" {
		ip = r.RemoteAddr
		if host, _, err := net.SplitHostPort(ip); err == nil {
			ip = host
		}
	}
	return httprate.CanonicalizeIP(ip), nil
}
