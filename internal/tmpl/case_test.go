package tmpl

import (
	"testing"
)

type caseTest struct {
	input    string
	expected string
}

func TestSnakeCase(t *testing.T) {
	for _, test := range []caseTest{
		{"ID", "id"},
		{"fooID", "foo_id"},
		{"word", "word"},
		{"TwoWords", "two_words"},
		{"two_words", "two_words"},
		{"123abC", "123ab_c"},
		{"can'tBe-apostrophe", "cant_be_apostrophe"},
	} {
		actual := Snake(test.input)
		if test.expected != actual {
			t.Fatalf("expected %s, got %s", test.expected, actual)
		}
	}
}
