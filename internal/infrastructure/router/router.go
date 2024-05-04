package router

import (
	"github.com/NikolosHGW/gophermart/internal/app/handler"
	"github.com/go-chi/chi"
)

func NewRouter(handlers *handler.Handlers) *chi.Mux {
	router := chi.NewRouter()

	router.Route("/api/user", func(r chi.Router) {
		r.Post("/register", handlers.UserHandler.RegisterUser)
	})

	return router
}
