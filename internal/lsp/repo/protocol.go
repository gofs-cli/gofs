package repo

import "github.com/gofs-cli/gofs/internal/lsp/protocol"

// didOpen
type DidOpenRequest struct {
	TextDocument protocol.TextDocument `json:"textDocument"`
}

// didChange
type DidChangeRequest struct {
	TextDocument   protocol.TextDocument `json:"textDocument"`
	ContentChanges []ContentChange       `json:"contentChanges"`
}

type ContentChange struct {
	Text string `json:"text"`
}

// didClose
type DidCloseRequest struct {
	TextDocument protocol.TextDocument `json:"textDocument"`
}
