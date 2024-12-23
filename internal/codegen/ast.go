package codegen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

type AstField struct {
	StructName  string
	FieldNumber int
	FieldName   string
	FieldType   string
	DBType      string
}

type AstStruct struct {
	Package          string
	StructName       string
	AllFields        []AstField
	GofsFields       []AstField
	PkFields         []AstField
	SearchableFields []AstField
}

func GetAstStruct(gofile, gopackage string, goline int) (*AstStruct, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, gofile, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	t, err := findDecl(goline, f, fset)
	if err != nil {
		return nil, err
	}

	return astStructFromStructType(gopackage, t)
}

func dbType(t string) string {
	switch t {
	case "string":
		return "TEXT"
	case "int":
		return "INTEGER"
	default:
		return "TEXT"
	}
}

func getType(a ast.Expr) (string, error) {
	switch a := a.(type) {
	case *ast.Ident:
		return a.Name, nil
	case *ast.SelectorExpr:
		return a.X.(*ast.Ident).Name + "." + a.Sel.Name, nil
	case *ast.ArrayType:
		switch elt := a.Elt.(type) {
		case *ast.Ident:
			return "[]" + elt.Name, nil
		case *ast.SelectorExpr:
			ft, err := getType(elt)
			if err != nil {
				return "", err
			}
			return "[]" + ft, nil
		}
		return a.Elt.(*ast.Ident).Name, nil
	default:
		return "", fmt.Errorf("unhandled type")
	}
}

func findDecl(goline int, f *ast.File, fset *token.FileSet) (*ast.TypeSpec, error) {
	for _, decl := range f.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok {
			// skip declarations that are not types or on the line we are looking for
			if fset.Position(genDecl.TokPos).Line != goline || genDecl.Tok != token.TYPE {
				continue
			}
			// check that its a struct
			if _, ok := genDecl.Specs[0].(*ast.TypeSpec).Type.(*ast.StructType); !ok {
				return nil, fmt.Errorf("type is not a struct")
			}
			return genDecl.Specs[0].(*ast.TypeSpec), nil
		}
	}
	return nil, fmt.Errorf("struct not found")
}

func astStructFromStructType(gopackage string, t *ast.TypeSpec) (*AstStruct, error) {
	a := AstStruct{
		Package:    gopackage,
		StructName: t.Name.Name,
	}
	// parse the gofs tags for pk and searchable fields
	for _, field := range t.Type.(*ast.StructType).Fields.List {
		ft, err := getType(field.Type)
		if err != nil {
			return nil, err
		}

		a.AllFields = append(a.AllFields, AstField{
			StructName:  t.Name.Name,
			FieldNumber: len(a.AllFields),
			FieldName:   field.Names[0].Name,
			FieldType:   ft,
		})

		if field.Tag == nil {
			continue
		}

		if strings.Contains(field.Tag.Value, "gofs") {
			a.GofsFields = append(a.GofsFields, AstField{
				StructName:  t.Name.Name,
				FieldNumber: len(a.GofsFields),
				FieldName:   field.Names[0].Name,
				FieldType:   ft,
				DBType:      dbType(ft),
			})
			if strings.Contains(field.Tag.Value, "gofs:\"pk\"") {
				a.PkFields = append(a.PkFields, AstField{
					StructName:  t.Name.Name,
					FieldNumber: len(a.PkFields),
					FieldName:   field.Names[0].Name,
					FieldType:   ft,
					DBType:      dbType(ft),
				})
			}
			if strings.Contains(field.Tag.Value, "gofs:\"searchable\"") {
				a.SearchableFields = append(a.SearchableFields, AstField{
					StructName:  t.Name.Name,
					FieldNumber: len(a.SearchableFields),
					FieldName:   field.Names[0].Name,
					FieldType:   ft,
					DBType:      dbType(ft),
				})
			}

		}
	}
	return &a, nil
}
