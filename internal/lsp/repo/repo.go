package repo

import (
	"errors"
	"log"
	"os"
	"path"
	"sync"

	"golang.org/x/mod/modfile"

	"github.com/gofs-cli/gofs/internal/lsp/model"
	"github.com/gofs-cli/gofs/internal/lsp/pkg"
	routesFile "github.com/gofs-cli/gofs/internal/lsp/routes_file"
	templFile "github.com/gofs-cli/gofs/internal/lsp/templ_file"
	"github.com/gofs-cli/gofs/internal/lsp/uri"
)

type Repo struct {
	IsOpen        bool
	RootPath      string
	HasGofsConfig bool
	Module        string
	rt            routesFile.Routes     // routes file
	ot            []templFile.TemplFile // open templ files
	pkgs          map[string]pkg.Pkg    // loaded packages
	mutex         sync.Mutex
}

func NewRepo() *Repo {
	// open must be called first and creates the repo
	return &Repo{
		IsOpen: false,
	}
}

func (r *Repo) Open(rootPath string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.RootPath = rootPath

	// check if .gofs config folder exists
	if _, err := os.Stat(path.Join(rootPath, ".gofs")); errors.Is(err, os.ErrNotExist) {
		r.IsOpen = true
		r.HasGofsConfig = false
		return nil
	}
	r.HasGofsConfig = true

	// check if go.mod file exists
	modFile, err := os.ReadFile(path.Join(rootPath, "go.mod"))
	if err != nil {
		return err
	}

	mod, err := modfile.Parse(path.Join(rootPath, "go.mod"), modFile, nil)
	if err != nil {
		return err
	}
	r.Module = mod.Module.Mod.Path

	// read routes.go file and parse routes
	routesFile, err := os.ReadFile(path.Join(rootPath, "internal", "server", "routes.go"))
	if err != nil {
		return err
	}
	r.rt.Update(routesFile)
	r.ReloadPkgs()

	r.IsOpen = true
	return nil
}

func (r *Repo) IsGofs() bool {
	return r.HasGofsConfig
}

func (r *Repo) IsValidGofs() bool {
	return r.IsOpen && r.HasGofsConfig && r.Module != "" && r.rt.IsValid()
}

func (r *Repo) GetTemplFile(path string) *templFile.TemplFile {
	if !r.IsOpen {
		return nil
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	for _, templ := range r.ot {
		if templ.Path == path {
			return &templ
		}
	}
	return nil
}

func (r *Repo) GetPkgFunc(pkg, f string) *pkg.Func {
	if !r.IsOpen {
		return nil
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if p, ok := r.pkgs[pkg]; ok {
		return p.GetFunc(f)
	}
	return nil
}

func (r *Repo) OpenTemplFile(req DidOpenRequest) {
	if !r.IsOpen {
		return
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	uris, err := templFile.GetTemplUris(req.TextDocument.Text)
	if err != nil {
		// create an empty entry if the file has parse errors
		r.ot = append(r.ot, templFile.TemplFile{
			Path:           req.TextDocument.Path,
			Text:           req.TextDocument.Text,
			Uris:           []uri.Uri{},
			UrisRouteIndex: []int{},
		})
		return
	}
	uriRouteIndex := make([]int, len(uris))
	for i := range uris {
		uriRouteIndex[i] = r.rt.RouteIndex(uris[i])
		if uriRouteIndex[i] == -1 {
			uris[i].Diag = append(uris[i].Diag, model.Diag{
				Severity: model.SeverityError,
				Message:  "no route found for uri",
			})
		}
	}
	r.ot = append(r.ot, templFile.TemplFile{
		Path:           req.TextDocument.Path,
		Text:           req.TextDocument.Text,
		Uris:           uris,
		UrisRouteIndex: uriRouteIndex,
	})
}

func (r *Repo) ChangeTemplFile(req DidChangeRequest) {
	if !r.IsOpen {
		return
	}

	if len(req.ContentChanges) == 0 {
		return
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	for i, templ := range r.ot {
		if templ.Path == req.TextDocument.Path {
			// TODO handle multiple content changes, this works because client sends entire file
			r.ot[i].Text = req.ContentChanges[0].Text
			uris, err := templFile.GetTemplUris(req.ContentChanges[0].Text)
			if err != nil {
				return
			}
			uriRouteIndex := make([]int, len(uris))
			for j := range uris {
				uriRouteIndex[j] = r.rt.RouteIndex(uris[j])
				if uriRouteIndex[j] == -1 {
					uris[j].Diag = append(uris[j].Diag, model.Diag{
						Severity: model.SeverityError,
						Message:  "no route found for uri",
					})
				}
			}
			r.ot[i].Uris = uris
			r.ot[i].UrisRouteIndex = uriRouteIndex
			return
		}
	}
}

func (r *Repo) CloseTemplFile(req DidCloseRequest) {
	if !r.IsOpen {
		return
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	for i, templ := range r.ot {
		if templ.Path == req.TextDocument.Path {
			r.ot = append(r.ot[:i], r.ot[i+1:]...)
			return
		}
	}
}

func (r *Repo) RecalculateTemplUrls() {
	// can be called if the repo is not open
	for i := range r.ot {
		uris, err := templFile.GetTemplUris(r.ot[i].Text)
		if err != nil {
			log.Printf("error getting templ urls: %v", err.Error())
			return
		}
		uriRouteIndex := make([]int, len(uris))
		for j := range uris {
			uriRouteIndex[j] = r.rt.RouteIndex(uris[j])
			if uriRouteIndex[j] == -1 {
				uris[j].Diag = append(uris[j].Diag, model.Diag{
					Severity: model.SeverityError,
					Message:  "no route found for uri",
				})
			}
		}
		r.ot[i].Uris = uris
		r.ot[i].UrisRouteIndex = uriRouteIndex
	}
}

func (r *Repo) ReloadPkgs() {
	// can be called if the repo is not open
	// clear old pkgs
	r.pkgs = make(map[string]pkg.Pkg)
	for _, route := range r.rt.Routes() {
		if route.Pkg == "" {
			continue
		}
		if _, ok := r.pkgs[route.Pkg]; !ok {
			pkg, err := pkg.GetPkg(route.Pkg)
			if err != nil {
				continue
			}
			r.pkgs[route.Pkg] = *pkg
		}
	}
}

func (r *Repo) UpdateRoutes(b []byte) {
	if !r.IsOpen {
		return
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.rt.Update(b)
	r.RecalculateTemplUrls()
	r.ReloadPkgs()
}

func (r *Repo) Routes() []routesFile.Route {
	if !r.IsOpen {
		return nil
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	return r.rt.Routes()
}

func (r *Repo) GetRoute(index int) (*routesFile.Route, error) {
	if !r.IsOpen {
		return nil, errors.New("repo is not open")
	}

	return r.rt.GetRoute(index)
}
