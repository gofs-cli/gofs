package hover

import (
	"log"

	templFile "github.com/gofs-cli/gofs/internal/lsp/templ_file"
)

func HoveredUri(t templFile.TemplFile, line, col int) int {
	log.Printf("the length of hovered templ uris: %d", len(t.Uris))
	for i, r := range t.Uris {
		log.Printf("current uri index: %d  uri: %s", i, r.Raw)
	}

	for i, u := range t.Uris {
		if u.From.Line <= line && u.From.Col <= col && u.To.Line >= line && u.To.Col >= col {
			return i
		}
	}
	return -1
}
