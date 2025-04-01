package routesFile

import (
	"reflect"
	"testing"

	"github.com/gofs-cli/gofs/internal/lsp/model"
	"github.com/gofs-cli/gofs/internal/lsp/uri"
)

func testFile() []byte {
	return []byte(`package main

import (
	"strings"

	"github.com/org/project/post"
	d "github.com/org/project/delete"
)

func (s *Server) Routes() error {
	s.middleware(s.r)
	s.r.Handle("GET /assets/*", s.handleAssets(assets))
	s.r.Handle("POST /post", post.PostHandler(s.db))
	s.r.Handle("DELETE /delete", d.DeleteHandler(s.db))
	s.r.Handle("GET /img/*", handleImg())
}

`)
}

func TestGetImports(t *testing.T) {
	t.Parallel()
	t.Run("GetImports", func(t *testing.T) {
		imports := getImports(testFile())
		if len(imports) != 3 {
			t.Fatalf("expected 2 imports, got %d", len(imports))
		}
		if v, ok := imports["post"]; !ok || v != "github.com/org/project/post" {
			t.Error("expected github.com/org/project/post, got ", v)
		}
		if v, ok := imports["d"]; !ok || v != "github.com/org/project/delete" {
			t.Error("expected github.com/org/project/delete, got ", v)
		}
	})
}

func TestGetRoutes(t *testing.T) {
	t.Parallel()
	t.Run("GetRoutes", func(t *testing.T) {
		routes := getRoutes(testFile())
		expect := []Route{
			{
				Uri: uri.Uri{
					Verb: "GET",
					Raw:  `"/assets/*"`,
					Seg:  []string{"assets", "{}"},
					Diag: []model.Diag{},
					From: model.Pos{Line: 11, Col: 12},
					To:   model.Pos{Line: 11, Col: 27},
				},
				Handler: Handler{
					Call: "handleAssets",
					From: model.Pos{Line: 11, Col: 29},
					To:   model.Pos{Line: 11, Col: 51},
				},
				Pkg: "", // cannot match package because its local
			},
			{
				Uri: uri.Uri{
					Verb: "POST",
					Raw:  `"/post"`,
					Seg:  []string{"post"},
					Diag: []model.Diag{},
					From: model.Pos{Line: 12, Col: 12},
					To:   model.Pos{Line: 12, Col: 24},
				},
				Handler: Handler{
					Call: "PostHandler",
					From: model.Pos{Line: 12, Col: 26},
					To:   model.Pos{Line: 12, Col: 48},
				},
				Pkg: "github.com/org/project/post",
			},
			{
				Uri: uri.Uri{
					Verb: "DELETE",
					Raw:  `"/delete"`,
					Seg:  []string{"delete"},
					Diag: []model.Diag{},
					From: model.Pos{Line: 13, Col: 12},
					To:   model.Pos{Line: 13, Col: 28},
				},
				Handler: Handler{
					Call: "DeleteHandler",
					From: model.Pos{Line: 13, Col: 30},
					To:   model.Pos{Line: 13, Col: 51},
				},
				Pkg: "github.com/org/project/delete",
			},
			{
				Uri: uri.Uri{
					Verb: "GET",
					Raw:  `"/img/*"`,
					Seg:  []string{"img", "{}"},
					Diag: []model.Diag{},
					From: model.Pos{Line: 14, Col: 12},
					To:   model.Pos{Line: 14, Col: 24},
				},
				Handler: Handler{
					Call: "handleImg",
					From: model.Pos{Line: 14, Col: 26},
					To:   model.Pos{Line: 14, Col: 37},
				},
				Pkg: "", // cannot match package because its local
			},
		}
		if len(routes) != len(expect) {
			t.Fatalf("expected routes to match")
		}
		for i := range routes {
			if !reflect.DeepEqual(routes[i], expect[i]) {
				t.Fatalf("expected:\n%v\ngot:\n%v", expect[i], routes[i])
			}
		}
	})
}
