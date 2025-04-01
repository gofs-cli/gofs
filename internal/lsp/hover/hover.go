package hover

import (
	templFile "github.com/gofs-cli/gofs/internal/lsp/templ_file"
)

func HoveredUri(t templFile.TemplFile, line, col int) int {
	for i, u := range t.Uris {
		if u.From.Line <= line && u.From.Col <= col && u.To.Line >= line && u.To.Col >= col {
			return i
		}
	}
	return -1
}
