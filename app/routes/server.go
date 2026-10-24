package routes

import (
	"carf-chorf.com/app/app/features/home"
	"carf-chorf.com/app/pkg/print_routes"
	"github.com/go-chi/chi/v5"
)

func ServerRoutes(router chi.Router) {
	router.Get("/", home.Index)

	print_routes.PrintRoutes(router)
}
