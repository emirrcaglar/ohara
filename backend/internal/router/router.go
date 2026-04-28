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
	audioHandler := &handler.AudioHandler{DB: database}

	mux.HandleFunc("GET /api/manga", mangaHandler.HandleMangaList)
	mux.HandleFunc("GET /api/audio", audioHandler.HandleAudioList)

	mux.HandleFunc("GET /api/manga/{id}/resume", mangaHandler.HandleMangaResume)
	mux.HandleFunc("GET /api/manga/{id}/page/{page}", mangaHandler.HandleMangaPage)
	mux.HandleFunc("POST /api/manga/{id}/progress/{page}", mangaHandler.HandleMangaProgress)
	mux.HandleFunc("GET /api/manga/{id}/info", mangaHandler.HandleMangaInfo)

	mux.HandleFunc("GET /audio/{id}/stream", audioHandler.HandleAudioStream)

	if spaHandler, err := ui.SPAHandler(); err == nil {
		mux.Handle("/", spaHandler)
	}

	return mux
}
