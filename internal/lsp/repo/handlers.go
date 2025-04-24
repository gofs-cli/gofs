package repo

import (
	"context"
	"path"
	"path/filepath"

	"github.com/gofs-cli/gofs/internal/lsp/jsonrpc2"
	"github.com/gofs-cli/gofs/internal/lsp/protocol"
)

func DidOpen(r *Repo) jsonrpc2.Handler {
	return func(ctx context.Context, que chan protocol.Response, req protocol.Request, params any) {
		// only support valid gofs repos
		if !r.IsValidGofs() {
			return
		}

		p, ok := params.(DidOpenRequest)
		if !ok {
			que <- protocol.NewResponseError(req.Id, protocol.ResponseError{
				Code:    protocol.ErrorCodeInvalidParams,
				Message: "error converting request to DidOpenRequest",
			})
			return
		}

		// only support opening templ files
		if filepath.Ext(p.TextDocument.Path) == ".templ" {
			r.OpenTemplFile(p)
		}
	}
}

func DidChange(r *Repo) jsonrpc2.Handler {
	return func(ctx context.Context, que chan protocol.Response, req protocol.Request, params any) {
		// only support valid gofs repos
		if !r.IsValidGofs() {
			return
		}

		p, ok := params.(DidChangeRequest)
		if !ok {
			que <- protocol.NewResponseError(req.Id, protocol.ResponseError{
				Code:    protocol.ErrorCodeInvalidParams,
				Message: "error converting request to DidChangeRequest",
			})
			return
		}

		// replace templ file content
		if filepath.Ext(p.TextDocument.Path) == ".templ" {
			r.ChangeTemplFile(p)
			return
		}
		// else check if routes file changed
		if path.Base(p.TextDocument.Path) == "routes.go" {
			b := []byte(p.ContentChanges[0].Text)
			r.UpdateRoutes(b)
			return
		}
	}
}

func DidClose(r *Repo) jsonrpc2.Handler {
	return func(ctx context.Context, que chan protocol.Response, req protocol.Request, params any) {
		// only support valid gofs repos
		if !r.IsValidGofs() {
			return
		}

		p, ok := params.(DidCloseRequest)
		if !ok {
			que <- protocol.NewResponseError(req.Id, protocol.ResponseError{
				Code:    protocol.ErrorCodeInvalidParams,
				Message: "error converting request to DidCloseRequest",
			})
			return
		}

		if filepath.Ext(p.TextDocument.Path) != ".templ" {
			return
		}
		r.CloseTemplFile(p)
	}
}

func DidSave(r *Repo) jsonrpc2.Handler {
	return func(ctx context.Context, que chan protocol.Response, req protocol.Request, params any) {
		// do nothing
	}
}
