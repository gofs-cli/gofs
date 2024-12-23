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

func getAstStruct(gofile, gostruct, gopackage string) (*AstStruct, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, gofile, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	st, err := findDecl(gostruct, f)
	if err != nil {
		return nil, err
	}

	return astStructFromStructType(gostruct, gopackage, st)
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

func findDecl(gostruct string, f *ast.File) (*ast.StructType, error) {
	for _, decl := range f.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok {
			// find the struct declaration
			if genDecl.Tok != token.TYPE || genDecl.Specs[0].(*ast.TypeSpec).Name.Name != gostruct {
				continue
			}
			// we know gostruct is a typespec, so check that its a struct
			t, ok := genDecl.Specs[0].(*ast.TypeSpec).Type.(*ast.StructType)
			if !ok {
				return nil, fmt.Errorf("type %s is not a struct", gostruct)
			}
			return t, nil
		}
	}
	return nil, fmt.Errorf("struct %s not found", gostruct)
}

func astStructFromStructType(gostruct, gopackage string, st *ast.StructType) (*AstStruct, error) {
	a := AstStruct{
		Package:    gopackage,
		StructName: gostruct,
	}
	// parse the gofs tags for pk and searchable fields
	for _, field := range st.Fields.List {
		ft, err := getType(field.Type)
		if err != nil {
			return nil, err
		}

		a.AllFields = append(a.AllFields, AstField{
			StructName:  gostruct,
			FieldNumber: len(a.AllFields),
			FieldName:   field.Names[0].Name,
			FieldType:   ft,
		})

		if field.Tag == nil {
			continue
		}

		if strings.Contains(field.Tag.Value, "gofs") {
			a.GofsFields = append(a.GofsFields, AstField{
				StructName:  gostruct,
				FieldNumber: len(a.GofsFields),
				FieldName:   field.Names[0].Name,
				FieldType:   ft,
				DBType:      dbType(ft),
			})
			if strings.Contains(field.Tag.Value, "gofs:\"pk\"") {
				a.PkFields = append(a.PkFields, AstField{
					StructName:  gostruct,
					FieldNumber: len(a.PkFields),
					FieldName:   field.Names[0].Name,
					FieldType:   ft,
					DBType:      dbType(ft),
				})
			}
			if strings.Contains(field.Tag.Value, "gofs:\"searchable\"") {
				a.SearchableFields = append(a.SearchableFields, AstField{
					StructName:  gostruct,
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
