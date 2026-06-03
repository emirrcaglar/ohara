package router

import (
	"context"
	"net/http"
	"time"

	"ohara/src/internal/db"
	"ohara/src/internal/handler"
	"ohara/src/internal/logger"
	"ohara/src/internal/media/cbz"
	"ohara/src/internal/scanner"
	"ohara/src/internal/worker"
	"ohara/src/ui"

	_ "net/http/pprof"
)

func WithAuth(database *db.DB, log *logger.Logger, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err != nil {
			if log != nil {
				log.Warn("[http] unauthorized path=%s reason=missing_session_cookie", r.URL.Path)
			}
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		user, err := database.GetUserByUsername(cookie.Value)
		if err != nil || !user.IsApproved {
			if log != nil {
				log.Warn("[http] unauthorized path=%s reason=invalid_or_unapproved_user username=%s", r.URL.Path, cookie.Value)
			}
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), handler.UserKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func WithRole(role string, log *logger.Logger, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := handler.GetUser(r.Context())
		if user == nil || user.Role != role {
			if log != nil {
				username := "unknown"
				if user != nil {
					username = user.Username
				}
				log.Warn("[http] forbidden path=%s required_role=%s username=%s", r.URL.Path, role, username)
			}
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func SetupRoutes(database *db.DB, dataDir string, log *logger.Logger) http.Handler {
	mux := http.DefaultServeMux

	cbzService := cbz.NewCBZService(database)
	mangaHandler := &handler.MangaHandler{DB: database, Cache: handler.NewPageCache(dataDir, database), Inflight: handler.NewInflight(), CBZService: cbzService, Log: log}
	audioHandler := &handler.AudioHandler{DB: database, Log: log}
	videoHandler := &handler.VideoHandler{DB: database, Log: log}
	logHandler := &handler.LogHandler{Logger: log}
	authHandler := &handler.AuthHandler{DB: database, Log: log}
	adminHandler := &handler.AdminHandler{DB: database, Log: log}
	preferencesHandler := &handler.PreferencesHandler{DB: database, Log: log}
	deploymentHandler := &handler.DeploymentHandler{DB: database, Log: log}

	cacheWorker := worker.NewCacheWorker(dataDir, database, *cbzService)
	cacheWorker.Start()

	scanner := scanner.NewScanner(database, cbzService, cacheWorker, log)
	uploadHandler := handler.NewUploadHandler(database, scanner, log, dataDir)

	mux.HandleFunc("POST /api/auth/login", authHandler.HandleLogin)
	mux.HandleFunc("POST /api/auth/register", authHandler.HandleRegister)
	mux.HandleFunc("POST /api/auth/logout", authHandler.HandleLogout)
	mux.HandleFunc("GET /api/auth/me", authHandler.HandleMe)

	// Admin routes
	mux.HandleFunc("GET /api/admin/users/pending", WithAuth(database, log, WithRole("admin", log, adminHandler.HandleListPendingUsers)))
	mux.HandleFunc("POST /api/admin/users/{id}/approve", WithAuth(database, log, WithRole("admin", log, adminHandler.HandleApproveUser)))
	mux.HandleFunc("POST /api/admin/users/{id}/reject", WithAuth(database, log, WithRole("admin", log, adminHandler.HandleRejectUser)))

	mux.HandleFunc("GET /api/manga", WithAuth(database, log, mangaHandler.HandleMangaList))
	mux.HandleFunc("GET /api/audio", WithAuth(database, log, audioHandler.HandleAudioList))
	mux.HandleFunc("GET /api/video", WithAuth(database, log, videoHandler.HandleVideoList))

	mux.HandleFunc("GET /api/preferences", WithAuth(database, log, preferencesHandler.HandlePreferencesList))
	mux.HandleFunc("GET /api/preferences/{key}", WithAuth(database, log, preferencesHandler.HandlePreferenceGet))
	mux.HandleFunc("PUT /api/preferences/{key}", WithAuth(database, log, preferencesHandler.HandlePreferenceUpsert))
	mux.HandleFunc("DELETE /api/preferences/{key}", WithAuth(database, log, preferencesHandler.HandlePreferenceDelete))

	mux.HandleFunc("DELETE /api/manga/{id}", WithAuth(database, log, mangaHandler.HandleMangaDelete))
	mux.HandleFunc("GET /api/manga/{id}/resume", WithAuth(database, log, mangaHandler.HandleMangaResume))
	mux.HandleFunc("GET /api/manga/{id}/page/{page}", WithAuth(database, log, mangaHandler.HandleMangaPage))
	mux.HandleFunc("POST /api/manga/{id}/progress/{page}", WithAuth(database, log, mangaHandler.HandleMangaProgress))
	mux.HandleFunc("GET /api/manga/{id}/info", WithAuth(database, log, mangaHandler.HandleMangaInfo))

	mux.HandleFunc("GET /api/video/{id}/info", WithAuth(database, log, videoHandler.HandleVideoInfo))
	mux.HandleFunc("PUT /api/video/{id}/state", WithAuth(database, log, videoHandler.HandleVideoState))
	mux.HandleFunc("DELETE /api/video/{id}", WithAuth(database, log, videoHandler.HandleVideoDelete))

	mux.HandleFunc("GET /audio/{id}/stream", WithAuth(database, log, audioHandler.HandleAudioStream))
	mux.HandleFunc("GET /video/{id}/stream", WithAuth(database, log, videoHandler.HandleVideoStream))

	mux.HandleFunc("GET /api/uploads", WithAuth(database, log, uploadHandler.HandleUploadsList))
	mux.HandleFunc("POST /api/uploads/init", WithAuth(database, log, uploadHandler.HandleChunkedUploadInit))
	mux.HandleFunc("PUT /api/uploads/{id}/chunk", WithAuth(database, log, uploadHandler.HandleChunkedUploadChunk))
	mux.HandleFunc("POST /api/uploads/{id}/complete", WithAuth(database, log, uploadHandler.HandleChunkedUploadComplete))
	mux.HandleFunc("GET /api/uploads/{id}/status", WithAuth(database, log, uploadHandler.HandleChunkedUploadStatus))
	mux.HandleFunc("POST /api/uploads/{id}/pause", WithAuth(database, log, uploadHandler.HandleChunkedUploadPause))
	mux.HandleFunc("DELETE /api/uploads/{id}", WithAuth(database, log, uploadHandler.HandleChunkedUploadCancel))

	mux.HandleFunc("GET /api/logs", WithAuth(database, log, logHandler.HandleSnapshot))
	mux.HandleFunc("GET /api/logs/stream", WithAuth(database, log, logHandler.HandleStream))
	mux.HandleFunc("GET /api/deployments/latest", WithAuth(database, log, deploymentHandler.HandleLatest))

	if spaHandler, err := ui.SPAHandler(); err == nil {
		mux.Handle("/", spaHandler)
	}

	return withRequestLogging(log, mux)
}

func withRequestLogging(log *logger.Logger, next http.Handler) http.Handler {
	if log == nil {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()
		next.ServeHTTP(recorder, r)
		log.Info("[http] %s %s status=%d duration=%s", r.Method, r.URL.Path, recorder.status, time.Since(start))
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Flush() {
	if flusher, ok := r.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}
