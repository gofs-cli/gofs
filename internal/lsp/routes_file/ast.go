package routesFile

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log/slog"
	"path"
	"strings"

	"github.com/gofs-cli/gofs/internal/lsp/model"
	"github.com/gofs-cli/gofs/internal/lsp/uri"
)

func getImports(raw []byte) map[string]string {
	imports := make(map[string]string)
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", raw, parser.AllErrors)
	if err != nil {
		slog.Error("error parsing routes.go file: ", "err", err)
		return imports
	}

	ast.Inspect(file, func(n ast.Node) bool {
		if imp, ok := n.(*ast.ImportSpec); ok {
			v := strings.Trim(imp.Path.Value, `"`)
			if imp.Name != nil {
				imports[imp.Name.Name] = v
			} else {
				p := path.Base(v)
				imports[p] = v
			}
			return false
		}
		return true
	})
	return imports
}

func getRoutes(raw []byte) []Route {
	imports := getImports(raw)
	var routes []Route

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", raw, parser.ParseComments)
	if err != nil {
		slog.Error("error parsing routes.go file: ", "err", err)
		return routes
	}

	ast.Inspect(file, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			// Check if the function being called is a selector expression (i.e., a method call like s.Method())
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				// Check if the method being called is the routing "Handle" or "HandleFunc"
				// and tje call and has 2 arguments viz. route and handler
				// TODO HandleFunc
				if (sel.Sel.Name == "Handle") && len(call.Args) == 2 {
					// Extract the verb, route, and handler from the arguments
					var verb, pkg string
					var u uri.Uri

					// The route must be a basic literal (e.g., "/foo/bar")
					if routeLit, ok := call.Args[0].(*ast.BasicLit); ok {
						// strip the quotes from the whole route
						parts := strings.SplitN(routeLit.Value[1:len(routeLit.Value)-1], " ", 2)
						if len(parts) != 2 {
							return false
						}
						verb = parts[0]
						// add the quotes to the uri part of the route
						u = uri.NewUriFromTo(verb, `"`+parts[1]+`"`,
							model.Pos{
								Line: fset.Position(call.Args[0].Pos()).Line - 1,
								Col:  fset.Position(call.Args[0].Pos()).Column - 1,
							},
							model.Pos{
								Line: fset.Position(call.Args[0].End()).Line - 1,
								Col:  fset.Position(call.Args[0].End()).Column - 1,
							},
						)
					}

					handler := Handler{
						From: model.Pos{
							Line: fset.Position(call.Args[1].Pos()).Line - 1,
							Col:  fset.Position(call.Args[1].Pos()).Column - 1,
						},
						To: model.Pos{
							Line: fset.Position(call.Args[1].End()).Line - 1,
							Col:  fset.Position(call.Args[1].End()).Column - 1,
						},
					}
					// The handler is either a selector + identifier (e.g., post.Handler(assets))
					// where the selector is the package name
					// the handler is an identifier (e.g., Handler(assets))
					switch handlerAst := call.Args[1].(*ast.CallExpr).Fun.(type) {
					case *ast.Ident:
						handler.Call = handlerAst.Name
					case *ast.SelectorExpr:
						pkg = handlerAst.X.(*ast.Ident).Name
						handler.Call = handlerAst.Sel.Name
					}

					routes = append(routes, Route{
						Uri:     u,
						Handler: handler,
						Pkg:     imports[pkg],
					})
				}
			}
		}
		return true
	})

	return routes
}
