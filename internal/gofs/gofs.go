package gofs

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	GofsDir       = ".gofs"
	TemplatesFile = "templates.json"
)

type Template struct {
	Name      string `json:"name"`
	Tmpl      string `json:"tmpl"`
	OutputDir string `json:"output_dir"`
	Suffix    string `json:"suffix"`
}

func LoadTemplates(projectRoot string) ([]Template, error) {
	templatesFile := filepath.Join(projectRoot, GofsDir, TemplatesFile)
	f, err := os.Open(templatesFile)
	if err != nil {
		return nil, err
	}

	var t []Template
	err = json.NewDecoder(f).Decode(&t)
	if err != nil {
		return nil, err
	}

	return t, nil
}
