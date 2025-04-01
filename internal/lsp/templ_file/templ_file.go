package templFile

import "github.com/gofs-cli/gofs/internal/lsp/uri"

type TemplFile struct {
	Path           string
	Text           string
	Uris           []uri.Uri
	UrisRouteIndex []int
}
