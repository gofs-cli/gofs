package pkg

import "github.com/gofs-cli/gofs/internal/lsp/model"

type Func struct {
	Name string
	File string
	Pos  model.Pos
}

type Pkg struct {
	Files []string
	Funcs []Func
}

func (p *Pkg) GetFunc(name string) *Func {
	for _, f := range p.Funcs {
		if f.Name == name {
			return &f
		}
	}
	return nil
}
