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
	Title         string `json:"title"`
	Duration      int    `json:"duration"`
	FileExtension string `json:"fileExtension"`
}

type VideoDetailResponse struct {
	ID            int64  `json:"id"`
	Title         string `json:"title"`
	Path          string `json:"path"`
	Duration      int    `json:"duration"`
	FileExtension string `json:"fileExtension"`
}

type VideoLibraryResponse struct {
	Items []VideoResponse `json:"items"`
	Total int             `json:"total"`
}

func (h *VideoHandler) videoByID(idStr string) (*db.VideoRow, int) {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, http.StatusBadRequest
	}

	video, err := h.DB.GetVideoByID(id)
	if err != nil {
		return nil, http.StatusInternalServerError
	}
	if video == nil {
		return nil, http.StatusNotFound
	}

	return video, 0
}

func (h *VideoHandler) HandleVideoList(w http.ResponseWriter, r *http.Request) {
	videos, err := h.DB.ListVideo()
	if err != nil {
		http.Error(w, "Failed to load video library", http.StatusInternalServerError)
		return
	}

	items := make([]VideoResponse, 0, len(videos))
	for _, video := range videos {
		items = append(items, VideoResponse{
			ID:            video.ID,
			Title:         video.Title,
			Duration:      video.Duration,
			FileExtension: filepath.Ext(video.Path),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(VideoLibraryResponse{Items: items, Total: len(items)})
}

func (h *VideoHandler) HandleVideoInfo(w http.ResponseWriter, r *http.Request) {
	video, status := h.videoByID(r.PathValue("id"))
	if status != 0 {
		http.Error(w, "Video not found", status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(VideoDetailResponse{
		ID:            video.ID,
		Title:         video.Title,
		Path:          video.Path,
		Duration:      video.Duration,
		FileExtension: filepath.Ext(video.Path),
	})
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
	video, status := h.videoByID(r.PathValue("id"))
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

func (h *VideoHandler) HandleVideoDelete(w http.ResponseWriter, r *http.Request) {
	video, status := h.videoByID(r.PathValue("id"))
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
