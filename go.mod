module github.com/kynrai/gofs

go 1.23.0

require (
	github.com/a-h/templ v0.3.819
	golang.org/x/mod v0.20.0
	golang.org/x/tools v0.24.0
	module/placeholder v0.0.0
)

require github.com/a-h/parse v0.0.0-20240121214402-3caf7543159a // indirect

replace module/placeholder => ./template
