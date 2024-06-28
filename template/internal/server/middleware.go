package server

import (
	"module/placeholder/internal/auth"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (s *Server) assetsMiddlewares(h http.Handler) http.Handler {
	middlewares := []func(http.Handler) http.Handler{
		cors.Handler(cors.Options{
			AllowedOrigins:   s.conf.AllowedOrigins,
			AllowedMethods:   []string{"GET", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Content-Type", "X-CSRF-Token"},
			AllowCredentials: true,
			MaxAge:           300,
		}),
		middleware.Compress(5),
	}
	for _, m := range middlewares {
		h = m(h)
	}
	return h
}

func (s *Server) routeMiddlewares(h http.Handler) http.Handler {
	middlewares := []func(http.Handler) http.Handler{
		cors.Handler(cors.Options{
			AllowedOrigins:   s.conf.AllowedOrigins,
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			AllowCredentials: true,
			MaxAge:           300,
		}),
		middleware.RedirectSlashes,
		middleware.Recoverer,
		middleware.Compress(5),
		middleware.Logger,
		auth.Middleware(s.conf.Local),
	}
	for _, m := range middlewares {
		h = m(h)
	}
	return h
}
