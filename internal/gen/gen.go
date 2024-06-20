package gen

import (
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/ast/astutil"
)

type Parser struct {
	// DirPath is the path to the folder to parse.
	// This should be a directory containing the go files to parse.
	DirPath        string
	CurrentModName string
	NewModName     string
}

func NewParser(dirPath, defaultModuleName, newModuleName string) *Parser {
	return &Parser{
		DirPath:        dirPath,
		CurrentModName: defaultModuleName,
		NewModName:     newModuleName}
}

func (p *Parser) Parse() error {
	fileSystem := os.DirFS(p.DirPath)
	return fs.WalkDir(fileSystem, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		path = filepath.Join(p.DirPath, path)
		if !d.IsDir() {
			if strings.HasSuffix(path, ".mod") {
				err := p.UpdateMod(path, p.NewModName)
				if err != nil {
					log.Fatal(err)
				}
			}
			if strings.HasSuffix(path, ".go") {
				err := p.UpdateFile(path, p.CurrentModName, p.NewModName)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
		return nil
	})
}

func (p *Parser) UpdateMod(path, modName string) error {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	file, err := modfile.Parse(path, bytes, nil)
	if err != nil {
		return err
	}
	file.AddModuleStmt(modName)

	newBytes := modfile.Format(file.Syntax)
	return os.WriteFile(path, newBytes, 0644)
}

func (p *Parser) UpdateFile(path, oldModName, newModName string) error {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	imports := astutil.Imports(fset, file)
	for _, para := range imports {
		for _, imp := range para {
			oldPath, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				return err
			}
			if strings.Contains(oldPath, oldModName) {
				newPath := strings.Replace(oldPath, oldModName, newModName, 1)
				rewritten := astutil.RewriteImport(fset, file, oldPath, newPath)
				if !rewritten {
					return err
				}
			}
		}
	}
	newFile, err := os.OpenFile(path, os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	return format.Node(newFile, fset, file)
}
