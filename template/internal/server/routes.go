package server

import (
	"net/http"

	"github.com/gofs-cli/gofs/template/internal/server/assets"
	"github.com/gofs-cli/gofs/template/internal/server/handlers"
	"github.com/gofs-cli/gofs/template/internal/server/handlers/page"
	"github.com/gofs-cli/gofs/template/internal/ui"
	"github.com/gofs-cli/gofs/template/internal/ui/pages/validation"
)

func (s *Server) Routes() {
	// filserver route for assets
	assetMux := http.NewServeMux()
	assetMux.Handle("GET /{path...}", http.StripPrefix("/assets/", handlers.NewHashedAssets(assets.FS)))
	s.r.Handle("GET /assets/{path...}", s.assetsMiddlewares(assetMux))

	// handlers for normal routes with all general middleware
	routesMux := http.NewServeMux()
	routesMux.Handle("GET /{$}", page.Index())
	routesMux.Handle("GET /modal", ui.Modal())

	routesMux.Handle("GET /hello", http.HandlerFunc(hello))

	routesMux.Handle("/validate", validation.HandleNameValidation())

	routesMux.Handle("GET /toast-success", ui.Success())
	routesMux.Handle("GET /toast-info", ui.Info())
	routesMux.Handle("GET /toast-warning", ui.Warning())
	routesMux.Handle("GET /toast-error", ui.Error())

	s.r.Handle("/", s.routeMiddlewares(routesMux))

	s.srv.Handler = s.r
}

func hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World!"))
}
