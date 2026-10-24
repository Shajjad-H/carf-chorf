package routes

import (
	"net/http"

	"carf-chorf.com/app/app/features/home"
	"carf-chorf.com/app/app/features/static"
	"carf-chorf.com/app/pkg/print_routes"
	"github.com/go-chi/chi/v5"
)

func ServerRoutes(router chi.Router) {
	router.Handle("/public/*", http.StripPrefix("/public", static.Handler("public")))

	router.Get("/", home.Index)

	print_routes.PrintRoutes(router)
}
