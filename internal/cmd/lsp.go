package cmd

import (
	"flag"
	"os"

	"github.com/gofs-cli/gofs/internal/lsp"
)

const lspUsage = `usage: gofs lsp [-debug] [-stdio]

Experimental: "lsp" starts the gofs language server.
`

func init() {
	Gofs.AddCmd(Command{
		Name:  "lsp",
		Short: "Experimental: start the gofs language server",
		Long:  lspUsage,
		Cmd:   cmdLsp,
	})
}

func cmdLsp() {
	fs := flag.NewFlagSet("lsp", flag.ExitOnError)
	debug := fs.Bool("debug", false, "enable debug logging")
	err := fs.Parse(os.Args[2:]) // skip program name and command
	if err != nil {
		os.Stderr.WriteString("lsp: error parsing flags: " + err.Error() + "\n")
		os.Exit(1)
	}

	lsp.Start(*debug)
}
