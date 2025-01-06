package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/gofs-cli/gofs/internal/codegen"
	"github.com/gofs-cli/gofs/internal/tmpl"
)

const codegenUsage = `usage: gofs codegen [template] [template] ...

Experimental: "codegen" generates code from go templates. 
This should be used as a go:generate directive in the source code.

Example:
//go:generate gofs codegen db
type Foo struct {
	ID   string  'json:"id"  gofs:"pk"'
	Bar  string  'json:"bar" gofs:"searchable'
}

`

func init() {
	Gofs.AddCmd(Command{
		Name:  "codegen",
		Short: "generate go code from struct",
		Long:  codegenUsage,
		Cmd:   cmdCodegen,
	})
}

func cmdCodegen() {
	gofile := os.Getenv("GOFILE")
	if gofile == "" {
		fmt.Println("codegen: GOFILE not set")
		os.Exit(1)
	}

	gopackage := os.Getenv("GOPACKAGE")
	if gopackage == "" {
		fmt.Println("codegen: GOPACKAGE not set")
		os.Exit(1)
	}

	golinestr := os.Getenv("GOLINE")
	if golinestr == "" {
		fmt.Println("codegen: GOLINE not set")
		os.Exit(1)
	}
	goline, err := strconv.Atoi(golinestr)
	if err != nil {
		fmt.Println("codegen: GOLINE not a number")
		os.Exit(1)
	}

	projectRoot, err := findProjectRoot()
	if err != nil {
		fmt.Println("codegen: ", err)
		os.Exit(1)
	}

	templates, err := codegen.LoadTemplates(filepath.Join(projectRoot, GofsDir))
	if err != nil {
		fmt.Println("codegen: Error loading templates ", err)
		os.Exit(1)
	}

	// the struct is on the line after go generate
	a, err := codegen.GetAstStruct(gofile, gopackage, goline+1)
	if err != nil || a == nil {
		fmt.Println("codegen: ", err)
		os.Exit(1)
	}

	for _, template := range templates {
		if slices.Contains(os.Args, template.Name) {
			t := filepath.Join(projectRoot, GofsDir, codegen.TemplatesDir, template.Tmpl)
			o := strings.ToLower(a.StructName) + template.Suffix
			if template.OutputDir != "" {
				o = filepath.Join(projectRoot, template.OutputDir, o)
			}

			err = tmpl.Generate(o, t, a)
			if err != nil {
				fmt.Println("codegen: ", err)
				os.Exit(1)
			}
		}
	}
}

func findProjectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if cwd == "/" {
			break
		}
		cwd, _ = filepath.Split(cwd)
		// the gofs directory is the root of the project
		if _, err := os.Stat(filepath.Join(cwd, GofsDir)); err == nil {
			return cwd, nil
		}
		cwd = filepath.Clean(cwd)
	}
	return "", fmt.Errorf("gofs directory not found")
}
