package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	cbzReader "ohara/src/internal/media/cbz"
)

type MangaHandler struct {
	BaseDir string
}

func (h *MangaHandler) HandleMangaList(w http.ResponseWriter, r *http.Request) {
	matches, _ := filepath.Glob(filepath.Join(h.BaseDir, "*.cbz"))

	var cards strings.Builder
	for _, path := range matches {
		name := strings.TrimSuffix(filepath.Base(path), ".cbz")
		nameQ := url.QueryEscape(name)
		nameP := url.PathEscape(name)
		cards.WriteString(fmt.Sprintf(`
		<a class="manga-card" href="/?manga=%s&page=0">
			<img src="/manga/%s/page/0" alt="%s" loading="lazy">
			<span>%s</span>
		</a>`, nameQ, nameP, name, name))
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
		<html lang="en">
			<head>
			<meta charset="UTF-8">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
				<title>Library - Ohara</title>
				<link rel="stylesheet" href="/static/style.css">
				<style>
					body { padding: 20px; }
					.grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(160px, 1fr)); gap: 16px; }
					.manga-card { display: flex; flex-direction: column; align-items: center; text-decoration: none; color: white; background: #1e1e1e; border-radius: 8px; overflow: hidden; transition: transform 0.15s; }
					.manga-card:hover { transform: scale(1.04); }
					.manga-card img { width: 100%%; aspect-ratio: 2/3; object-fit: cover; background: #333; }
					.manga-card span { padding: 8px; font-size: 0.85rem; text-align: center; word-break: break-word; }
				</style>
			</head>
			<body>
				<div class="grid">%s</div>
			</body>
		</html>`, cards.String())

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func (h *MangaHandler) HandleMangaPage(w http.ResponseWriter, r *http.Request) {
	pageStr := r.PathValue("page")
	mangaName := r.PathValue("name")
	fullPath := filepath.Join(h.BaseDir, mangaName+".cbz")

	pageIdx, err := strconv.Atoi(pageStr)
	if err != nil || pageIdx < 0 {
		http.Error(w, "Invalid page number", http.StatusBadRequest)
		return
	}

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		http.Error(w, "Manga not found", http.StatusNotFound)
		return
	}

	manga, err := cbzReader.Open(fullPath)
	if err != nil {
		fmt.Printf("Error opening cbzReader: %v\n", err)
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
	w.Header().Set("Cache-Control", "public, max-age=3600")

	if _, err := io.Copy(w, rc); err != nil {
		fmt.Printf("Stream error: %v\n", err)
	}
}

func (h *MangaHandler) HandleMangaInfo(w http.ResponseWriter, r *http.Request) {
	mangaName := r.PathValue("name")
	fullPath := filepath.Join(h.BaseDir, mangaName+".cbz")

	manga, err := cbzReader.Open(fullPath)
	if err != nil {
		fmt.Printf("Error opening cbzReader: %v\n", err)
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

func (h *MangaHandler) HandleMangaReader(w http.ResponseWriter, r *http.Request) {
	mangaName := r.PathValue("name")
	pageStr := r.PathValue("page")

	if pageStr == "" {
		pageStr = "0"
	}

	redirectURL := fmt.Sprintf("/?manga=%s&page=%s", mangaName, pageStr)
	http.Redirect(w, r, redirectURL, http.StatusMovedPermanently)
}

func (h *MangaHandler) HandleMangaSnippet(w http.ResponseWriter, r *http.Request) {
	mangaName := r.PathValue("name")
	pageStr := r.PathValue("page")

	pageIdx, err := strconv.Atoi(pageStr)
	if err != nil || pageIdx < 0 {
		http.Error(w, "Invalid page number", http.StatusBadRequest)
		return
	}

	fullPath := filepath.Join(h.BaseDir, mangaName+".cbz")
	manga, err := cbzReader.Open(fullPath)
	if err != nil {
		http.Error(w, "Manga not found", http.StatusNotFound)
		return
	}
	defer manga.Close()

	if pageIdx >= manga.PageCount {
		pageIdx = manga.PageCount - 1
	}

	snippet := fmt.Sprintf(`<div id="image-container" class="image-wrapper">
    <img id="manga-page" src="/manga/%s/page/%d" alt="Manga Page">
</div>`, mangaName, pageIdx)

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(snippet))
}
