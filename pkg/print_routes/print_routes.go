package print_routes

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func PrintRoutes(r chi.Routes) {
	chi.Walk(r, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		fmt.Printf("%-7s %s\n", method, route)
		return nil
	})
}
