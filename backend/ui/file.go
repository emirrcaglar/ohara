package ui

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

//go:embed *
var Files embed.FS

func SPAHandler() (http.Handler, error) {
	distFS, err := fs.Sub(Files, "dist")
	if err != nil {
		return nil, err
	}

	fileServer := http.FileServer(http.FS(distFS))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestPath := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
		if requestPath == "" || requestPath == "." {
			requestPath = "index.html"
		}

		file, err := distFS.Open(requestPath)
		if err == nil {
			file.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		fallbackRequest := r.Clone(r.Context())
		fallbackRequest.URL.Path = "/"
		fileServer.ServeHTTP(w, fallbackRequest)
	}), nil
}
