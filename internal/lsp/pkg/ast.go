package pkg

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log/slog"
	"strings"

	"golang.org/x/tools/go/packages"

	"github.com/gofs-cli/gofs/internal/lsp/model"
)

func GetPkg(name string) (*Pkg, error) {
	cfg := &packages.Config{Mode: packages.NeedFiles}
	p, err := packages.Load(cfg, name)
	if err != nil {
		return nil, err
	}
	f := make([]Func, 0)
	for _, g := range p[0].GoFiles {
		// don't parse templ generated files
		if strings.HasSuffix(g, "_templ.go") {
			continue
		}

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, g, nil, parser.AllErrors)
		if err != nil {
			slog.Error("error parsing file", "file", g, "err", err)
			continue
		}

		ast.Inspect(file, func(n ast.Node) bool {
			if fd, ok := n.(*ast.FuncDecl); ok {
				f = append(f, Func{
					Name: fd.Name.Name,
					File: g,
					Pos: model.Pos{
						Line: fset.Position(fd.Pos()).Line - 1,   // model.Pos is zero base
						Col:  fset.Position(fd.Pos()).Column - 1, // model.Pos is zero base
					},
				})
				return false
			}
			return true
		})
	}

	return &Pkg{
		Files: p[0].GoFiles,
		Funcs: f,
	}, nil
}
