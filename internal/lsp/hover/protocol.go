package hover

import "github.com/gofs-cli/gofs/internal/lsp/protocol"

type HoverRequest struct {
	TextDocument protocol.TextDocument `json:"textDocument"`
	Position     protocol.Position     `json:"position"`
}

type HoverResponse struct {
	Contents string `json:"contents"`
}

type HoverResponseMarkup struct {
	Contents protocol.MarkupContent `json:"contents"`
}
