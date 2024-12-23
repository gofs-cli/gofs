package codegen

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	TemplatesDir = "templates"
	cfg          = "templates.json"
)

type Template struct {
	Name      string `json:"name"`
	Tmpl      string `json:"tmpl"`
	OutputDir string `json:"output_dir"`
	Suffix    string `json:"suffix"`
}

func LoadTemplates(gofsDir string) ([]Template, error) {
	cfgFile := filepath.Join(gofsDir, TemplatesDir, cfg)
	f, err := os.Open(cfgFile)
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
