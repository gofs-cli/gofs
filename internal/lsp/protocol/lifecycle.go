package protocol

// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/

// Initialize request
type InitializeRequest struct {
	Capabilities ClientCapabilities `json:"capabilities"`
	RootPath     string             `json:"rootPath"`
}

type ClientCapabilities struct{}

// Initialize response
type InitializeResponse struct {
	Capabilities ServerCapabilities `json:"capabilities"`
}

type ServerCapabilities struct {
	TextDocumentSync   TextDocumentSyncKind `json:"textDocumentSync"`
	HoverProvider      bool                 `json:"hoverProvider"`
	DiagnosticProvider DiagnosticOptions    `json:"diagnosticProvider"`
}

type DiagnosticOptions struct {
	Identifier            string `json:"identifier"`
	InterFileDependencies bool   `json:"interFileDependencies"`
	WorkspaceDiagnostics  bool   `json:"workspaceDiagnostics"`
}

type TextDocumentSyncKind int

const (
	TextDocumentSyncKindNone = iota
	TextDocumentSyncKindFull
	TextDocumentSyncKindIncremental
)
