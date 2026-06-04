package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"ohara/src/internal/db"
	"ohara/src/internal/logger"
)

type VideoHandler struct {
	DB  *db.DB
	Log *logger.Logger
}

type VideoResponse struct {
	ID            int64  `json:"id"`
	CatalogID     *int64 `json:"catalogId"`
	Title         string `json:"title"`
	Duration      int    `json:"duration"`
	Width         int    `json:"width"`
	Height        int    `json:"height"`
	Position      int    `json:"position"`
	Completed     bool   `json:"completed"`
	LastError     string `json:"lastError"`
	FileExtension string `json:"fileExtension"`
}

type VideoDetailResponse struct {
	ID            int64  `json:"id"`
	CatalogID     *int64 `json:"catalogId"`
	Title         string `json:"title"`
	Path          string `json:"path"`
	Duration      int    `json:"duration"`
	Width         int    `json:"width"`
	Height        int    `json:"height"`
	Position      int    `json:"position"`
	Completed     bool   `json:"completed"`
	LastError     string `json:"lastError"`
	FileExtension string `json:"fileExtension"`
}

type VideoStateRequest struct {
	Duration  int    `json:"duration"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Position  int    `json:"position"`
	Completed bool   `json:"completed"`
	LastError string `json:"lastError"`
}

type VideoLibraryResponse struct {
	Items []VideoResponse `json:"items"`
	Total int             `json:"total"`
}

func videoResponse(video db.VideoRow) VideoResponse {
	return VideoResponse{
		ID:            video.ID,
		CatalogID:     video.CatalogID,
		Title:         video.Title,
		Duration:      video.Duration,
		Width:         video.Width,
		Height:        video.Height,
		Position:      video.Position,
		Completed:     video.Completed,
		LastError:     video.LastError,
		FileExtension: filepath.Ext(video.Path),
	}
}

func (h *VideoHandler) videoByID(r *http.Request, idStr string) (*db.VideoRow, int) {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, http.StatusBadRequest
	}

	user := GetUser(r.Context())
	if user == nil {
		return nil, http.StatusUnauthorized
	}

	video, err := h.DB.GetVideoByID(user.ID, id)
	if err != nil {
		return nil, http.StatusInternalServerError
	}
	if video == nil {
		return nil, http.StatusNotFound
	}

	return video, 0
}

func (h *VideoHandler) HandleVideoList(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	videos, err := h.DB.ListVideo(user.ID)
	if err != nil {
		http.Error(w, "Failed to load video library", http.StatusInternalServerError)
		return
	}

	items := make([]VideoResponse, 0, len(videos))
	for _, video := range videos {
		items = append(items, videoResponse(video))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(VideoLibraryResponse{Items: items, Total: len(items)})
}

func (h *VideoHandler) HandleVideoInfo(w http.ResponseWriter, r *http.Request) {
	video, status := h.videoByID(r, r.PathValue("id"))
	if status != 0 {
		http.Error(w, "Video not found", status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(VideoDetailResponse{
		ID:            video.ID,
		CatalogID:     video.CatalogID,
		Title:         video.Title,
		Path:          video.Path,
		Duration:      video.Duration,
		Width:         video.Width,
		Height:        video.Height,
		Position:      video.Position,
		Completed:     video.Completed,
		LastError:     video.LastError,
		FileExtension: filepath.Ext(video.Path),
	})
}

func (h *VideoHandler) HandleVideoState(w http.ResponseWriter, r *http.Request) {
	video, status := h.videoByID(r, r.PathValue("id"))
	if status != 0 {
		http.Error(w, "Video not found", status)
		return
	}

	user := GetUser(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req VideoStateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid video state", http.StatusBadRequest)
		return
	}

	if req.Duration < 0 {
		req.Duration = 0
	}
	if req.Width < 0 {
		req.Width = 0
	}
	if req.Height < 0 {
		req.Height = 0
	}
	if req.Position < 0 {
		req.Position = 0
	}
	if len(req.LastError) > 500 {
		req.LastError = req.LastError[:500]
	}

	if err := h.DB.UpdateVideoMetadata(video.ID, req.Duration, req.Width, req.Height); err != nil {
		http.Error(w, "Failed to save video metadata", http.StatusInternalServerError)
		return
	}
	if err := h.DB.UpsertVideoProgress(user.ID, video.ID, req.Position, req.Completed, req.LastError); err != nil {
		http.Error(w, "Failed to save video progress", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func videoContentType(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".mp4", ".m4v":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".mov":
		return "video/quicktime"
	case ".avi":
		return "video/x-msvideo"
	case ".mkv":
		return "video/x-matroska"
	default:
		return "application/octet-stream"
	}
}

func (h *VideoHandler) HandleVideoStream(w http.ResponseWriter, r *http.Request) {
	video, status := h.videoByID(r, r.PathValue("id"))
	if status != 0 {
		http.Error(w, "Video not found", status)
		return
	}

	file, err := os.Open(video.Path)
	if err != nil {
		if h.Log != nil {
			h.Log.Error("[video] failed to open file path=%s err=%v", video.Path, err)
		}
		http.Error(w, "File not found on disk", http.StatusNotFound)
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		http.Error(w, "Could not stat file", http.StatusInternalServerError)
		return
	}

	r.Header.Del("If-Modified-Since")
	r.Header.Del("If-None-Match")
	r.Header.Del("If-Match")
	r.Header.Del("If-Unmodified-Since")

	w.Header().Set("Content-Type", videoContentType(video.Path))
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Accept-Ranges", "bytes")
	http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
}

func (h *VideoHandler) HandleVideoCatalogUpdate(w http.ResponseWriter, r *http.Request) {
	video, status := h.videoByID(r, r.PathValue("id"))
	if status != 0 {
		http.Error(w, "Video not found", status)
		return
	}

	var req MediaCatalogUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if !validMediaCatalog(w, h.DB, req.CatalogID) {
		return
	}

	if err := h.DB.UpdateVideoCatalog(video.ID, req.CatalogID); err != nil {
		if h.Log != nil {
			h.Log.Error("[video] catalog update failed video=%d catalog=%v err=%v", video.ID, req.CatalogID, err)
		}
		http.Error(w, "Failed to move video", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *VideoHandler) HandleVideoDelete(w http.ResponseWriter, r *http.Request) {
	video, status := h.videoByID(r, r.PathValue("id"))
	if status != 0 {
		http.Error(w, "Video not found", status)
		return
	}

	if err := h.DB.DeleteVideo(video.ID); err != nil {
		if h.Log != nil {
			h.Log.Error("[video] delete failed video=%d err=%v", video.ID, err)
		}
		http.Error(w, "Failed to delete video", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
