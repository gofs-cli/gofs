package protocol

// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/

type TextDocument struct {
	Path string `json:"uri"`
	Text string `json:"text"`
}

type MarkupContent struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}
