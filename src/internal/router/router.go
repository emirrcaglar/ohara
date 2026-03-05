package router

import (
	"net/http"

	"ohara/src/internal/db"
	"ohara/src/internal/handler"
	"ohara/src/ui"
)

func SetupRoutes(baseDir string, database *db.DB) http.Handler {
	mux := http.NewServeMux()

	mangaHandler := &handler.MangaHandler{BaseDir: baseDir, DB: database}

	mux.Handle("GET /static/", http.FileServer(http.FS(ui.Files)))

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		data, _ := ui.Files.ReadFile("index.html")
		w.Header().Set("Content-Type", "text/html")
		w.Write(data)
	})

	mux.HandleFunc("GET /library", mangaHandler.HandleMangaList)

	mux.HandleFunc("GET /manga/{name}/page/{page}", mangaHandler.HandleMangaPage)
	mux.HandleFunc("GET /manga/{name}/snippet/{page}", mangaHandler.HandleMangaSnippet)

	mux.HandleFunc("GET /manga/{name}/reader/", mangaHandler.HandleMangaReader)
	mux.HandleFunc("GET /manga/{name}/reader/{page}", mangaHandler.HandleMangaReader)

	mux.HandleFunc("GET /manga/{name}/info", mangaHandler.HandleMangaInfo)

	return mux
}
