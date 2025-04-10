package protocol

// didOpen
type DidOpenRequest struct {
	TextDocument TextDocument `json:"textDocument"`
}

// didChange
type DidChangeRequest struct {
	TextDocument   TextDocument    `json:"textDocument"`
	ContentChanges []ContentChange `json:"contentChanges"`
}

type ContentChange struct {
	Text string `json:"text"`
}

// didClose
type DidCloseRequest struct {
	TextDocument TextDocument `json:"textDocument"`
}
