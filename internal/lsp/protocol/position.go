package protocol

// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/

type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}
