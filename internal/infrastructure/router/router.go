package router

import (
	"github.com/NikolosHGW/gophermart/internal/app/handler"
	"github.com/NikolosHGW/gophermart/internal/infrastructure/middleware"
	"github.com/go-chi/chi"
)

func NewRouter(handlers *handler.Handlers, middlewares *middleware.Middlewares) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middlewares.Logger.WithLogging)
	r.Use(middlewares.Gzip.WithGzip)

	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", handlers.UserHandler.RegisterUser)
		r.Post("/login", handlers.UserHandler.LoginUser)
	})

	return r
}
