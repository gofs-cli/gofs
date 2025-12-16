module github.com/gofs-cli/gofs

go 1.25.5

require (
	github.com/a-h/templ v0.3.960
	github.com/gofs-cli/azure-app-template v0.0.4
	github.com/gofs-cli/template v1.0.8
	golang.org/x/mod v0.31.0
	golang.org/x/tools v0.40.0
)

require (
	github.com/a-h/parse v0.0.0-20250122154542-74294addb73e // indirect
	github.com/gofs-cli/gofs/templates/fs-app v0.0.0-00010101000000-000000000000
)

replace github.com/gofs-cli/gofs/templates/fs-app => ./templates/fs-app
