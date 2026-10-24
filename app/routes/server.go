package routes

import (
	"net/http"

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

	print_routes.PrintRoutes(router)
}
