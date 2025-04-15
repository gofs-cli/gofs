package repo

import (
	"context"
	"log/slog"
	"path"
	"path/filepath"

	"github.com/gofs-cli/gofs/internal/lsp/jsonrpc2"
	"github.com/gofs-cli/gofs/internal/lsp/protocol"
)

func DidOpen(r *Repo) jsonrpc2.Handler {
	return func(ctx context.Context, que chan protocol.Response, req protocol.Request) {
		// only support valid gofs repos
		if !r.IsValidGofs() {
			return
		}
		t, err := protocol.DecodeParams[DidOpenRequest](req)
		if err != nil {
			slog.Error("error converting request to DidOpenRequest", "err", err)
			return
		}

		// only support opening templ files
		if filepath.Ext(t.TextDocument.Path) == ".templ" {
			r.OpenTemplFile(*t)
		}
	}
}

func DidChange(r *Repo) jsonrpc2.Handler {
	return func(ctx context.Context, que chan protocol.Response, req protocol.Request) {
		// only support valid gofs repos
		if !r.IsValidGofs() {
			return
		}
		t, err := protocol.DecodeParams[DidChangeRequest](req)
		if err != nil {
			slog.Error("error converting request to DidChangeRequest", "err", err)
			return
		}

		// replace templ file content
		if filepath.Ext(t.TextDocument.Path) == ".templ" {
			r.ChangeTemplFile(*t)
			return
		}
		// else check if routes file changed
		if path.Base(t.TextDocument.Path) == "routes.go" {
			b := []byte(t.ContentChanges[0].Text)
			r.UpdateRoutes(b)
			return
		}
	}
}

func DidClose(r *Repo) jsonrpc2.Handler {
	return func(ctx context.Context, que chan protocol.Response, req protocol.Request) {
		// only support valid gofs repos
		if !r.IsValidGofs() {
			return
		}
		t, err := protocol.DecodeParams[DidCloseRequest](req)
		if err != nil {
			slog.Error("error converting request to DidCloseRequest: %s", "err", err)
			return
		}

		if filepath.Ext(t.TextDocument.Path) != ".templ" {
			return
		}
		r.CloseTemplFile(*t)
	}
}

func DidSave(r *Repo) jsonrpc2.Handler {
	return func(ctx context.Context, que chan protocol.Response, req protocol.Request) {
		// do nothing
	}
}
