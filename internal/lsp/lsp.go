package lsp

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/gofs-cli/gofs/internal/lsp/diagnostic"
	"github.com/gofs-cli/gofs/internal/lsp/hover"
	"github.com/gofs-cli/gofs/internal/lsp/jsonrpc2"
	"github.com/gofs-cli/gofs/internal/lsp/protocol"
	"github.com/gofs-cli/gofs/internal/lsp/repo"
)

func Start(debug bool) {
	// InitLogger sets up slog to log into ~/.gofs/debug.log if debug == true.
	// Otherwise it silences the logger.
	// TODO: clarify the debug mode
	initLogger(debug)

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
		slog.Error("fatal error", "err", err)
		os.Exit(1)
	}

	// lifecycle handlers
	s.HandleLifecycle("initialize", jsonrpc2.Initialize(s))
	s.HandleLifecycle("initialized", jsonrpc2.Initialized(s))
	s.HandleLifecycle("shutdown", jsonrpc2.Shutdown(s))

	// repo handlers
	s.HandleRequest("textDocument/didOpen", repo.DidOpen(r), jsonrpc2.DecodeParams[repo.DidOpenRequest]())
	s.HandleRequest("textDocument/didClose", repo.DidClose(r), jsonrpc2.DecodeParams[repo.DidCloseRequest]())
	s.HandleRequest("textDocument/didChange", repo.DidChange(r), jsonrpc2.DecodeParams[repo.DidChangeRequest]())
	s.HandleRequest("textDocument/didSave", repo.DidSave(r), nil)

	// language handlers
	s.HandleRequest("textDocument/hover", hover.Hover(r), jsonrpc2.DecodeParams[hover.HoverRequest]())
	s.HandleRequest("textDocument/diagnostic", diagnostic.Diagnostic(r), jsonrpc2.DecodeParams[diagnostic.DiagnosticRequest]())

	err = s.ListenAndServe()
	if err != nil {
		slog.Error("fatal error", "err", err)
		os.Exit(1)
	}
}

func initLogger(debug bool) {
	if !debug {
		slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
		return
	}

	home, err := os.UserHomeDir()
	if err != nil {
		panic("failed to get user home dir: " + err.Error())
	}

	logDir := filepath.Join(home, ".gofs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		panic("failed to create log dir: " + err.Error())
	}

	logPath := filepath.Join(logDir, "debug.log")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		panic("failed to open log file: " + err.Error())
	}

	handler := slog.NewTextHandler(f, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	slog.SetDefault(slog.New(handler))
}
