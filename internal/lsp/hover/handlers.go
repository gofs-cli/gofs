package hover

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/gofs-cli/gofs/internal/lsp/jsonrpc2"
	"github.com/gofs-cli/gofs/internal/lsp/protocol"
	"github.com/gofs-cli/gofs/internal/lsp/repo"
)

func Hover(r *repo.Repo) jsonrpc2.Handler {
	return func(ctx context.Context, que chan protocol.Response, req protocol.Request) {
		// only support valid gofs repos
		if !r.IsValidGofs() {
			que <- protocol.NewEmptyResponse(req.Id, HoverResponse{})
			return
		}

		// decode request
		p, err := protocol.DecodeParams[HoverRequest](req)
		if err != nil {
			que <- protocol.NewResponseError(req.Id, protocol.ResponseError{
				Code:    protocol.ErrorCodeInvalidParams,
				Message: "error converting request to HoverRequest",
			})
			return
		}

		// only support hover over templ files
		if filepath.Ext(p.TextDocument.Path) != ".templ" {
			que <- protocol.NewEmptyResponse(req.Id, HoverResponse{})
			return
		}

		// get the templ file
		templFile := r.GetTemplFile(p.TextDocument.Path)
		if templFile == nil {
			que <- protocol.NewResponseError(req.Id, protocol.ResponseError{
				Code:    protocol.ErrorCodeInternalError,
				Message: "templ file not found",
			})
			return
		}

		uriIndex := HoveredUri(*templFile, p.Position.Line, p.Position.Character)
		if uriIndex == -1 {
			que <- protocol.NewEmptyResponse(req.Id, HoverResponse{})
			return
		}

		// no route found for uri
		routeIndex := templFile.UrisRouteIndex[uriIndex]
		if routeIndex == -1 {
			que <- protocol.NewEmptyResponse(req.Id, HoverResponse{})
			return
		}

		// uri has a route
		route, _ := r.GetRoute(routeIndex)

		links := fmt.Sprintf("[routes.go](%s/internal/server/routes.go#%v)", r.RootPath, route.Uri.From.Line+1)

		handler := r.GetPkgFunc(route.Pkg, route.Handler.Call)
		if handler != nil {
			links += fmt.Sprintf(" | [%s](%s#%v)", route.Handler.Call, handler.File, handler.Pos.Line+1)
		}

		b, err := json.Marshal(HoverResponseMarkup{
			Contents: protocol.MarkupContent{
				Kind:  "markdown",
				Value: fmt.Sprintf("```go\n\n%s\n\n```\n\ngo to %s", route.Uri.Raw, links),
			},
		})
		if err != nil {
			que <- protocol.NewResponseError(req.Id, protocol.ResponseError{
				Code:    protocol.ErrorCodeInternalError,
				Message: fmt.Sprintf("json marshal error: %s", err),
			})
			return
		}
		que <- protocol.NewResponse(req.Id, json.RawMessage(b))
	}
}
