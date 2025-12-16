package folder

import "embed"

//go:embed all:*
var Folder embed.FS

var ModuleName = "github.com/gofs-cli/gofs/templates/fs-app"
