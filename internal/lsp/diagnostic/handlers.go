package diagnostic

import (
	"context"
	"encoding/json"
	"log/slog"
	"path"
	"path/filepath"

	"github.com/gofs-cli/gofs/internal/lsp/jsonrpc2"
	"github.com/gofs-cli/gofs/internal/lsp/protocol"
	"github.com/gofs-cli/gofs/internal/lsp/repo"
)

func Diagnostic(r *repo.Repo) jsonrpc2.Handler {
	return func(ctx context.Context, que chan protocol.Response, req protocol.Request) error {
		// only support valid gofs repos
		if !r.IsValidGofs() {
			que <- protocol.NewEmptyResponse(req.Id, FullDiagnosticResponse{})
			return nil
		}

		p, err := protocol.DecodeParams[DiagnosticRequest](req)
		if err != nil {
			return jsonrpc2.ErrInvalidParams
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
				slog.Error("templ file not found", "path", p.TextDocument.Path)
				return jsonrpc2.ErrInternalError
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
			return jsonrpc2.ErrInternalError
		}

		que <- protocol.NewResponse(req.Id, json.RawMessage(b))
		return nil
	}
}
