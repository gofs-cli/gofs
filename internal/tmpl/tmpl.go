package tmpl

import (
	"os"
	"text/template"
)

// Returns a template with the standard gofs function map
func New(name string, text string) (*template.Template, error) {
	funcs := template.FuncMap{
		"Snake": Snake,
		"Add": func(i, j int) int {
			return i + j
		},
	}

	return template.New(name).Funcs(funcs).Parse(text)
}

func Generate(targetFile, templateFile string, a any) error {
	b, err := os.ReadFile(templateFile)
	if err != nil {
		return err
	}
	tmpl := string(b)

	f, err := os.Create(targetFile)
	if err != nil {
		return err
	}
	defer f.Close()

	t, err := New("template", tmpl)
	if err != nil {
		return err
	}
	return t.Execute(f, a)
}
