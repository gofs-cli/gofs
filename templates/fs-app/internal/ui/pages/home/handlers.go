package home

import (
	"net/http"

	"github.com/a-h/templ"

	"github.com/gofs-cli/gofs/templates/fs-app/internal/ui"
	"github.com/gofs-cli/gofs/templates/fs-app/internal/ui/components/header"
)

func Index() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		templ.Handler(ui.IndexPage(layout(header.Header(), body()))).ServeHTTP(w, r)
	})
}
