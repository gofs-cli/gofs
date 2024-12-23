package codegen

import (
	"fmt"
	"go/parser"
	"go/token"
	"testing"
)

func TestFindDeclReturnStructType(t *testing.T) {
	t.Parallel()

	input := `package main

type Foo struct {
	Bar  string
}
	`

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", input, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}

	st, err := findDecl(3, f, fset)
	if err != nil || st == nil {
		t.Fatal("Expected to find struct Foo and did not")
	}
	if st.Name.Name != "Foo" {
		t.Fatal("Expected name to be Foo but got ", st.Name.Name)
	}
}

func TestFindDeclReturnError(t *testing.T) {
	t.Parallel()

	input := `package main

type Foo struct {
	Bar  string
}
	`

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", input, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}

	st, err := findDecl(2, f, fset)
	if err == nil || st != nil {
		t.Fatal("Should not find struct on line 2 and found it")
	}
}

func TestAstStructFromStructType(t *testing.T) {
	t.Parallel()

	input := `package main

type Foo struct {
	ID        string  ` + "`" + `json:"id"  gofs:"pk"` + "`" + ` 
	Bar       string  ` + "`" + `json:"bar" gofs:"searchable"` + "`" + `
	TestInt   int
	TestTime  time.Time
	TestArray []SomeType
}
	`

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", input, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}

	st, err := findDecl(3, f, fset)
	if err != nil || st == nil {
		t.Fatal("Expected to find struct Foo and did not")
	}

	astStruct, err := astStructFromStructType("main", st)
	if err != nil || astStruct == nil {
		t.Fatal(err)
	}

	expected := &AstStruct{
		Package:    "main",
		StructName: "Foo",
		AllFields: []AstField{
			{
				StructName:  "Foo",
				FieldNumber: 0,
				FieldName:   "ID",
				FieldType:   "string",
			},
			{
				StructName:  "Foo",
				FieldNumber: 1,
				FieldName:   "Bar",
				FieldType:   "string",
			},
			{
				StructName:  "Foo",
				FieldNumber: 2,
				FieldName:   "TestInt",
				FieldType:   "int",
			},
			{
				StructName:  "Foo",
				FieldNumber: 3,
				FieldName:   "TestTime",
				FieldType:   "time.Time",
			},
			{
				StructName:  "Foo",
				FieldNumber: 4,
				FieldName:   "TestArray",
				FieldType:   "[]SomeType",
			},
		},
		GofsFields: []AstField{
			{
				StructName:  "Foo",
				FieldNumber: 0,
				FieldName:   "ID",
				FieldType:   "string",
				DBType:      "TEXT",
			},
			{
				StructName:  "Foo",
				FieldNumber: 1,
				FieldName:   "Bar",
				FieldType:   "string",
				DBType:      "TEXT",
			},
		},
		PkFields: []AstField{
			{
				StructName:  "Foo",
				FieldNumber: 0,
				FieldName:   "ID",
				FieldType:   "string",
				DBType:      "TEXT",
			},
		},
		SearchableFields: []AstField{
			{
				StructName:  "Foo",
				FieldNumber: 0,
				FieldName:   "Bar",
				FieldType:   "string",
				DBType:      "TEXT",
			},
		},
	}

	if fmt.Sprintf("%+v", astStruct) != fmt.Sprintf("%+v", expected) {
		t.Fatalf("Expected:\n%+v\n but got\n%+v\n", expected, astStruct)
	}
}
