package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/kynrai/gofs/internal/codegen"
	"github.com/kynrai/gofs/internal/gofs"
)

const codegenUsage = `usage: gofs codegen [template] -struct=[struct name]

Experimental: "codegen" generates code from go templates. 
This should be used as a go:generate directive in the source code.

Example:
//go:generate gofs codegen db -struct=Foo
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
		return
	}

	gopackage := os.Getenv("GOPACKAGE")
	if gopackage == "" {
		fmt.Println("codegen: GOPACKAGE not set")
		return
	}

	gostruct := ""
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "-struct=") {
			gostruct = strings.TrimPrefix(arg, "-struct=")
			break
		}
	}
	if gostruct == "" {
		fmt.Println("codegen: struct name not set")
		return
	}

	projectRoot, err := findProjectRoot()
	if err != nil {
		fmt.Println("codegen: ", err)
		return
	}

	templates, err := gofs.LoadTemplates(projectRoot)
	if err != nil {
		fmt.Println("codegen: Error loading templates ", err)
		return
	}

	for _, template := range templates {
		if slices.Contains(os.Args, template.Name) {
			o := strings.ToLower(gostruct) + template.Suffix
			if template.OutputDir != "" {
				o = filepath.Join(projectRoot, template.OutputDir, o)
			}
			t := filepath.Join(projectRoot, gofs.GofsDir, template.Tmpl)
			err := codegen.Codegen(gofile, gopackage, gostruct, o, t)
			if err != nil {
				fmt.Println("codegen: ", err)
				return
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
		if _, err := os.Stat(filepath.Join(cwd, gofs.GofsDir)); err == nil {
			return cwd, nil
		}
		cwd = filepath.Clean(cwd)
	}
	return "", fmt.Errorf("gofs directory not found")
}
