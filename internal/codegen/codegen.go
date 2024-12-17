package codegen

import (
	"fmt"

	"github.com/kynrai/gofs/internal/tmpl"
)

func Codegen(gofile, gopackage, gostruct, output, template string) error {
	a, err := getAstStruct(gofile, gostruct, gopackage)
	if err != nil {
		return fmt.Errorf("codegen: %s", err)
	}
	if a == nil {
		return fmt.Errorf("codegen: struct %s not found in file %s", gostruct, gofile)
	}

	err = tmpl.Generate(output, template, a)
	if err != nil {
		return fmt.Errorf("codegen: %s", err)
	}

	return nil
}
