package templFile

import (
	"fmt"

	"github.com/a-h/parse"
	templ "github.com/a-h/templ/parser/v2"
)

type NodeParser struct {
	TextFunc                   func(templ.Text) error
	ElementFunc                func(templ.Element) error
	RawElementFunc             func(templ.RawElement) error
	GoCommentFunc              func(templ.GoComment) error
	HTMLCommentFunc            func(templ.HTMLComment) error
	CallTemplateExpressionFunc func(templ.CallTemplateExpression) error
	TemplElementExpressionFunc func(templ.TemplElementExpression) error
	ChildrenExpressionFunc     func(templ.ChildrenExpression) error
	IfExpressionFunc           func(templ.IfExpression) error
	SwitchExpressionFunc       func(templ.SwitchExpression) error
	ForExpressionFunc          func(templ.ForExpression) error
	StringExpressionFunc       func(templ.StringExpression) error
	GoCodeFunc                 func(templ.GoCode) error
	WhitespaceFunc             func(templ.Whitespace) error
	DocType                    func(templ.DocType) error
}

func (np NodeParser) RecurseNode(n templ.Node) {
	switch n := n.(type) {
	case templ.Text:
		if np.TextFunc != nil {
			np.TextFunc(n)
		}
	case templ.Element:
		if np.ElementFunc != nil {
			np.ElementFunc(n)
		}
		for _, c := range n.Children {
			np.RecurseNode(c)
		}
	case templ.RawElement:
		if np.RawElementFunc != nil {
			np.RawElementFunc(n)
		}
	case templ.GoComment:
		if np.GoCommentFunc != nil {
			np.GoCommentFunc(n)
		}
	case templ.HTMLComment:
		if np.HTMLCommentFunc != nil {
			np.HTMLCommentFunc(n)
		}
	case templ.CallTemplateExpression:
		if np.CallTemplateExpressionFunc != nil {
			np.CallTemplateExpressionFunc(n)
		}
	case templ.TemplElementExpression:
		if np.TemplElementExpressionFunc != nil {
			np.TemplElementExpressionFunc(n)
		}
		for _, c := range n.Children {
			np.RecurseNode(c)
		}
	case templ.ChildrenExpression:
		if np.ChildrenExpressionFunc != nil {
			np.ChildrenExpressionFunc(n)
		}
	case templ.IfExpression:
		if np.IfExpressionFunc != nil {
			np.IfExpressionFunc(n)
		}
		for _, c := range n.Then {
			np.RecurseNode(c)
		}
		for _, c := range n.Else {
			np.RecurseNode(c)
		}
	case templ.SwitchExpression:
		if np.SwitchExpressionFunc != nil {
			np.SwitchExpressionFunc(n)
		}
		for _, cse := range n.Cases {
			for _, c := range cse.Children {
				np.RecurseNode(c)
			}
		}
	case templ.ForExpression:
		if np.ForExpressionFunc != nil {
			np.ForExpressionFunc(n)
		}
		for _, c := range n.Children {
			np.RecurseNode(c)
		}
	case templ.StringExpression:
		if np.StringExpressionFunc != nil {
			np.StringExpressionFunc(n)
		}
	case templ.GoCode:
		if np.GoCodeFunc != nil {
			np.GoCodeFunc(n)
		}
	case templ.Whitespace:
		if np.WhitespaceFunc != nil {
			np.WhitespaceFunc(n)
		}
	case templ.DocType:
		if np.DocType != nil {
			np.DocType(n)
		}
	}
}

func ElementAttribute(a []templ.Attribute, name string) *templ.ConstantAttribute {
	for _, v := range a {
		switch v := v.(type) {
		case templ.BoolConstantAttribute:
			continue
		case templ.ConstantAttribute:
			if v.Name == name {
				// add double quotes around the value
				return &templ.ConstantAttribute{
					Value: fmt.Sprintf("\"%s\"", v.Value),
					NameRange: templ.NewRange(
						parse.Position{
							Index: int(v.NameRange.To.Index) + 2,
							Line:  int(v.NameRange.To.Line),
							Col:   int(v.NameRange.To.Col) + 2,
						},
						parse.Position{
							Index: int(v.NameRange.To.Index) + 2 + len(v.Value),
							Line:  int(v.NameRange.To.Line),
							Col:   int(v.NameRange.To.Col) + 2 + len(v.Value),
						},
					),
				}
			}
		case templ.BoolExpressionAttribute:
			continue
		case templ.ExpressionAttribute:
			if v.Name == name {
				return &templ.ConstantAttribute{
					Value:     v.Expression.Value,
					NameRange: v.Expression.Range,
				}
			}
		case templ.SpreadAttributes:
			// TODO check whether the go code returns the attribute
			continue
		case templ.ConditionalAttribute:
			if t := ElementAttribute(v.Then, name); t != nil {
				return t
			}
			if t := ElementAttribute(v.Else, name); t != nil {
				return t
			}
		}
	}
	return nil
}
