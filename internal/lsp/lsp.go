package lsp

import (
	"log"
	"os"
	"path/filepath"

	"github.com/gofs-cli/gofs/internal/lsp/diagnostic"
	"github.com/gofs-cli/gofs/internal/lsp/hover"
	"github.com/gofs-cli/gofs/internal/lsp/jsonrpc2"
	"github.com/gofs-cli/gofs/internal/lsp/protocol"
	"github.com/gofs-cli/gofs/internal/lsp/repo"
)

func Start() {
	// open log file in user's home directory
	dirname, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	logFile, err := os.OpenFile(filepath.Join(dirname, "gofs.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	r := repo.NewRepo()

	// lsp uses stdin/stdout for communication
	conn := jsonrpc2.NewConn(os.Stdin, os.Stdout)
	s, err := jsonrpc2.NewServer(conn, r.Open, protocol.ServerCapabilities{
		TextDocumentSync: protocol.TextDocumentSyncKindFull,
		HoverProvider:    true,
		DiagnosticProvider: protocol.DiagnosticOptions{
			Identifier:            "gofs",
			InterFileDependencies: true,
			WorkspaceDiagnostics:  false,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	// lifecycle handlers
	s.HandleLifecycle("initialize", jsonrpc2.Initialize(s))
	s.HandleLifecycle("initialized", jsonrpc2.Initialized(s))
	s.HandleLifecycle("shutdown", jsonrpc2.Shutdown(s))

	// repo handlers
	s.HandleRequest("textDocument/didOpen", repo.DidOpen(r))
	s.HandleRequest("textDocument/didClose", repo.DidClose(r))
	s.HandleRequest("textDocument/didChange", repo.DidChange(r))
	s.HandleRequest("textDocument/didSave", repo.DidSave(r))

	// language handlers
	s.HandleRequest("textDocument/hover", hover.Hover(r))
	s.HandleRequest("textDocument/diagnostic", diagnostic.Diagnostic(r))

	err = s.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
