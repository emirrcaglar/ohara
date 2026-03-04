package router

import (
	"net/http"

	"ohara/src/internal/handler"
	"ohara/src/ui"
)

func SetupRoutes(baseDir string) http.Handler {
	mux := http.NewServeMux()

	mangaHandler := &handler.MangaHandler{BaseDir: baseDir}

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(ui.Files))))

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		data, _ := ui.Files.ReadFile("index.html")
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

type Router struct {
	// bundan kurtul
	BaseDir string
}
