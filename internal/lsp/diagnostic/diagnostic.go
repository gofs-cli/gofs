package diagnostic

import (
	"github.com/gofs-cli/gofs/internal/lsp/model"
	"github.com/gofs-cli/gofs/internal/lsp/protocol"
	"github.com/gofs-cli/gofs/internal/lsp/uri"
)

func UriDiagnostic(uri uri.Uri) []DiagnosticResponse {
	diagnostics := make([]DiagnosticResponse, 0)
	for _, d := range uri.Diag {
		message := ""
		switch d.Severity {
		case model.SeverityError:
			message = "Error: " + d.Message
		case model.SeverityWarning:
			message = "Warning: " + d.Message
		case model.SeverityInformation:
			message = "Info: " + d.Message
		case model.SeverityHint:
			message = "Hint: " + d.Message
		default:
			message = d.Message
		}
		diagnostics = append(diagnostics, DiagnosticResponse{
			Range: protocol.Range{
				Start: protocol.Position{
					Line:      uri.From.Line,
					Character: uri.From.Col,
				},
				End: protocol.Position{
					Line:      uri.To.Line,
					Character: uri.To.Col,
				},
			},
			Severity: d.Severity,
			Source:   "gofs",
			Message:  message,
		})
	}
	return diagnostics
}

// func UriDiagnostic(uri uri.Uri) DiagnosticResponse {

// 	message := ""
// 	severity := model.SeverityHint
// 	for _, d := range uri.Diag {
// 		if d.Severity < severity {
// 			severity = d.Severity
// 		}
// 		message += "-" + d.Message + "\n"
// 	}

// 	return DiagnosticResponse{
// 		Range: protocol.Range{
// 			Start: protocol.Position{
// 				Line:      uri.From.Line,
// 				Character: uri.From.Col,
// 			},
// 			End: protocol.Position{
// 				Line:      uri.To.Line,
// 				Character: uri.To.Col,
// 			},
// 		},
// 		Severity: severity,
// 		Source:   "gofs",
// 		Message:  message,
// 	}
// }
