package server

import (
	"net/http"
	"slices"

	"github.com/gofs-cli/gofs/templates/fs-app/internal/auth"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (s *Server) assetsMiddlewares(h http.Handler) http.Handler {
	middlewares := []func(http.Handler) http.Handler{
		middleware.Logger,
		cors.Handler(cors.Options{
			AllowedOrigins:   s.conf.AllowedOrigins,
			AllowedMethods:   []string{http.MethodGet, http.MethodOptions},
			AllowedHeaders:   []string{"X-CSRF-Token"},
			AllowCredentials: true,
			MaxAge:           300,
		}),
		middleware.Compress(5),
		middleware.SetHeader("Cache-Control", "max-age=604800, stale-while-revalidate=86400"),
	}
	slices.Reverse(middlewares)
	for _, m := range middlewares {
		h = m(h)
	}
	return h
}

func (s *Server) routeMiddlewares(h http.Handler) http.Handler {
	middlewares := []func(http.Handler) http.Handler{
		middleware.Logger,
		cors.Handler(cors.Options{
			AllowedOrigins:   s.conf.AllowedOrigins,
			AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
			AllowedHeaders:   []string{"Authorization", "X-CSRF-Token"},
			AllowCredentials: true,
			MaxAge:           300,
		}),
		middleware.NoCache,
		middleware.StripSlashes,
		middleware.Recoverer,
		middleware.Compress(5),
		auth.Middleware(s.conf.Env),
	}
	slices.Reverse(middlewares)
	for _, m := range middlewares {
		h = m(h)
	}
	return h
}
