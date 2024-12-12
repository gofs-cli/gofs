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
	PkFields         []AstField
	SearchableFields []AstField
}

func getAstStruct(gofile, gostruct, gopackage string) (*AstStruct, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, gofile, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	for _, decl := range f.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok {
			// find the struct declaration
			if genDecl.Tok == token.TYPE && genDecl.Specs[0].(*ast.TypeSpec).Name.Name == gostruct {
				t := genDecl.Specs[0].(*ast.TypeSpec)
				a := AstStruct{
					Package:    gopackage,
					StructName: gostruct,
				}
				if t.Type == nil {
					return nil, fmt.Errorf("struct %s has no type", gostruct)
				}
				// parse the gofs tags for pk and searchable fields
				for _, field := range t.Type.(*ast.StructType).Fields.List {
					if strings.Contains(field.Tag.Value, "gofs") {
						a.AllFields = append(a.AllFields, AstField{
							StructName:  gostruct,
							FieldNumber: len(a.AllFields),
							FieldName:   field.Names[0].Name,
							FieldType:   field.Type.(*ast.Ident).Name,
							DBType:      DbType(field.Type.(*ast.Ident).Name),
						})
						if strings.Contains(field.Tag.Value, "gofs:\"pk\"") {
							a.PkFields = append(a.PkFields, AstField{
								StructName:  gostruct,
								FieldNumber: len(a.PkFields),
								FieldName:   field.Names[0].Name,
								FieldType:   field.Type.(*ast.Ident).Name,
								DBType:      DbType(field.Type.(*ast.Ident).Name),
							})
						}
						if strings.Contains(field.Tag.Value, "gofs:\"searchable\"") {
							a.SearchableFields = append(a.SearchableFields, AstField{
								StructName:  gostruct,
								FieldNumber: len(a.SearchableFields),
								FieldName:   field.Names[0].Name,
								FieldType:   field.Type.(*ast.Ident).Name,
								DBType:      DbType(field.Type.(*ast.Ident).Name),
							})
						}

					}
				}
				return &a, nil
			}
		}
	}

	return nil, fmt.Errorf("struct %s not found in file %s", gostruct, gofile)
}

func DbType(t string) string {
	switch t {
	case "string":
		return "TEXT"
	case "int":
		return "INTEGER"
	default:
		return "TEXT"
	}
}
