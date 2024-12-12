package codegen

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/kynrai/gofs/internal/tmpl"
)

const (
	dbCrudTemplate = "db.tmpl"
	dbSqlTemplate  = "sql.tmpl"
	gofsDir        = ".gofs"
	migrationsDir  = "internal/db/migrations"
)

func Generate(gofile, gopackage string) error {
	gostruct := ""
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "-struct=") {
			gostruct = strings.TrimPrefix(arg, "-struct=")
			break
		}
	}
	if gostruct == "" {
		return fmt.Errorf("codegen: struct name not set")
	}

	a, err := getAstStruct(gofile, gostruct, gopackage)
	if err != nil {
		return fmt.Errorf("codegen: %s", err)
	}
	if a == nil {
		return fmt.Errorf("codegen: struct %s not found in file %s", gostruct, gofile)
	}

	projectRoot, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("codegen: %s", err)
	}
	filename := strings.Split(filepath.Base(strings.ToLower(gostruct)), ".")[0]

	if slices.Contains(os.Args, "db") {
		err = tmpl.Generate(
			filepath.Join(projectRoot, migrationsDir, filename+"s_generated.sql"),
			filepath.Join(projectRoot, gofsDir, dbSqlTemplate),
			a)
		if err != nil {
			return fmt.Errorf("codegen: %s", err)
		}

		err = tmpl.Generate(
			filename+"_db_generated.go",
			filepath.Join(projectRoot, gofsDir, dbCrudTemplate),
			a)
		if err != nil {
			return fmt.Errorf("codegen: %s", err)
		}
	}

	return nil
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
