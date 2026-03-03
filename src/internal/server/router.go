package server

import (
	"net/http"

	"ohara/src/internal/handler"
	"ohara/src/web"
)

func New(baseDir string) http.Handler {
	mux := http.NewServeMux()

	mangaHandler := &handler.MangaHandler{BaseDir: baseDir}

	// Serve embedded web files
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(web.Files))))

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		data, _ := web.Files.ReadFile("index.html")
		w.Header().Set("Content-Type", "text/html")
		w.Write(data)
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
