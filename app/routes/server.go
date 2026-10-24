package routes

import (
	"net/http"

	"carf-chorf.com/app/app/features/auth"
	"carf-chorf.com/app/app/features/contact"
	"carf-chorf.com/app/app/features/home"
	"carf-chorf.com/app/app/features/static"
	"carf-chorf.com/app/pkg/print_routes"
	"github.com/go-chi/chi/v5"
)

func ServerRoutes(router chi.Router) {
	handler := http.StripPrefix("/public", static.Handler("public"))
	router.Get("/public/*", handler.ServeHTTP)
	router.Head("/public/*", handler.ServeHTTP)

	router.Get("/", home.Index)

	router.Get("/contact", contact.Show)
	router.Post("/contact", contact.Submit)

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
