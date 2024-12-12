package cmd

import (
	"fmt"
	folder "module/placeholder"
	"os"

	"github.com/kynrai/gofs/internal/gen"
)

const (
	root              = "template"
	defaultModuleName = "module/placeholder"
)

const initUsage = `usage: gofs init [module-name] [dir]

"init" initializes a new module in the specified directory.
If no directory is specified, the current directory is used.

The module name should be a go module name, e.g. "github.com/user/module".

Example:
  gofs init mymodule /path/to/dir
  gofs init mymodule

`

func init() {
	Gofs.AddCmd(Command{
		Name:  "init",
		Short: "initialize a new gofs mdodule",
		Long:  initUsage,
		Cmd:   cmdInit,
	})
}

func cmdInit() error {
	args := os.Args[2:] // skip program name and command

	moduleName := ""
	dir := ""
	var err error

	switch {
	case len(args) == 0:
		return fmt.Errorf("init: missing module name")
	case len(args) == 1:
		moduleName = args[0]
		dir, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("init: %s", err)
		}
	case len(args) == 2:
		moduleName = args[0]
		dir = args[1]
	default:
		return fmt.Errorf("init: too many arguments")
	}

	parser := gen.NewParser(dir, defaultModuleName, moduleName, folder.Folder)
	parser.Parse()
	return nil
}
