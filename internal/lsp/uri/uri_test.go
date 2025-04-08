package uri

import (
	"reflect"
	"testing"

	"github.com/gofs-cli/gofs/internal/lsp/model"
)

func TestNewUri(t *testing.T) {
	t.Parallel()

	t.Run("only suffix wildcard", func(t *testing.T) {
		u := NewUri("GET", `"/"`)
		expected := Uri{
			Verb: "GET",
			Raw:  `"/"`,
			Seg:  []string{""},
			Diag: []model.Diag{},
		}
		if !reflect.DeepEqual(u, expected) {
			t.Errorf("expected:\n%v\ngot:\n%v", expected, u)
		}
	})

	t.Run("literal path with suffix wildcard", func(t *testing.T) {
		u := NewUri("GET", `"/{$}"`)
		expected := Uri{
			Verb: "GET",
			Raw:  `"/{$}"`,
			Seg:  []string{"{$}"},
			Diag: []model.Diag{},
		}
		if !reflect.DeepEqual(u, expected) {
			t.Errorf("expected:\n%v\ngot:\n%v", expected, u)
		}
	})

	t.Run("string literal", func(t *testing.T) {
		u := NewUri("GET", `"/foo/bar"`)
		expected := Uri{
			Verb: "GET",
			Raw:  `"/foo/bar"`,
			Seg:  []string{"foo", "bar"},
			Diag: []model.Diag{},
		}
		if !reflect.DeepEqual(u, expected) {
			t.Errorf("expected:\n%v\ngot:\n%v", expected, u)
		}
	})

	t.Run("path params", func(t *testing.T) {
		u := NewUri("GET", `"/foo/{bar}/foo"`)
		expected := Uri{
			Verb: "GET",
			Raw:  `"/foo/{bar}/foo"`,
			Seg:  []string{"foo", "{}", "foo"},
			Diag: []model.Diag{},
		}
		if !reflect.DeepEqual(u, expected) {
			t.Errorf("expected:\n%v\ngot:\n%v", expected, u)
		}
	})

	t.Run("trims space", func(t *testing.T) {
		u := NewUri("GET", ` "/foo/{bar}/foo" `)
		expected := Uri{
			Verb: "GET",
			Raw:  ` "/foo/{bar}/foo" `,
			Seg:  []string{"foo", "{}", "foo"},
			Diag: []model.Diag{},
		}
		if !reflect.DeepEqual(u, expected) {
			t.Errorf("expected:\n%v\ngot:\n%v", expected, u)
		}
	})

	t.Run("relative uri", func(t *testing.T) {
		u := NewUri("GET", `"foo/bar"`)
		expected := Uri{
			Verb: "GET",
			Raw:  `"foo/bar"`,
			Seg:  []string{"foo", "bar"},
			Diag: []model.Diag{},
		}
		if !reflect.DeepEqual(u, expected) {
			t.Errorf("expected:\n%v\ngot:\n%v", expected, u)
		}
	})
}

func TestDiags(t *testing.T) {
	t.Parallel()

	t.Run("invalid route pattern", func(t *testing.T) {
		u := NewUri("GET", `"/foo/{$}"`)
		expected := []model.Diag{
			{
				Severity: model.SeverityError,
				Message:  "invalid route pattern {$}: {$} is only allowed at the root path",
			},
		}
		if !reflect.DeepEqual(u.Diag, expected) {
			t.Errorf("expected:\n%v\ngot:\n%v", expected, u)
		}
	})

	t.Run("no warnings", func(t *testing.T) {
		u := NewUri("GET", `"/foo/bar"`)
		expected := []model.Diag{}
		if !reflect.DeepEqual(u.Diag, expected) {
			t.Errorf("expected:\n%v\ngot:\n%v", expected, u)
		}
	})

	t.Run("invalid chars in literal", func(t *testing.T) {
		u := NewUri("GET", `"/foo/b ar"`)
		expected := []model.Diag{
			{
				Severity: model.SeverityError,
				Message:  "invalid character in uri segment b ar",
			},
		}
		if !reflect.DeepEqual(u.Diag, expected) {
			t.Errorf("expected:\n%v\ngot:\n%v", expected, u)
		}
	})

	t.Run("invalid chars in call", func(t *testing.T) {
		u := NewUri("GET", `fmt.Sprintf("/foo/b ar")`)
		expected := []model.Diag{
			{
				Severity: model.SeverityError,
				Message:  "invalid character in uri segment b ar",
			},
		}
		if !reflect.DeepEqual(u.Diag, expected) {
			t.Errorf("expected:\n%v\ngot:\n%v", expected, u)
		}
	})
}

func TestSegments(t *testing.T) {
	t.Parallel()

	t.Run("root literal", func(t *testing.T) {
		u, _ := Segments(`"/"`)
		expected := []string{""}

		if !reflect.DeepEqual(u, expected) {
			t.Errorf("expected:\n%v\ngot:\n%v", expected, u)
		}
	})

	t.Run("suffix wildcard", func(t *testing.T) {
		u, _ := Segments(`"/{$}"`)
		expected := []string{"{$}"}

		if !reflect.DeepEqual(u, expected) {
			t.Errorf("expected:\n%v\ngot:\n%v", expected, u)
		}
	})

	t.Run("literal", func(t *testing.T) {
		u, _ := Segments(`"/foo/bar"`)
		expected := []string{"foo", "bar"}

		if !reflect.DeepEqual(u, expected) {
			t.Errorf("expected:\n%v\ngot:\n%v", expected, u)
		}
	})

	t.Run("literals with binary op", func(t *testing.T) {
		u, _ := Segments(`"/foo" + "/bar"`)
		expected := []string{"foo", "bar"}

		if !reflect.DeepEqual(u, expected) {
			t.Errorf("expected:\n%v\ngot:\n%v", expected, u)
		}
	})

	t.Run("literals and variables", func(t *testing.T) {
		s, _ := Segments(`"/foo" + someVar + "/bar"`)
		expected := []string{"foo", "{}", "bar"}

		if !reflect.DeepEqual(s, expected) {
			t.Errorf("expected:\n%v\ngot:\n%v", expected, s)
		}
	})

	t.Run("call", func(t *testing.T) {
		s, _ := Segments(`fmt.Sprintf("/foo/bar")`)
		expected := []string{"foo", "bar"}

		if !reflect.DeepEqual(s, expected) {
			t.Errorf("expected:\n%v\ngot:\n%v", expected, s)
		}
	})

	t.Run("call with vars", func(t *testing.T) {
		s, _ := Segments(`fmt.Sprintf("/foo/%s/bar", someVar)`)
		expected := []string{"foo", "{}", "bar"}

		if !reflect.DeepEqual(s, expected) {
			t.Errorf("expected:\n%v\ngot:\n%v", expected, s)
		}
	})

	t.Run("call mixed literal", func(t *testing.T) {
		s, d := Segments(`fmt.Sprintf("/foo/%s/bar", someVar) + "/foobar"`)
		expected := []string{"foo", "{}", "bar", "foobar"}

		if !reflect.DeepEqual(s, expected) {
			t.Errorf("expected:\n%v\ngot:\n%v", expected, s)
		}
		if len(d) != 1 {
			t.Fatal("expected one diagnostic")
		}
		if !reflect.DeepEqual(d[0], model.Diag{
			Severity: model.SeverityWarning,
			Message:  "mixed literal and function call, combine into a single Sprintf",
		}) {
			t.Errorf("expected diagnostic, got: %v", d[0])
		}
	})

	t.Run("standalone var", func(t *testing.T) {
		s, _ := Segments("someVar")
		expected := []string{"{}"}

		if !reflect.DeepEqual(s, expected) {
			t.Errorf("expected:\n%v\ngot:\n%v", expected, s)
		}
	})

	t.Run("selector var", func(t *testing.T) {
		s, _ := Segments(`"/grid/edit/modal/" + c.Fixture.ID`)
		expected := []string{"grid", "edit", "modal", "{}"}

		if !reflect.DeepEqual(s, expected) {
			t.Errorf("expected:\n%v\ngot:\n%v", expected, s)
		}
	})

	t.Run("invalid returns nil", func(t *testing.T) {
		s, d := Segments(`"invalid`)

		if s != nil {
			t.Errorf("expected nil, got: %v", s)
		}
		if len(d) != 1 {
			t.Fatal("expected one diagnostic")
		}
		if !reflect.DeepEqual(d[0], model.Diag{
			Severity: model.SeverityError,
			Message:  "invalid expression: 1:1: string literal not terminated",
		}) {
			t.Errorf("expected diagnostic, got: %v", d[0])
		}
	})
}

func TestIsMatch(t *testing.T) {
	t.Parallel()

	t.Run("root matches", func(t *testing.T) {
		u := NewUri("GET", `"/"`)
		if match := u.IsMatch(NewUri("GET", `"/"`)); !match {
			t.Errorf("expected match")
		}
	})

	t.Run("literal", func(t *testing.T) {
		u := NewUri("GET", `"/foo/bar"`)
		if match := u.IsMatch(NewUri("GET", `"/foo/bar"`)); !match {
			t.Errorf("expected match")
		}
		if match := u.IsMatch(NewUri("GET", `"/bar/foo"`)); match {
			t.Errorf("expected not to match")
		}
	})

	t.Run("literal binary opp", func(t *testing.T) {
		u := NewUri("GET", `"/foo/bar"`)
		if match := u.IsMatch(NewUri("GET", `"/foo" + "/bar"`)); !match {
			t.Errorf("expected match")
		}
		if match := u.IsMatch(NewUri("GET", `"/bar" + "/foo"`)); match {
			t.Errorf("expected not to match")
		}
	})

	t.Run("binary opp with vars", func(t *testing.T) {
		u := NewUri("GET", `"/foo/bar"`)
		if match := u.IsMatch(NewUri("GET", `"/foo" + someVar`)); !match {
			t.Errorf("expected match")
		}
		if match := u.IsMatch(NewUri("GET", `"/bar" + someVar`)); match {
			t.Errorf("expected not to match")
		}
	})

	t.Run("call with vars", func(t *testing.T) {
		u := NewUri("GET", `"/foo/bar"`)
		if match := u.IsMatch(NewUri("GET", `fmt.Sprintf("/foo/%s", someVar)`)); !match {
			t.Errorf("expected match")
		}
		if match := u.IsMatch(NewUri("GET", `fmt.Sprintf("/bar/%s", someVar)`)); match {
			t.Errorf("expected not to match")
		}
	})

	t.Run("call with too many segs", func(t *testing.T) {
		u := NewUri("GET", "/foo/bar")
		if match := u.IsMatch(NewUri("GET", `fmt.Sprintf("/foo/bar/%s", someVar)`)); match {
			t.Errorf("expected not to match")
		}
	})

	t.Run("call with var in middle", func(t *testing.T) {
		u := NewUri("GET", `"/foo/bar/foobar"`)
		if match := u.IsMatch(NewUri("GET", `fmt.Sprintf("/foo/%s/foobar", someVar)`)); !match {
			t.Errorf("expected to match")
		}
		if match := u.IsMatch(NewUri("GET", `fmt.Sprintf("/foo/%s/foo", someVar)`)); match {
			t.Errorf("expected not to match")
		}
	})

	t.Run("different verbs no match", func(t *testing.T) {
		u := NewUri("GET", `"/foo/bar"`)
		if match := u.IsMatch(NewUri("POST", `"/foo/bar"`)); match {
			t.Fatalf("expected not to match")
		}
	})

	t.Run("root matches exactly", func(t *testing.T) {
		u := NewUri("GET", `"/"`)
		if match := u.IsMatch(NewUri("GET", `"/"`)); !match {
			t.Errorf("expected root to match root")
		}
		if match := u.IsMatch(NewUri("GET", `"/foo"`)); match {
			t.Errorf("expected root not to match non-root")
		}
	})

	t.Run("wildcard path suffix {$} in the root", func(t *testing.T) {
		u := NewUri("GET", `"/"`)
		if match := u.IsMatch(NewUri("GET", `"/{$}"`)); !match {
			t.Errorf("expected root to match root")
		}
	})

	t.Run("exact literal match", func(t *testing.T) {
		u := NewUri("GET", `"/foo/bar"`)
		if match := u.IsMatch(NewUri("GET", `"/foo/bar"`)); !match {
			t.Errorf("expected match")
		}
		if match := u.IsMatch(NewUri("GET", `"/bar/foo"`)); match {
			t.Errorf("expected not to match")
		}
	})

	t.Run("wildcard segment {}", func(t *testing.T) {
		u := NewUri("GET", `"/foo/{}"`)
		if match := u.IsMatch(NewUri("GET", `"/foo/bar"`)); !match {
			t.Errorf("expected wildcard to match")
		}
		if match := u.IsMatch(NewUri("GET", `"/foo"`)); match {
			t.Errorf("expected not to match with missing segment")
		}
	})

	t.Run("wildcard path suffix {$} not in the root", func(t *testing.T) {
		u := NewUri("GET", `"/foo/bar"`)
		if match := u.IsMatch(NewUri("GET", `"/foo/{$}"`)); match {
			t.Errorf("expected suffix wildcard not to match")
		}
	})

	t.Run("trailing slash sensitivity", func(t *testing.T) {
		u := NewUri("GET", `"/foo"`)
		if match := u.IsMatch(NewUri("GET", `"/foo/"`)); !match {
			t.Errorf("expected match /foo with /foo/")
		}
	})

	t.Run("exact vs prefix mismatch", func(t *testing.T) {
		u := NewUri("GET", `"/foo/bar"`)
		if match := u.IsMatch(NewUri("GET", `"/foo/bar/baz"`)); match {
			t.Errorf("expected not to match longer path without {$}")
		}
	})

	t.Run("verb mismatch", func(t *testing.T) {
		u := NewUri("GET", `"/foo/bar"`)
		if match := u.IsMatch(NewUri("POST", `"/foo/bar"`)); match {
			t.Errorf("expected verbs to differ and not match")
		}
	})

	t.Run("long match with mixed variables", func(t *testing.T) {
		u := NewUri("GET", `"/foo/{}/baz"`)
		if match := u.IsMatch(NewUri("GET", `"/foo/bar/baz"`)); !match {
			t.Errorf("expected variable match in middle")
		}
	})
}
