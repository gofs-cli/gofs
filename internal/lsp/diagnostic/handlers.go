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
	return func(ctx context.Context, que chan protocol.Response, req protocol.Request) {
		// only support valid gofs repos
		if !r.IsValidGofs() {
			que <- protocol.NewEmptyResponse(req.Id, FullDiagnosticResponse{})
			return
		}

		// decode request
		p, err := protocol.DecodeParams[DiagnosticRequest](req)
		if err != nil {
			que <- protocol.NewResponseError(req.Id, protocol.ResponseError{
				Code:    protocol.ErrorCodeInvalidParams,
				Message: "error converting request to DiagnosticRequest",
			})
			return
		}

		diagnostics := make([]DiagnosticResponse, 0)

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
				que <- protocol.NewResponseError(req.Id, protocol.ResponseError{
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

		b, err := json.Marshal(FullDiagnosticResponse{
			Kind:  KindFull,
			Items: diagnostics,
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
