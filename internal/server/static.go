package server

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed ui/dist/*
var uiFS embed.FS

func handleUI(w http.ResponseWriter, r *http.Request) {
	sub, err := fs.Sub(uiFS, "ui/dist")
	if err != nil {
		http.Error(w, "internal server error: ui missing", http.StatusInternalServerError)
		return
	}

	path := r.URL.Path
	cleanPath := strings.TrimPrefix(path, "/")
	if cleanPath == "" {
		cleanPath = "."
	}

	f, err := sub.Open(cleanPath)
	if err != nil {
		// File does not exist, SPA fallback to index.html
		r.URL.Path = "/"
	} else {
		// File exists, but if it's a directory, let FileServer handle it normally
		f.Close()
	}

	http.FileServer(http.FS(sub)).ServeHTTP(w, r)
}
