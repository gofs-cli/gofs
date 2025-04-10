package diagnostic

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"path/filepath"

	"github.com/gofs-cli/gofs/internal/lsp/jsonrpc2"
	"github.com/gofs-cli/gofs/internal/lsp/protocol"
	"github.com/gofs-cli/gofs/internal/lsp/repo"
)

func Diagnostic(r *repo.Repo) jsonrpc2.Handler {
	return func(ctx context.Context, que chan protocol.Response, params any, id int) {
		// only support valid gofs repos
		if !r.IsValidGofs() {
			que <- protocol.NewEmptyResponse(id, protocol.FullDiagnosticResponse{})
			return
		}

		p := params.(*protocol.DiagnosticRequest)

		diagnostics := make([]protocol.DiagnosticResponse, 0)

		if path.Base(p.TextDocument.Path) == "routes.go" {
			for _, route := range r.Routes() {
				if len(route.Uri.Diag) == 0 {
					continue
				}
				diagnostics = append(diagnostics, UriDiagnostic(route.Uri)...)
			}
		}

		if filepath.Ext(p.TextDocument.Path) == ".templ" {
			// get the templ file
			templFile := r.GetTemplFile(p.TextDocument.Path)
			if templFile == nil {
				que <- protocol.NewResponseError(id, protocol.ResponseError{
					Code:    protocol.ErrorCodeInternalError,
					Message: "templ file not found",
				})
				return
			}

			for _, uri := range templFile.Uris {
				if len(uri.Diag) == 0 {
					continue
				}
				diagnostics = append(diagnostics, UriDiagnostic(uri)...)
			}
		}

		b, err := json.Marshal(protocol.FullDiagnosticResponse{
			Kind:  protocol.KindFull,
			Items: diagnostics,
		})
		if err != nil {
			que <- protocol.NewResponseError(id, protocol.ResponseError{
				Code:    protocol.ErrorCodeInternalError,
				Message: fmt.Sprintf("json marshal error: %s", err),
			})
			return
		}
		que <- protocol.NewResponse(id, json.RawMessage(b))
	}
}
