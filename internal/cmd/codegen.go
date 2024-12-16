package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/kynrai/gofs/internal/codegen"
)

const codegenUsage = `usage: gofs codegen db -struct=[struct name] < -migrationsDir=[dir] >

Experimental: "codegen" generates helper functions for reading and writing a struct to a database
based on gofs decorations. This should be used as a go:generate directive in the source code.
'migrationsDir' is an optional parameter to the directory where the generated sql file will be saved, 
and defaults to internal/db/migrations if not provided.

Example:
//go:generate gofs codegen db -struct=Foo
type Foo struct {
	ID   string  'json:"id"  gofs:"pk"'
	Bar  string  'json:"bar" gofs:"searchable'
}

`

const (
	dbCrudTemplate = "db.tmpl"
	dbSqlTemplate  = "sql.tmpl"
	gofsDir        = ".gofs"
	migrationsDir  = "internal/db/migrations"
)

func init() {
	Gofs.AddCmd(Command{
		Name:  "codegen",
		Short: "generate gofs db helper functions for a struct",
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

	md := migrationsDir
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "-migrationsDir=") {
			md = strings.TrimPrefix(arg, "-migrationsDir=")
			break
		}
	}

	projectRoot, err := findProjectRoot()
	if err != nil {
		fmt.Println("codegen: ", err)
		return
	}

	if slices.Contains(os.Args, "db") {
		err := codegen.CodegenDB(gofile, gopackage, gostruct, md, projectRoot)
		if err != nil {
			fmt.Println(err)
			return
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
		if _, err := os.Stat(filepath.Join(cwd, gofsDir)); err == nil {
			return cwd, nil
		}
		cwd = filepath.Clean(cwd)
	}
	return "", fmt.Errorf("gofs directory not found")
}
