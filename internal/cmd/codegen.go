package cmd

import (
	"fmt"
	"os"

	"github.com/kynrai/gofs/internal/codegen"
)

const codegenUsage = `usage: gofs codegen db -struct=[struct name]

Experimental: "codegen" generates helper functions for reading and writing a struct to a database
based on gofs decorations. This should be used as a go:generate directive in the source code.

Example:
//go:generate gofs codegen db -type=Foo
type Foo struct {
	ID   string  'json:"id"  gofs:"pk"'
	Bar  string  'json:"bar" gofs:"searchable'
}

`

func init() {
	Gofs.AddCmd(Command{
		Name:  "codegen",
		Short: "generate gofs db helper functions for a struct",
		Long:  codegenUsage,
		Cmd:   cmdCodegen,
	})
}

func cmdCodegen() error {
	gofile := os.Getenv("GOFILE")
	if gofile == "" {
		return fmt.Errorf("codegen: GOFILE not set")
	}

	gopackage := os.Getenv("GOPACKAGE")
	if gopackage == "" {
		return fmt.Errorf("codegen: GOPACKAGE not set")
	}

	return codegen.Generate(gofile, gopackage)
}
