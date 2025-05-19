package cmd

import (
	"embed"
	"flag"
	"fmt"
	"os"

	azureTemplate "github.com/gofs-cli/azure-app-template"
	defaultTemplate "github.com/gofs-cli/template"

	"github.com/gofs-cli/gofs/internal/gen"
)

const (
	root              = "template"
	defaultModuleName = "github.com/gofs-cli/template"
)

const initUsage = `usage: gofs init [module-name] [dir]

"init" initializes a new module in the specified directory.
If no directory is specified, the current directory is used.

The module name should be a go module name, e.g. "github.com/user/module".

flags:
  -template
    Name of the template to use for the project. By default this will ues the basic bare bones template.

    Available names:
      - azure
          This template creates an app suitable for deployment to azure apps and expects azure auth tokens from Entra ID


Example:
  gofs init mymodule /path/to/dir
  gofs init mymodule
  gofs init mymodule -template=azure

`

func init() {
	Gofs.AddCmd(Command{
		Name:  "init",
		Short: "initialize a new gofs mdodule",
		Long:  initUsage,
		Cmd:   cmdInit,
	})
}

func cmdInit() {
	var template string
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.StringVar(&template, "template", "default", "the template to use for the generated project")

	args := os.Args[2:] // skip program name and command
	fs.Parse(args)

	fmt.Println("template is: ", template)
	moduleName := ""
	dir := ""
	var err error

	switch {
	case len(args) == 0:
		fmt.Println("init: missing module name")
		fmt.Print(initUsage)
		return
	case len(args) == 1:
		moduleName = args[0]
		dir, err = os.Getwd()
		if err != nil {
			fmt.Println("init: ", err)
			return
		}
	case len(args) == 2:
		moduleName = args[0]
		dir = args[1]
	default:
		fmt.Println("init: too many arguments")
		fmt.Print(initUsage)
		return
	}

	var selectedTemplate embed.FS
	switch template {
	case "azure":
		selectedTemplate = azureTemplate.Folder
	default:
		selectedTemplate = defaultTemplate.Folder
	}
	parser, err := gen.NewParser(dir, defaultModuleName, moduleName, selectedTemplate)
	if err != nil {
		fmt.Println("init: ", err)
		return
	}
	parser.Parse()
}
