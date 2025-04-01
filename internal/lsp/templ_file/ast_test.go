package templFile

import (
	"reflect"
	"testing"

	"github.com/gofs-cli/gofs/internal/lsp/model"
	"github.com/gofs-cli/gofs/internal/lsp/uri"
)

func TestGetTemplUris(t *testing.T) {
	t.Parallel()
	t.Run("GetTemplUris", func(t *testing.T) {
		src := `package stats

import (
	"fmt"
	"github.com/atos-digital/EFL-Live-Picker/internal/app/season"
	"github.com/atos-digital/EFL-Live-Picker/internal/app/user"
	"github.com/atos-digital/EFL-Live-Picker/internal/controllers"
	"github.com/atos-digital/EFL-Live-Picker/ui/components/navigation"
)

templ varLink(text, img, link string) {
	<a hx-get={ link }>
		{ text }
	</a>
}

templ literalLink() {
	<a hx-post="/link">
		Literal Link
	</a>
}

`
		uris, err := GetTemplUris(src)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		expect := []uri.Uri{
			uri.NewUriFromTo(
				"GET",
				"link",
				model.Pos{Line: 11, Col: 13},
				model.Pos{Line: 11, Col: 17},
			),
			uri.NewUriFromTo(
				"POST",
				`"/link"`,
				model.Pos{Line: 17, Col: 4},
				model.Pos{Line: 17, Col: 11},
			),
		}
		if !reflect.DeepEqual(uris, expect) {
			t.Errorf("expected %v, got %v", expect, uris)
		}
	})
}
