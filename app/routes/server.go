package routes

import (
	"net"
	"net/http"
	"strconv"
	"time"

	"carf-chorf.com/app/app/features/auth"
	"carf-chorf.com/app/app/features/checkout"
	"carf-chorf.com/app/app/features/contact"
	"carf-chorf.com/app/app/features/dashboard"
	"carf-chorf.com/app/app/features/home"
	"carf-chorf.com/app/app/features/pages"
	"carf-chorf.com/app/app/features/payment"
	"carf-chorf.com/app/app/features/pro"
	"carf-chorf.com/app/app/features/reports"
	"carf-chorf.com/app/app/features/static"
	"carf-chorf.com/app/app/features/vehicle"
	auth_middleware "carf-chorf.com/app/app/middlewares/auth"
	auth_service "carf-chorf.com/app/app/services/auth"
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

	router.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/public/favicon.ico", http.StatusMovedPermanently)
	})

	router.Get("/", home.Index)

	// A lookup can spend provider credits, so cap it per IP before the handler
	// runs Turnstile verification and the carbpi request.
	router.With(httprate.LimitBy(10, 10*time.Minute, lookupRateKey,
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Too many lookups. Please try again in a few minutes.", http.StatusTooManyRequests)
		}),
	)).Post("/lookup", vehicle.Lookup)

	router.Get("/checkout", checkout.Index)
	router.With(httprate.LimitBy(30, 15*time.Minute, lookupRateKey,
		httprate.WithLimitHandler(rateLimitHandler("Too many checkout attempts. Please try again later.")),
	)).Post("/checkout", checkout.Create)
	router.Get("/payment", payment.Show)
	router.Get("/payment/success", payment.Success)
	router.With(httprate.LimitBy(60, 15*time.Minute, lookupRateKey,
		httprate.WithLimitHandler(rateLimitHandler("Too many payment attempts. Please try again later.")),
	)).Post("/payment/create-order", payment.CreateOrder)
	router.With(httprate.LimitBy(60, 15*time.Minute, lookupRateKey,
		httprate.WithLimitHandler(rateLimitHandler("Too many payment attempts. Please try again later.")),
	)).Post("/payment/capture-order", payment.CaptureOrder)

	router.Get("/reports/view", reports.View)
	router.Get("/reports/content", reports.Content)

	router.With(auth_middleware.RequireAuth).Get("/dashboard", dashboard.Index)
	router.With(auth_middleware.RequireAuth).Get("/dashboard/orders", dashboard.Orders)
	router.With(auth_middleware.RequireAuth).Get("/dashboard/transactions", dashboard.Transactions)
	router.With(auth_middleware.RequireAuth).Get("/dashboard/history", dashboard.LegacyHistory)
	router.With(auth_middleware.RequireAuth).
		With(httprate.LimitBy(30, time.Hour, userRateKey,
			httprate.WithLimitHandler(rateLimitHandler("Too many vehicle lookups. Please try again later.")),
		)).
		Post("/dashboard/report/lookup", dashboard.Lookup)
	router.With(auth_middleware.RequireAuth).Post("/dashboard/report/purchase", dashboard.Purchase)

	router.Get("/contact", contact.Show)
	router.With(httprate.LimitBy(5, 15*time.Minute, lookupRateKey,
		httprate.WithLimitHandler(rateLimitHandler("Too many messages. Please try again later.")),
	)).Post("/contact", contact.Submit)
	router.Get("/pricing", pages.Pricing)
	router.Get("/pro", pro.Index)
	router.Get("/pro/checkout", pro.CheckoutPage)
	router.With(httprate.LimitBy(10, 15*time.Minute, lookupRateKey,
		httprate.WithLimitHandler(rateLimitHandler("Too many checkout attempts. Please try again later.")),
	)).Post("/pro/checkout", pro.Create)
	router.Get("/sample-report", pages.SampleReport)
	router.Get("/terms", pages.Terms)
	router.Get("/cgv", pages.CGV)
	router.Get("/privacy", pages.Privacy)
	router.Get("/legal-notice", pages.LegalNotice)

	router.Get("/login", auth.LoginPage)
	router.With(httprate.LimitBy(10, 15*time.Minute, lookupRateKey,
		httprate.WithLimitHandler(rateLimitHandler("Too many login attempts. Please try again later.")),
	)).Post("/login", auth.LoginSubmit)
	router.Get("/getstarted", auth.GetStartedPage)
	router.With(httprate.LimitBy(10, 15*time.Minute, lookupRateKey,
		httprate.WithLimitHandler(rateLimitHandler("Too many sign-up attempts. Please try again later.")),
	)).Post("/getstarted", auth.GetStartedSubmit)
	router.Get("/otp", auth.OTPPage)
	router.With(httprate.LimitBy(10, 15*time.Minute, lookupRateKey,
		httprate.WithLimitHandler(rateLimitHandler("Too many code requests. Please try again later.")),
	)).Post("/otp", auth.ResendOTP)
	router.With(httprate.LimitBy(30, 15*time.Minute, lookupRateKey,
		httprate.WithLimitHandler(rateLimitHandler("Too many verification attempts. Please try again later.")),
	)).Post("/verify", auth.Verify)
	router.Post("/logout", auth.Logout)

	print_routes.PrintRoutes(router)
}

// lookupRateKey buckets requests by the client IP resolved by the trusted
// proxy middleware. It falls back to RemoteAddr so rate limiting still works
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

func userRateKey(r *http.Request) (string, error) {
	if state, ok := auth_service.Current(r); ok {
		return "user:" + strconv.FormatUint(uint64(state.UserID), 10), nil
	}
	return lookupRateKey(r)
}

func rateLimitHandler(message string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, message, http.StatusTooManyRequests)
	}
}
