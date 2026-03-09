package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"ohara/src/internal/db"
	"ohara/src/internal/utils/imgutil"

	cbzReader "ohara/src/internal/media/cbz"
)

type MangaHandler struct {
	DB       *db.DB
	Cache    *PageCache
	Inflight *Inflight
}

func (h *MangaHandler) mangaByID(idStr string) (*db.MangaRow, int, error) {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("invalid id")
	}
	m, err := h.DB.GetMangaByID(id)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	if m == nil {
		return nil, http.StatusNotFound, fmt.Errorf("manga not found")
	}
	return m, 0, nil
}

func (h *MangaHandler) HandleMangaList(w http.ResponseWriter, r *http.Request) {
	mangas, err := h.DB.ListManga(1)
	if err != nil {
		http.Error(w, "Failed to load library", http.StatusInternalServerError)
		return
	}

	var cards strings.Builder
	for _, m := range mangas {
		cards.WriteString(fmt.Sprintf(`
		<a class="manga-card" href="/manga/%d/resume">
			<div class="cover-wrap">
				<img src="/manga/%d/page/0" alt="%s" loading="lazy">
				<span class="progress-badge">%d / %d</span>
			</div>
			<span class="title">%s</span>
		</a>`, m.ID, m.ID, m.Title, m.Progress, m.PageCount, m.Title))
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
					.cover-wrap { position: relative; width: 100%%; }
					.cover-wrap img { width: 100%%; aspect-ratio: 2/3; object-fit: cover; background: #333; display: block; }
					.progress-badge { position: absolute; bottom: 0; left: 0; right: 0; background: rgba(0,0,0,0.65); color: #ccc; font-size: 0.75rem; text-align: center; padding: 3px 0; }
					.title { padding: 8px; font-size: 0.85rem; text-align: center; word-break: break-word; }
				</style>
			</head>
			<body>
				<div class="grid">%s</div>
			</body>
		</html>`, cards.String())

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func (h *MangaHandler) compressPage(m *db.MangaRow, pageIdx int) ([]byte, bool, error) {
	if data, ok := h.Cache.Get(m.ID, pageIdx); ok {
		return data, true, nil
	}

	data, err := h.Inflight.Do(m.ID, pageIdx, func() ([]byte, error) {
		if data, ok := h.Cache.Get(m.ID, pageIdx); ok {
			return data, nil
		}

		t := time.Now()
		manga, err := cbzReader.Open(m.Path)
		if err != nil {
			return nil, err
		}
		defer manga.Close()
		openDur := time.Since(t)

		t = time.Now()
		rc, err := manga.GetPageReader(pageIdx)
		if err != nil {
			return nil, err
		}
		defer rc.Close()

		var buf bytes.Buffer
		if err := imgutil.Compress(rc, &buf, 1200, 70); err != nil {
			return nil, err
		}
		compressDur := time.Since(t)

		data := buf.Bytes()
		h.Cache.Set(m.ID, pageIdx, data)

		fmt.Printf("[compress] manga=%d page=%d size=%dKB open=%v compress=%v\n",
			m.ID, pageIdx, len(data)/1024, openDur, compressDur)

		return data, nil
	})

	return data, false, err
}

func (h *MangaHandler) prefetchAhead(m *db.MangaRow, fromPage, count int) {
	go func() {
		for i := 1; i <= count; i++ {
			p := fromPage + i
			if p >= m.PageCount {
				break
			}
			if _, ok := h.Cache.Get(m.ID, p); ok {
				continue
			}
			fmt.Printf("[prefetch] manga=%d page=%d compressing...\n", m.ID, p)
			if _, _, err := h.compressPage(m, p); err != nil {
				fmt.Printf("[prefetch] manga=%d page=%d error: %v\n", m.ID, p, err)
			}
		}
	}()
}

func (h *MangaHandler) HandleMangaPage(w http.ResponseWriter, r *http.Request) {
	t0 := time.Now()

	m, status, err := h.mangaByID(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}

	pageIdx, err := strconv.Atoi(r.PathValue("page"))
	if err != nil || pageIdx < 0 {
		http.Error(w, "Invalid page number", http.StatusBadRequest)
		return
	}

	data, cached, err := h.compressPage(m, pageIdx)
	if err != nil {
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}

	source := "compressed"
	if cached {
		source = "cache"
	}

	go h.prefetchAhead(m, pageIdx, 15)

	fmt.Printf("[page] manga=%d page=%d size=%dKB source=%s total=%v\n",
		m.ID, pageIdx, len(data)/1024, source, time.Since(t0))

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.Write(data)
}

func (h *MangaHandler) HandleMangaProgress(w http.ResponseWriter, r *http.Request) {
	m, status, err := h.mangaByID(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}

	pageIdx, err := strconv.Atoi(r.PathValue("page"))
	if err != nil || pageIdx < 0 {
		http.Error(w, "Invalid page number", http.StatusBadRequest)
		return
	}

	if err := h.DB.UpsertProgress(1, m.ID, pageIdx); err != nil {
		fmt.Printf("[progress] save error manga=%d page=%d: %v\n", m.ID, pageIdx, err)
		http.Error(w, "Failed to save progress", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *MangaHandler) HandleMangaInfo(w http.ResponseWriter, r *http.Request) {
	m, status, err := h.mangaByID(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}

	manga, err := cbzReader.Open(m.Path)
	if err != nil {
		http.Error(w, "Could not open manga file", http.StatusInternalServerError)
		return
	}
	defer manga.Close()

	data, err := json.MarshalIndent(manga, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (h *MangaHandler) HandleMangaResume(w http.ResponseWriter, r *http.Request) {
	m, status, err := h.mangaByID(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}

	page, err := h.DB.GetProgress(1, m.ID)
	if err != nil {
		http.Error(w, "Failed to fetch progress", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/reader?manga=%d&page=%d&total=%d", m.ID, page, m.PageCount), http.StatusFound)
}
