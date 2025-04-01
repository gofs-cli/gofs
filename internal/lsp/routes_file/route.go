package routesFile

import (
	"github.com/gofs-cli/gofs/internal/lsp/model"
	"github.com/gofs-cli/gofs/internal/lsp/uri"
)

type Handler struct {
	Call string
	From model.Pos
	To   model.Pos
}

type Route struct {
	Uri     uri.Uri
	Handler Handler
	Pkg     string
}
