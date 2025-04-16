package cmd

import (
	"flag"

	"github.com/gofs-cli/gofs/internal/lsp"
)

const lspUsage = `usage: gofs lsp

Experimental: "lsp" starts the gofs language server.
`

func init() {
	Gofs.AddCmd(Command{
		Name:  "lsp",
		Short: "Experimental: start the gofs language server",
		Long:  initUsage,
		Cmd:   cmdLsp,
	})
}

func cmdLsp() {
	fs := flag.NewFlagSet("lsp", flag.ExitOnError)
	debug := fs.Bool("debug", false, "enable debug logging")
	_ = fs.Parse(flag.Args()[1:]) // skip "lsp" itself

	lsp.Start(*debug)
}
