package cmd

import (
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
	lsp.Start()
}
