package main

import (
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
)

func main() {
	startup()

	r := chi.NewRouter()

	r.Use(middleware.ClientIPFromHeader("X-Forwarded-For"))
	r.Use(ip_middleware.WriteRemoteAddr)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	routes.ServerRoutes(r)

	simple_logger.Successf("server running on: %s as %s", kernel.Server.Address, kernel.Server.Env)
	http.ListenAndServe(kernel.Server.Address, r)

}

func startup() {
	kernel.Init()

	renderer.InitTemplates("templates")

	database.GetDB()

	session_service.Init()
}
