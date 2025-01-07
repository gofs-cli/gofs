module github.com/gofs-cli/gofs

go 1.23.2

require (
	github.com/a-h/templ v0.3.819
	github.com/gofs-cli/gofs/template v0.0.0-20250107141900-45361be18d14
	golang.org/x/mod v0.22.0
	golang.org/x/tools v0.29.0
)

require github.com/a-h/parse v0.0.0-20240121214402-3caf7543159a // indirect

// This is useful for local development of the template.
// replace github.com/gofs-cli/gofs/template => ./template
