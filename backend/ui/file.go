package ui

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

//go:embed all:dist
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
			setStaticCacheHeaders(w, requestPath)
			fileServer.ServeHTTP(w, r)
			return
		}

		if isStaticAssetRequest(requestPath) {
			w.Header().Set("Cache-Control", "no-store")
			http.NotFound(w, r)
			return
		}

		setIndexCacheHeaders(w)
		fallbackRequest := r.Clone(r.Context())
		fallbackRequest.URL.Path = "/"
		fileServer.ServeHTTP(w, fallbackRequest)
	}), nil
}

func isStaticAssetRequest(requestPath string) bool {
	return strings.HasPrefix(requestPath, "assets/") || strings.Contains(path.Base(requestPath), ".")
}

func setStaticCacheHeaders(w http.ResponseWriter, requestPath string) {
	if requestPath == "index.html" {
		setIndexCacheHeaders(w)
		return
	}

	if strings.HasPrefix(requestPath, "assets/") {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	}
}

func setIndexCacheHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Clear-Site-Data", `"cache"`)
}
