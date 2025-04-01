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

// Go router matching rules
// / matches root
// /{$} matches any after
// /foo matches /foo exactly
// /foo/{$} matches /foo/ with trailing slash and any after
//
// Go route matches longest path first

var isValidUriChars = regexp.MustCompile(`^[A-Za-z\-\_\.\~\{\}\*]+$`).MatchString

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
		// there should only be
		// - basic literals e.g "/foo"
		// - identifiers e.g. someVar
		// - selector expressions e.g. my.Var
		// - binary expressions e.g. "/foo" + "/bar"
		// - function calls e.g. fmt.Sprintf("/foo/%s", someVar)
		switch x := n.(type) {
		case *ast.BasicLit:
			hasLit = true
			s, d := LiteralSegments(x.Value)
			seg = append(seg, s...)
			diag = append(diag, d...)
			// nothing to recurse
			return false
		case *ast.Ident:
			// add the variable
			seg = append(seg, "{}")
			// nothing to recurse
			return false
		case *ast.SelectorExpr:
			// add the variable
			seg = append(seg, "{}")
			// nothing to recurse
			return false
		case *ast.BinaryExpr:
			// traverse the binary expression
			return true
		case *ast.CallExpr:
			hasCall = true
			call := make([]string, 0)
			ast.Inspect(x.Fun, func(n ast.Node) bool {
				if y, ok := n.(*ast.SelectorExpr); ok {
					call = append(call, y.Sel.Name)
				}
				return true
			})
			if !slices.Contains(call, "Sprintf") {
				diag = append(diag, model.Diag{
					Severity: model.SeverityWarning,
					Message:  fmt.Sprintf("unexpected function call %s, use Sprintf instead", strings.Join(call, ".")),
				})
				return false
			}
			for _, arg := range x.Args {
				if fArg, ok := arg.(*ast.BasicLit); ok {
					s, d := LiteralSegments(fArg.Value)
					seg = append(seg, s...)
					diag = append(diag, d...)
				}
			}
			return false
		case nil:
			return false
		}
		// we should not encounter any other types, so do nothing
		diag = append(diag, model.Diag{
			Severity: model.SeverityError,
			Message:  fmt.Sprintf("unexpected code: %T", n),
		})
		return false
	})

	if hasLit && hasCall {
		diag = append(diag, model.Diag{
			Severity: model.SeverityWarning,
			Message:  "mixed literal and function call, combine into a single Sprintf",
		})
	}
	return seg, diag
}

func LiteralSegments(pattern string) ([]string, []model.Diag) {
	// trim spaces and double quotes
	trimmed := strings.Trim(strings.TrimSpace(pattern), `"`)
	// remove leading slash
	trimmed = strings.Trim(trimmed, "/")

	parts := strings.Split(trimmed, "/")
	seg, diag := make([]string, 0), make([]model.Diag, 0)
	for _, p := range parts {
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

func (u *Uri) IsMatch(uri Uri) bool {
	if uri.Verb != u.Verb {
		// verb does not match
		return false
	}

	if len(uri.Seg) == 0 {
		// no segments to match
		return false
	}

	// if there are different number of segments in the pattern than the uri, it cannot match
	// i.e. /foo/bar cannot match /foo
	if len(uri.Seg) != len(u.Seg) {
		return false
	}

	// segments must match allowing for variables
	for i, s := range uri.Seg {
		if s == "{}" || u.Seg[i] == "{}" {
			continue
		}
		if u.Seg[i] == "*" {
			// route has a wildcard so it can match any segment
			continue
		}
		if u.Seg[i] != s {
			return false
		}
	}

	return true
}
