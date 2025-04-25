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

	// repo handlers
	s.HandleRequest("textDocument/didOpen", repo.DidOpen(r))
	s.HandleRequest("textDocument/didClose", repo.DidClose(r))
	s.HandleRequest("textDocument/didChange", repo.DidChange(r))
	s.HandleRequest("textDocument/didSave", repo.DidSave(r))

	// hover handlers
	s.HandleRequest("textDocument/hover", hover.Hover(r))

	// diagnostic handlers
	s.HandleRequest("textDocument/diagnostic", diagnostic.Diagnostic(r))

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
