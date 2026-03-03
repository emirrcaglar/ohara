package server

import (
	"net/http"

	"ohara/src/internal/handler"
)

func New(baseDir string) http.Handler {
	mux := http.NewServeMux()

	mangaHandler := &handler.MangaHandler{BaseDir: baseDir}

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})

	mux.HandleFunc("GET /manga/{name}/page/{page}", mangaHandler.HandleMangaPage)
	mux.HandleFunc("GET /manga/{name}/snippet/{page}", mangaHandler.HandleMangaSnippet)

	mux.HandleFunc("GET /manga/{name}/reader/", mangaHandler.HandleMangaReader)
	mux.HandleFunc("GET /manga/{name}/reader/{page}", mangaHandler.HandleMangaReader)

	mux.HandleFunc("GET /manga/{name}/info", mangaHandler.HandleMangaInfo)

	return mux
}

type Server struct {
	BaseDir string
}
