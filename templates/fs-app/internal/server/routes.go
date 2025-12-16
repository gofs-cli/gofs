package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gofs-cli/gofs/templates/fs-app/internal/repository"
	"github.com/gofs-cli/gofs/templates/fs-app/internal/server/assets"
	"github.com/gofs-cli/gofs/templates/fs-app/internal/server/handlers"
	"github.com/gofs-cli/gofs/templates/fs-app/internal/ui/pages/home"
	"github.com/gofs-cli/gofs/templates/fs-app/internal/ui/pages/notfound"
)

func (s *Server) Routes() {
	// filserver route for assets
	assetMux := http.NewServeMux()
	assetMux.Handle("GET /{path...}", http.StripPrefix("/assets/", handlers.NewHashedAssets(assets.FS)))
	s.r.Handle("GET /assets/{path...}", s.assetsMiddlewares(assetMux))

	// handlers for normal routes with all general middleware
	routesMux := http.NewServeMux()
	routesMux.Handle("GET /{$}", home.Index())
	routesMux.Handle("GET /", notfound.Index())

	s.r.Handle("/", s.routeMiddlewares(routesMux))

	s.r.Handle("GET /users", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		users, err := s.repo.GetUsers(r.Context())
		if err != nil {
			http.Error(w, "failed to get users", http.StatusInternalServerError)
			return
		}
		err = json.NewEncoder(w).Encode(users)
		if err != nil {
			http.Error(w, "failed to encode users", http.StatusInternalServerError)
			return
		}
	}))

	s.r.Handle("GET /insertusers", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var user repository.InsertUserParams
		user.Email = "test@example.com"
		user.Name = "Test User"
		if err := s.repo.InsertUser(r.Context(), user); err != nil {
			log.Println("server: error inserting user:", err)
			http.Error(w, "failed to insert user", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	// ensure the server uses the updated handler with all routes and middleware
	s.srv.Handler = s.r
}
