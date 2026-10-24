package routes

import (
	"carf-chorf.com/app/pkg/print_routes"
	"github.com/go-chi/chi/v5"
)

func ServerRoutes(router chi.Router) {

	print_routes.PrintRoutes(router)
}
