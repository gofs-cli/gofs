package routesFile

import (
	"errors"

	"github.com/gofs-cli/gofs/internal/lsp/uri"
)

type Routes struct {
	routes []Route
	b      []byte // routes.go file content
}

func (r *Routes) IsValid() bool {
	return r.b != nil
}

func (r *Routes) SetDefault() {
	r.routes = []Route{}
	r.b = []byte{}
}

func (r *Routes) Update(b []byte) {
	r.routes = getRoutes(b)
	r.b = b
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
	for i, route := range r.routes {
		if match := route.Uri.IsMatch(uri); match {
			return i
		}
	}

	return -1
}
