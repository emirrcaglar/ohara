package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"ohara/src/internal/reader"
)

func (s *Server) HandleViewPage(w http.ResponseWriter, r *http.Request) {
	pageStr := r.PathValue("page")
	mangaName := r.PathValue("name")
	fullPath := filepath.Join(s.BaseDir, mangaName+".cbz")

	pageIdx, err := strconv.Atoi(pageStr)
	if err != nil || pageIdx < 0 {
		http.Error(w, "Invalid page number", http.StatusBadRequest)
		return
	}

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		http.Error(w, "Manga not found", http.StatusNotFound)
		return
	}

	manga, err := reader.Open(fullPath)
	if err != nil {
		fmt.Printf("Error opening CBZ: %v\n", err)
		http.Error(w, "Could not open manga file", http.StatusInternalServerError)
		return
	}
	defer manga.Close()

	rc, err := manga.GetPageReader(pageIdx)
	if err != nil {
		// Distinguish between "page out of bounds" and "read error" if possible
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}
	defer rc.Close()

	w.Header().Set("Content-Type", "image/jpeg")
	// Cache-Control is important for images so the browser doesn't re-request instantly
	w.Header().Set("Cache-Control", "public, max-age=3600")

	if _, err := io.Copy(w, rc); err != nil {
		fmt.Printf("Stream error: %v\n", err)
	}
}

func (s *Server) HandleMangaInfo(w http.ResponseWriter, r *http.Request) {
	mangaName := r.PathValue("name")
	fullPath := filepath.Join(s.BaseDir, mangaName+".cbz")

	manga, err := reader.Open(fullPath)
	if err != nil {
		fmt.Printf("Error opening CBZ: %v\n", err)
		http.Error(w, "Could not open manga file", http.StatusInternalServerError)
		return
	}
	defer manga.Close()

	m, err := json.MarshalIndent(manga, "", "  ")
	if err != nil {
		w.Write([]byte(err.Error()))
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(m)
}
