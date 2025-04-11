package repo

import (
	"context"
	"path"
	"path/filepath"

	"github.com/gofs-cli/gofs/internal/lsp/jsonrpc2"
	"github.com/gofs-cli/gofs/internal/lsp/protocol"
)

func DidOpen(r *Repo) jsonrpc2.Handler {
	return func(ctx context.Context, que chan protocol.Response, params any, id int) {
		// only support valid gofs repos
		if !r.IsValidGofs() {
			return
		}

		p := params.(*protocol.DidOpenRequest)

		// only support opening templ files
		if filepath.Ext(p.TextDocument.Path) == ".templ" {
			r.OpenTemplFile(*p)
		}
	}
}

func DidChange(r *Repo) jsonrpc2.Handler {
	return func(ctx context.Context, que chan protocol.Response, params any, id int) {
		// only support valid gofs repos
		if !r.IsValidGofs() {
			return
		}

		p := params.(*protocol.DidChangeRequest)

		// replace templ file content
		if filepath.Ext(p.TextDocument.Path) == ".templ" {
			r.ChangeTemplFile(*p)
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
	return func(ctx context.Context, que chan protocol.Response, params any, id int) {
		// only support valid gofs repos
		if !r.IsValidGofs() {
			return
		}

		p := params.(*protocol.DidCloseRequest)

		if filepath.Ext(p.TextDocument.Path) != ".templ" {
			return
		}
		r.CloseTemplFile(*p)
	}
}

func DidSave(r *Repo) jsonrpc2.Handler {
	return func(ctx context.Context, que chan protocol.Response, params any, id int) {
		// do nothing
	}
}
