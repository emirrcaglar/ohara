package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"ohara/src/internal/db"
	"ohara/src/internal/media/cbz"
	"ohara/src/internal/utils/imgutil"
)

type MangaHandler struct {
	DB         *db.DB
	Cache      *PageCache
	Inflight   *Inflight
	CBZService cbz.ICBZService
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

	items := make([]MangaResponse, 0, len(mangas))
	for _, m := range mangas {
		items = append(items, MangaResponse{
			ID:            m.ID,
			Title:         m.Title,
			PageCount:     m.PageCount,
			CurrentPage:   m.Progress,
			FileExtension: filepath.Ext(m.Path),
		})
	}

	response := MangaLibraryResponse{
		Items: items,
		Total: len(items),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type MangaResponse struct {
	ID            int64  `json:"id"`
	Title         string `json:"title"`
	PageCount     int    `json:"pageCount"`
	CurrentPage   int    `json:"currentPage"`
	FileExtension string `json:"fileExtension"`
}

type MangaLibraryResponse struct {
	Items []MangaResponse `json:"items"`
	Total int             `json:"total"`
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
		manga, err := h.CBZService.Open(m.Path)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(MangaDetailResponse{
		ID:          m.ID,
		Title:       m.Title,
		Path:        m.Path,
		PageCount:   m.PageCount,
		CurrentPage: m.Progress,
	})
}

type MangaDetailResponse struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Path        string `json:"path"`
	PageCount   int    `json:"pageCount"`
	CurrentPage int    `json:"currentPage"`
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
