package web

import (
	"embed"
	"io/fs"
	"net/http"
)

// Static files are embedded in the binary
//go:embed frontend/dist/*
var StaticFS embed.FS

// Handler returns an http.FileSystem for serving embedded static files
func Handler() (http.FileSystem, error) {
	sub, err := fs.Sub(StaticFS, "frontend/dist")
	if err != nil {
		return nil, err
	}
	return http.FS(sub), nil
}
