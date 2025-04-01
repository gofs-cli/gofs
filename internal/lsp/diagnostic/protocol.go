package diagnostic

import "github.com/gofs-cli/gofs/internal/lsp/protocol"

type DiagnosticRequest struct {
	TextDocument protocol.TextDocument `json:"textDocument"`
}

type FullDiagnosticResponse struct {
	Kind  string               `json:"kind"`
	Items []DiagnosticResponse `json:"items"`
}

type PublishDiagnosticsParams struct {
	DocumentUri string               `json:"uri"`
	Diagnostics []DiagnosticResponse `json:"diagnostics"`
}

type DiagnosticResponse struct {
	Range           protocol.Range   `json:"range"`
	Severity        int              `json:"severity"`
	Code            int              `json:"code"`
	CodeDescription *CodeDescription `json:"codeDescription,omitempty"`
	Source          string           `json:"source"`
	Message         string           `json:"message"`
}

type CodeDescription struct {
	Href string `json:"href"`
}

const (
	SeverityError       = 1
	SeverityWarning     = 2
	SeverityInformation = 3
	SeverityHint        = 4
)

const (
	KindFull      = "full"
	KindUnchanged = "unchanged"
)
