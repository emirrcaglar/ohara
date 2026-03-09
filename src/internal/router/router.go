package router

import (
	"net/http"

	"ohara/src/internal/db"
	"ohara/src/internal/handler"
	"ohara/src/ui"
)

func SetupRoutes(database *db.DB, dataDir string) http.Handler {
	mux := http.NewServeMux()

	mangaHandler := &handler.MangaHandler{DB: database, Cache: handler.NewPageCache(dataDir), Inflight: handler.NewInflight()}

	mux.Handle("GET /static/", http.FileServer(http.FS(ui.Files)))

	mux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		data, _ := ui.Files.ReadFile("home.html")
		w.Header().Set("Content-Type", "text/html")
		w.Write(data)
	})

	mux.HandleFunc("GET /reader", func(w http.ResponseWriter, r *http.Request) {
		data, _ := ui.Files.ReadFile("index.html")
		w.Header().Set("Content-Type", "text/html")
		w.Write(data)
	})

	mux.HandleFunc("GET /manga/library", mangaHandler.HandleMangaList)

	mux.HandleFunc("GET /manga/{id}/resume", mangaHandler.HandleMangaResume)
	mux.HandleFunc("GET /manga/{id}/page/{page}", mangaHandler.HandleMangaPage)
	mux.HandleFunc("POST /manga/{id}/progress/{page}", mangaHandler.HandleMangaProgress)
	mux.HandleFunc("GET /manga/{id}/info", mangaHandler.HandleMangaInfo)

	return mux
}
