package routesFile

import (
	"errors"

	"github.com/gofs-cli/gofs/internal/lsp/model"
	"github.com/gofs-cli/gofs/internal/lsp/uri"
)

type Routes struct {
	routes []Route
	b      []byte // routes.go file content
}

func NewRoutes(b []byte) *Routes {
	r := getRoutes(b)
	// TODO: return nil if parsing fails
	return &Routes{
		b:      b,
		routes: r,
	}
}

func (r *Routes) Routes() []Route {
	return r.routes
}

func (r *Routes) GetRoute(i int) (*Route, error) {
	if i > len(r.Routes()) {
		return nil, errors.New("invalid index")
	}
	return &r.routes[i], nil
}

func (r *Routes) RouteIndex(uri uri.Uri) int {
	bestIndex := -1
	bestLevel := model.NoMatch

	for i, route := range r.routes {
		if level := route.Uri.MatchLevel(uri); level > bestLevel {
			bestLevel = level
			bestIndex = i
		}
	}

	return bestIndex
}
