package repo

import (
	"context"
	"log"
	"path"
	"path/filepath"

	"github.com/google/uuid"

	"github.com/gofs-cli/gofs/internal/lsp/jsonrpc2"
	"github.com/gofs-cli/gofs/internal/lsp/protocol"
)

func DidOpen(r *Repo) jsonrpc2.Handler {
	return func(ctx context.Context, que chan protocol.Response, req protocol.Request) {
		item := Item{Id: uuid.NewString(), Action: ACTION_EDIT}
		r.AccessQueue.AddToQueue(item)
		defer r.AccessQueue.RemoveFromQueue(item)

		// only support valid gofs repos
		if !r.IsValidGofs() {
			return
		}
		t, err := protocol.DecodeParams[DidOpenRequest](req)
		if err != nil {
			log.Printf("error converting request to DidOpenRequest: %s", err)
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
		item := Item{Id: uuid.NewString(), Action: ACTION_EDIT}
		r.AccessQueue.AddToQueue(item)
		defer r.AccessQueue.RemoveFromQueue(item)

		// only support valid gofs repos
		if !r.IsValidGofs() {
			return
		}
		t, err := protocol.DecodeParams[DidChangeRequest](req)
		if err != nil {
			log.Printf("error converting request to DidChangeRequest: %s", err)
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
		item := Item{Id: uuid.NewString(), Action: ACTION_EDIT}
		r.AccessQueue.AddToQueue(item)
		defer r.AccessQueue.RemoveFromQueue(item)

		// only support valid gofs repos
		if !r.IsValidGofs() {
			return
		}
		t, err := protocol.DecodeParams[DidCloseRequest](req)
		if err != nil {
			log.Printf("error converting request to DidCloseRequest: %s", err)
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
