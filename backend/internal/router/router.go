package router

import (
	"context"
	"net/http"

	"ohara/src/internal/db"
	"ohara/src/internal/handler"
	"ohara/src/internal/logger"
	"ohara/src/internal/media/cbz"
	"ohara/src/internal/scanner"
	"ohara/src/ui"
)

func WithAuth(database *db.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		user, err := database.GetUserByUsername(cookie.Value)
		if err != nil || !user.IsApproved {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), handler.UserKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func WithRole(role string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := handler.GetUser(r.Context())
		if user == nil || user.Role != role {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func SetupRoutes(database *db.DB, dataDir string, log *logger.Logger) http.Handler {
	mux := http.NewServeMux()

	cbzService := cbz.NewCBZService(database)
	scanner := scanner.NewScanner(database, cbzService)
	mangaHandler := &handler.MangaHandler{DB: database, Cache: handler.NewPageCache(dataDir), Inflight: handler.NewInflight(), CBZService: cbzService}
	audioHandler := &handler.AudioHandler{DB: database}
	uploadHandler := handler.NewUploadHandler(database, scanner)
	logHandler := &handler.LogHandler{Logger: log}
	authHandler := &handler.AuthHandler{DB: database}
	adminHandler := &handler.AdminHandler{DB: database}

	mux.HandleFunc("POST /api/auth/login", authHandler.HandleLogin)
	mux.HandleFunc("POST /api/auth/register", authHandler.HandleRegister)
	mux.HandleFunc("POST /api/auth/logout", authHandler.HandleLogout)
	mux.HandleFunc("GET /api/auth/me", authHandler.HandleMe)

	// Admin routes
	mux.HandleFunc("GET /api/admin/users/pending", WithAuth(database, WithRole("admin", adminHandler.HandleListPendingUsers)))
	mux.HandleFunc("POST /api/admin/users/{id}/approve", WithAuth(database, WithRole("admin", adminHandler.HandleApproveUser)))

	mux.HandleFunc("GET /api/manga", WithAuth(database, mangaHandler.HandleMangaList))
	mux.HandleFunc("GET /api/audio", WithAuth(database, audioHandler.HandleAudioList))

	mux.HandleFunc("GET /api/manga/{id}/resume", WithAuth(database, mangaHandler.HandleMangaResume))
	mux.HandleFunc("GET /api/manga/{id}/page/{page}", WithAuth(database, mangaHandler.HandleMangaPage))
	mux.HandleFunc("POST /api/manga/{id}/progress/{page}", WithAuth(database, mangaHandler.HandleMangaProgress))
	mux.HandleFunc("GET /api/manga/{id}/info", WithAuth(database, mangaHandler.HandleMangaInfo))

	mux.HandleFunc("GET /audio/{id}/stream", WithAuth(database, audioHandler.HandleAudioStream))

	mux.HandleFunc("POST /api/upload", WithAuth(database, uploadHandler.HandleUpload))

	mux.HandleFunc("GET /api/logs", WithAuth(database, logHandler.HandleSnapshot))
	mux.HandleFunc("GET /api/logs/stream", WithAuth(database, logHandler.HandleStream))

	if spaHandler, err := ui.SPAHandler(); err == nil {
		mux.Handle("/", spaHandler)
	}

	return mux
}
