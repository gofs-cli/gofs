package repo

import (
	"context"
	"path"
	"path/filepath"

	"github.com/gofs-cli/gofs/internal/lsp/jsonrpc2"
	"github.com/gofs-cli/gofs/internal/lsp/protocol"
)

func DidOpen(r *Repo) jsonrpc2.Handler {
	return func(ctx context.Context, que chan protocol.Response, req protocol.Request) error {
		// only support valid gofs repos
		if !r.IsValidGofs() {
			return nil
		}

		p, err := protocol.DecodeParams[DidOpenRequest](req)
		if err != nil {
			return jsonrpc2.ErrInvalidParams
		}

		// only support opening templ files
		if filepath.Ext(p.TextDocument.Path) == ".templ" {
			r.OpenTemplFile(*p)
		}

		return nil
	}
}

func DidChange(r *Repo) jsonrpc2.Handler {
	return func(ctx context.Context, que chan protocol.Response, req protocol.Request) error {
		// only support valid gofs repos
		if !r.IsValidGofs() {
			return nil
		}

		p, err := protocol.DecodeParams[DidChangeRequest](req)
		if err != nil {
			return jsonrpc2.ErrInvalidParams
		}

		// replace templ file content
		if filepath.Ext(p.TextDocument.Path) == ".templ" {
			r.ChangeTemplFile(*p)
			return nil
		}
		// else check if routes file changed
		if path.Base(p.TextDocument.Path) == "routes.go" {
			b := []byte(p.ContentChanges[0].Text)
			r.UpdateRoutes(b)
			return nil
		}

		return nil
	}
}

func DidClose(r *Repo) jsonrpc2.Handler {
	return func(ctx context.Context, que chan protocol.Response, req protocol.Request) error {
		// only support valid gofs repos
		if !r.IsValidGofs() {
			return nil
		}

		p, err := protocol.DecodeParams[DidCloseRequest](req)
		if err != nil {
			return jsonrpc2.ErrInvalidParams
		}

		if filepath.Ext(p.TextDocument.Path) != ".templ" {
			return nil
		}
		r.CloseTemplFile(*p)

		return nil
	}
}

func DidSave(r *Repo) jsonrpc2.Handler {
	return func(ctx context.Context, que chan protocol.Response, req protocol.Request) error {
		// do nothing
		return nil
	}
}
