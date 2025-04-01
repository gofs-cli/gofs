package templFile

import (
	templ "github.com/a-h/templ/parser/v2"

	"github.com/gofs-cli/gofs/internal/lsp/model"
	"github.com/gofs-cli/gofs/internal/lsp/uri"
)

func GetTemplUris(src string) ([]uri.Uri, error) {
	t, err := templ.ParseString(src)
	if err != nil {
		return nil, err
	}

	uris := make([]uri.Uri, 0)

	for _, n := range t.Nodes {
		switch n := n.(type) {
		case templ.HTMLTemplate:
			np := NodeParser{
				ElementFunc: func(n templ.Element) error {
					if hx := ElementAttribute(n.Attributes, "hx-get"); hx != nil {
						uris = append(uris, uri.NewUriFromTo(
							"GET",
							hx.Value,
							model.Pos{
								Line: int(hx.NameRange.From.Line),
								Col:  int(hx.NameRange.From.Col),
							},
							model.Pos{
								Line: int(hx.NameRange.To.Line),
								Col:  int(hx.NameRange.To.Col),
							},
						))
					}
					if hx := ElementAttribute(n.Attributes, "hx-post"); hx != nil {
						uris = append(uris, uri.NewUriFromTo(
							"POST",
							hx.Value,
							model.Pos{
								Line: int(hx.NameRange.From.Line),
								Col:  int(hx.NameRange.From.Col),
							},
							model.Pos{
								Line: int(hx.NameRange.To.Line),
								Col:  int(hx.NameRange.To.Col),
							},
						))
					}
					if hx := ElementAttribute(n.Attributes, "hx-put"); hx != nil {
						uris = append(uris, uri.NewUriFromTo(
							"PUT",
							hx.Value,
							model.Pos{
								Line: int(hx.NameRange.From.Line),
								Col:  int(hx.NameRange.From.Col),
							},
							model.Pos{
								Line: int(hx.NameRange.To.Line),
								Col:  int(hx.NameRange.To.Col),
							},
						))
					}
					if hx := ElementAttribute(n.Attributes, "hx-delete"); hx != nil {
						uris = append(uris, uri.NewUriFromTo(
							"DELETE",
							hx.Value,
							model.Pos{
								Line: int(hx.NameRange.From.Line),
								Col:  int(hx.NameRange.From.Col),
							},
							model.Pos{
								Line: int(hx.NameRange.To.Line),
								Col:  int(hx.NameRange.To.Col),
							},
						))
					}
					return nil
				},
			}
			for _, child := range n.Children {
				np.RecurseNode(child)
			}
		}
	}

	return uris, nil
}
