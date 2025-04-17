package repo

import (
	"errors"
	"log/slog"
	"os"
	"path"
	"sync"
	"time"

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
	rt            routesFile.Routes  // routes file
	ot            sync.Map           // open templ files
	pkgs          map[string]pkg.Pkg // loaded packages
	shouldLoad    sync.Map           // files that are being loaded or have loaded
}

func NewRepo() *Repo {
	// open must be called first and creates the repo
	return &Repo{
		IsOpen:     false,
		ot:         sync.Map{},
		shouldLoad: sync.Map{},
	}
}

func (r *Repo) Open(rootPath string) error {
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
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	} else if errors.Is(err, os.ErrNotExist) {
		r.rt.SetDefault()
	} else {
		r.rt.Update(routesFile)
		r.ReloadPkgs()
	}

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
	for {
		// TODO add a counter to abort after n attempts
		if templ, ok := r.ot.Load(path); ok {
			t := templ.(templFile.TemplFile)
			return &t
		}
		if _, ok := r.shouldLoad.Load(path); ok {
			// the file is loading so wait
			time.Sleep(100 * time.Millisecond)
		} else {
			// a request to load the file was not sent
			return nil
		}
	}
}

func (r *Repo) GetPkgFunc(pkg, f string) *pkg.Func {
	if p, ok := r.pkgs[pkg]; ok {
		return p.GetFunc(f)
	}
	return nil
}

func (r *Repo) OpenTemplFile(req DidOpenRequest) {
	r.shouldLoad.Store(req.TextDocument.Path, true)
	uris, err := templFile.GetTemplUris(req.TextDocument.Text)
	if err != nil {
		// create an empty entry if the file has parse errors
		r.ot.Store(req.TextDocument.Path, templFile.TemplFile{
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

	r.ot.Store(req.TextDocument.Path, templFile.TemplFile{
		Path:           req.TextDocument.Path,
		Text:           req.TextDocument.Text,
		Uris:           uris,
		UrisRouteIndex: uriRouteIndex,
	})
}

func (r *Repo) ChangeTemplFile(req DidChangeRequest) {
	if len(req.ContentChanges) == 0 {
		return
	}

	t := r.GetTemplFile(req.TextDocument.Path)
	if t == nil {
		return
	}
	uris, err := templFile.GetTemplUris(req.ContentChanges[0].Text)
	if err != nil {
		return
	}
	// TODO handle multiple content changes, this works because client sends entire file
	t.Text = req.ContentChanges[0].Text
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
	t.Uris = uris
	t.UrisRouteIndex = uriRouteIndex
	r.ot.Store(req.TextDocument.Path, *t)
}

func (r *Repo) CloseTemplFile(req DidCloseRequest) {
	r.shouldLoad.Delete(req.TextDocument.Path)
	r.ot.Delete(req.TextDocument.Path)
}

func (r *Repo) RecalculateTemplUrls() {
	m := sync.Map{}
	// can be called if the repo is not open
	r.ot.Range(func(key, value any) bool {
		t := value.(templFile.TemplFile)
		uris, err := templFile.GetTemplUris(t.Text)
		if err != nil {
			slog.Error("error getting templ urls", "err", err)
			return true
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
		t.Uris = uris
		t.UrisRouteIndex = uriRouteIndex
		m.Store(key, t)
		return true
	})
	r.ot.Clear()
	m.Range(func(key, value any) bool {
		r.ot.Store(key, value)
		return true
	})
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

	r.rt.Update(b)
	r.RecalculateTemplUrls()
	r.ReloadPkgs()
}

func (r *Repo) Routes() []routesFile.Route {
	return r.rt.Routes()
}

func (r *Repo) GetRoute(index int) (*routesFile.Route, error) {
	if !r.IsOpen {
		return nil, errors.New("repo is not open")
	}

	return r.rt.GetRoute(index)
}
