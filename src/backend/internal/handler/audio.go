package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"ohara/src/internal/db"
	"ohara/src/internal/logger"
)

type AudioHandler struct {
	DB  *db.DB
	Log *logger.Logger
}

func (h *AudioHandler) HandleAudioList(w http.ResponseWriter, r *http.Request) {
	tracks, err := h.DB.ListAudio()
	if err != nil {
		http.Error(w, "Failed to load audio library", http.StatusInternalServerError)
		return
	}

	items := make([]AudioResponse, 0, len(tracks))
	for _, t := range tracks {
		items = append(items, AudioResponse{
			ID:            t.ID,
			Title:         t.Title,
			Artist:        t.Artist,
			Album:         t.Album,
			Duration:      t.Duration,
			FileExtension: filepath.Ext(t.Path),
		})
	}

	response := AudioLibraryResponse{
		Items: items,
		Total: len(items),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type AudioResponse struct {
	ID            int64  `json:"id"`
	Title         string `json:"title"`
	Artist        string `json:"artist"`
	Album         string `json:"album"`
	Duration      int    `json:"duration"`
	FileExtension string `json:"fileExtension"`
}

type AudioLibraryResponse struct {
	Items []AudioResponse `json:"items"`
	Total int             `json:"total"`
}

func (h *AudioHandler) HandleAudioStream(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	track, err := h.DB.GetAudioByID(id)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if track == nil {
		http.Error(w, "Track not found", http.StatusNotFound)
		return
	}

	file, err := os.Open(track.Path)
	if err != nil {
		if h.Log != nil {
			h.Log.Error("[audio] failed to open file path=%s err=%v", track.Path, err)
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

	http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
}
