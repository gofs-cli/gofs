package uri

import (
	"fmt"
	"go/ast"
	"go/parser"
	"regexp"
	"slices"
	"strings"

	"github.com/gofs-cli/gofs/internal/lsp/model"
)

// Go router matching rules:
// "/" matches only the root
// "/{$}" matches anything starting from root (e.g., "/foo", "/bar/baz")
// "/foo" matches "/foo" exactly (no trailing slash)
// "/foo/{$}" matches "/foo/" and any subpath (e.g., "/foo/bar", "/foo/bar/baz")
// Go route matching uses longest path match first (most specific route wins)

var isValidUriChars = regexp.MustCompile(`^[A-Za-z0-9\-\_\.\~\{\}\*]+$`).MatchString

type Uri struct {
	Verb string
	Raw  string
	Seg  []string
	Diag []model.Diag
	From model.Pos
	To   model.Pos
}

func NewUri(verb, raw string) Uri {
	// get the segments and warnings from the pattern
	seg, diag := Segments(strings.TrimSpace(raw))

	return Uri{
		Verb: verb,
		Raw:  raw,
		Seg:  seg,
		Diag: diag,
	}
}

func NewUriFrom(verb, raw string, from model.Pos) Uri {
	u := NewUri(verb, raw)
	u.From = from
	return u
}

func NewUriFromTo(verb, raw string, from, to model.Pos) Uri {
	u := NewUriFrom(verb, raw, from)
	u.To = to
	return u
}

func Segments(pattern string) ([]string, []model.Diag) {
	expr, err := parser.ParseExpr(pattern)
	if err != nil {
		return nil, []model.Diag{
			{
				Severity: model.SeverityError,
				Message:  "invalid expression: " + err.Error(),
			},
		}
	}

	// traverse AST and extract the basic literals
	seg, diag := make([]string, 0), make([]model.Diag, 0)
	hasLit := false
	hasCall := false
	ast.Inspect(expr, func(n ast.Node) bool {
		return handleNodeType(n, &hasLit, &hasCall, &seg, &diag)
	})

	if hasLit && hasCall {
		diag = append(diag, model.Diag{
			Severity: model.SeverityWarning,
			Message:  "mixed literal and function call, combine into a single Sprintf",
		})
	}
	return seg, diag
}

func handleNodeType(n ast.Node,
	hasLit *bool,
	hasCall *bool,
	seg *[]string,
	diag *[]model.Diag,
) bool {
	// there should only be
	// - basic literals e.g "/foo"
	// - identifiers e.g. someVar
	// - selector expressions e.g. my.Var
	// - binary expressions e.g. "/foo" + "/bar"
	// - function calls e.g. fmt.Sprintf("/foo/%s", someVar)
	// - wildcard e.g. "{$}"

	switch x := n.(type) {
	case *ast.BasicLit:
		*hasLit = true
		s, d := LiteralSegments(x.Value)
		*seg = append(*seg, s...)
		*diag = append(*diag, d...)
		// nothing to recurse
		return false
	case *ast.Ident:
		// add the variable
		*seg = append(*seg, "{}")
		// nothing to recurse
		return false
	case *ast.SelectorExpr:
		// add the variable
		*seg = append(*seg, "{}")
		// nothing to recurse
		return false
	case *ast.BinaryExpr:
		// traverse the binary expression
		return true
	case *ast.CallExpr:
		*hasCall = true
		call := make([]string, 0)
		ast.Inspect(x.Fun, func(n ast.Node) bool {
			if y, ok := n.(*ast.SelectorExpr); ok {
				call = append(call, y.Sel.Name)
			}
			return true
		})
		if !slices.Contains(call, "Sprintf") {
			*diag = append(*diag, model.Diag{
				Severity: model.SeverityWarning,
				Message:  fmt.Sprintf("unexpected function call %s, use Sprintf instead", strings.Join(call, ".")),
			})
			return false
		}
		for _, arg := range x.Args {
			if fArg, ok := arg.(*ast.BasicLit); ok {
				s, d := LiteralSegments(fArg.Value)
				*seg = append(*seg, s...)
				*diag = append(*diag, d...)
			}
		}
		return false
	case nil:
		return false
	}

	*diag = append(*diag, model.Diag{
		Severity: model.SeverityError,
		Message:  fmt.Sprintf("unexpected code: %T", n),
	})
	return false
}

func LiteralSegments(pattern string) ([]string, []model.Diag) {
	// trim spaces and double quotes
	trimmed := strings.Trim(strings.TrimSpace(pattern), `"`)
	// remove leading slash
	trimmed = strings.Trim(trimmed, "/")

	parts := strings.Split(trimmed, "/")
	seg, diag := make([]string, 0), make([]model.Diag, 0)
	for i, p := range parts {
		if p == "{$}" {
			if i != 0 {
				diag = append(diag, model.Diag{
					Severity: model.SeverityError,
					Message:  fmt.Sprintf("invalid route pattern %s: {$} is only allowed at the root path", p),
				})
			}
			seg = append(seg, "{$}")
			continue
		}
		if strings.HasPrefix(p, "%") || // variable in Sprintf
			strings.HasPrefix(p, "*") || // wildcard in route
			strings.HasPrefix(p, "{") { // variable in route
			// if the segment is a variable, add it as a variable
			seg = append(seg, "{}")
		} else {
			// else add it as is with warnings
			seg = append(seg, p)
			if p != "" && !isValidUriChars(p) {
				diag = append(diag, model.Diag{
					Severity: model.SeverityError,
					Message:  fmt.Sprintf("invalid character in uri segment %s", p),
				})
			}
		}
	}
	return seg, diag
}

// *** Recommended patterns ***
//
// string literals - ideal because they are clear and concise
// hx-get="/foo/bar"
//
// string literals and variables - developer can easily identify the route
// hx-get={ "/foo/" + bar }
// hx-get={ bar + "/foo/" }
// hx-get={ "/foo/" + bar + "/foo" + ... }
//
// function call to fmt.Sprintf - developer can easily identify the route
// hx-get={ fmt.Sprintf("/foo/%s", bar) }
//
// variables - allowed for generic components, not recommended for other cases
//           - route cannot be identified and any route is possible
// hx-get={ foo }

func (u *Uri) MatchLevel(uri Uri) int {
	if uri.Verb != u.Verb || len(uri.Seg) == 0{
		// verb does not match
		// no segments to match
		return model.NoMatch
	}

	// if there are different number of segments in the pattern than the uri, it cannot match
	// i.e. /foo/bar cannot match /foo
	if len(uri.Seg) != len(u.Seg) {
		return model.NoMatch
	}

	// "/{$}" catch-all route starting from root
	if len(uri.Seg) == 1 && uri.Seg[0] == "{$}" {
		return model.WildcardMatch
	}

	// Exact match
	if slices.Equal(u.Seg, uri.Seg) {
		return model.ExactMatch
	}

	hasVar := false
	hasWildcard := false
	for i, s := range uri.Seg {
		if s == "{$}" {
			//"{$}" is only allowed at the root path
			return model.NoMatch
		}

		if s == "{}" || u.Seg[i] == "{}" {
			hasVar = true
			continue
		}
		if u.Seg[i] == "*" {
			hasWildcard = true
			continue
		}
		if u.Seg[i] != s {
			return model.NoMatch
		}
	}

	switch {
	case hasWildcard:
		return model.WildcardMatch
	case hasVar:
		return model.VariableMatch
	default:
		return model.ExactMatch
	}

	// // segments must match allowing for variables
	// for i, s := range uri.Seg {
	// 	if s == "{}" || u.Seg[i] == "{}" {
	// 		continue
	// 	}
	// 	if u.Seg[i] == "*" {
	// 		// route has a wildcard so it can match any segment
	// 		continue
	// 	}

	// 	// "{$}" is only allowed at the root path
	// 	if s == "{$}" {
	// 		return false
	// 	}

	// 	if u.Seg[i] != s {
	// 		return false
	// 	}
	// }

	// return true
}
