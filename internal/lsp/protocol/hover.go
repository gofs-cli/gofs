package protocol

type HoverRequest struct {
	TextDocument TextDocument `json:"textDocument"`
	Position     Position     `json:"position"`
}

type HoverResponse struct {
	Contents string `json:"contents"`
}

type HoverResponseMarkup struct {
	Contents MarkupContent `json:"contents"`
}
