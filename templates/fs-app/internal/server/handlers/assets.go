package handlers

import (
	"embed"
	"fmt"
	"hash/crc32"
	"io"
	"net/http"
	"sync"
)

type HashedAssets struct {
	assets embed.FS
	hashes sync.Map
}

func NewHashedAssets(assets embed.FS) *HashedAssets {
	return &HashedAssets{assets: assets}
}

func (h *HashedAssets) calculateHash(path string) (string, error) {
	hasher := crc32.NewIEEE()
	file, err := h.assets.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return fmt.Sprintf(`"%d-%v"`, len(path), hasher.Sum32()), nil
}

func (h *HashedAssets) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	etag, ok := h.hashes.Load(r.URL.Path)
	if !ok {
		var err error
		etag, err = h.calculateHash(r.URL.Path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		h.hashes.Store(r.URL.Path, etag)
	}

	ifNoneMatch := r.Header.Get("If-None-Match")

	etagString, ok := etag.(string)
	if !ok {
		http.Error(w, "etag is not a string", http.StatusInternalServerError)
		return
	}

	if ifNoneMatch == etagString {
		w.WriteHeader(http.StatusNotModified)
	} else {
		w.Header().Add("ETag", etagString)
		http.FileServerFS(h.assets).ServeHTTP(w, r)
	}
}
