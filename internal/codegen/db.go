package codegen

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/kynrai/gofs/internal/tmpl"
)

const (
	dbCrudTemplate = "db.tmpl"
	dbSqlTemplate  = "sql.tmpl"
	gofsDir        = ".gofs"
	sqlFileSuffix  = "s_generated.sql"
	goFileSuffix   = "_db_generated.go"
)

func CodegenDB(gofile, gopackage, gostruct, migrationsDir, projectRoot string) error {
	a, err := getAstStruct(gofile, gostruct, gopackage)
	if err != nil {
		return fmt.Errorf("codegen: %s", err)
	}
	if a == nil {
		return fmt.Errorf("codegen: struct %s not found in file %s", gostruct, gofile)
	}

	filename := strings.Split(filepath.Base(strings.ToLower(gostruct)), ".")[0]

	// generate the sql file in the migrations directory
	err = tmpl.Generate(
		filepath.Join(projectRoot, migrationsDir, filename+sqlFileSuffix),
		filepath.Join(projectRoot, gofsDir, dbSqlTemplate),
		a)
	if err != nil {
		return fmt.Errorf("codegen: %s", err)
	}

	// generate the go file in the current directory
	err = tmpl.Generate(
		filename+goFileSuffix,
		filepath.Join(projectRoot, gofsDir, dbCrudTemplate),
		a)
	if err != nil {
		return fmt.Errorf("codegen: %s", err)
	}

	return nil
}
