package router

import (
	"net/http"

	"ohara/src/internal/db"
	"ohara/src/internal/handler"
	"ohara/src/ui"
)

func SetupRoutes(database *db.DB) http.Handler {
	mux := http.NewServeMux()

	mangaHandler := &handler.MangaHandler{DB: database}

	mux.Handle("GET /static/", http.FileServer(http.FS(ui.Files)))

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		data, _ := ui.Files.ReadFile("index.html")
		w.Header().Set("Content-Type", "text/html")
		w.Write(data)
	})

	mux.HandleFunc("GET /library", mangaHandler.HandleMangaList)

	mux.HandleFunc("GET /manga/{id}/resume", mangaHandler.HandleMangaResume)
	mux.HandleFunc("GET /manga/{id}/page/{page}", mangaHandler.HandleMangaPage)
	mux.HandleFunc("GET /manga/{id}/snippet/{page}", mangaHandler.HandleMangaSnippet)
	mux.HandleFunc("GET /manga/{id}/info", mangaHandler.HandleMangaInfo)

	return mux
}
