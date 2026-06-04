package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"

	"ohara/src/internal/db"
	"ohara/src/internal/logger"
	"ohara/src/internal/media/cbz"
	"ohara/src/internal/utils/imgutil"
)

type MangaHandler struct {
	DB         *db.DB
	Cache      *PageCache
	Inflight   *Inflight
	CBZService cbz.ICBZService
	Log        *logger.Logger
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
	user := GetUser(r.Context())
	mangas, err := h.DB.ListManga(user.ID)
	if err != nil {
		http.Error(w, "Failed to load library", http.StatusInternalServerError)
		return
	}

	items := make([]MangaResponse, 0, len(mangas))
	for _, m := range mangas {
		items = append(items, MangaResponse{
			ID:            m.ID,
			CatalogID:     m.CatalogID,
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
	CatalogID     *int64 `json:"catalogId"`
	Title         string `json:"title"`
	PageCount     int    `json:"pageCount"`
	CurrentPage   int    `json:"currentPage"`
	FileExtension string `json:"fileExtension"`
}

type MangaLibraryResponse struct {
	Items []MangaResponse `json:"items"`
	Total int             `json:"total"`
}

type MediaCatalogUpdateRequest struct {
	CatalogID *int64 `json:"catalogId"`
}

func (h *MangaHandler) compressPage(m *db.MangaRow, pageIdx int) ([]byte, bool, error) {
	if data, ok := h.Cache.Get(m.ID, pageIdx); ok {
		return data, true, nil
	}

	data, err := h.Inflight.Do(m.ID, pageIdx, func() ([]byte, error) {
		if data, ok := h.Cache.Get(m.ID, pageIdx); ok {
			return data, nil
		}

		manga, err := h.CBZService.Open(m.Path)
		if err != nil {
			return nil, err
		}
		defer manga.Close()

		rc, err := manga.GetPageReader(pageIdx)
		if err != nil {
			return nil, err
		}
		defer rc.Close()

		var buf bytes.Buffer
		if err := imgutil.Compress(rc, &buf, 1200, 70); err != nil {
			return nil, err
		}

		data := buf.Bytes()
		h.Cache.Set(m.ID, pageIdx, data)

		return data, nil
	})

	return data, false, err
}

func (h *MangaHandler) HandleMangaPage(w http.ResponseWriter, r *http.Request) {

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

	data, _, err := h.compressPage(m, pageIdx)
	if err != nil {
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.Write(data)
}

func (h *MangaHandler) HandleMangaDelete(w http.ResponseWriter, r *http.Request) {
	m, status, err := h.mangaByID(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}

	if err := h.DB.DeleteManga(m.ID); err != nil {
		if h.Log != nil {
			h.Log.Error("[manga] delete failed manga=%d err=%v", m.ID, err)
		}
		http.Error(w, "Failed to delete manga", http.StatusInternalServerError)
		return
	}

	if err := h.Cache.DeleteManga(m.ID); err != nil {
		if h.Log != nil {
			h.Log.Error("[manga] cache delete failed manga=%d err=%v", m.ID, err)
		}
		http.Error(w, "Failed to delete manga cache", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *MangaHandler) HandleMangaCatalogUpdate(w http.ResponseWriter, r *http.Request) {
	m, status, err := h.mangaByID(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), status)
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

	if err := h.DB.UpdateMangaCatalog(m.ID, req.CatalogID); err != nil {
		if h.Log != nil {
			h.Log.Error("[manga] catalog update failed manga=%d catalog=%v err=%v", m.ID, req.CatalogID, err)
		}
		http.Error(w, "Failed to move manga", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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

	user := GetUser(r.Context())
	if err := h.DB.UpsertProgress(user.ID, m.ID, pageIdx); err != nil {
		if h.Log != nil {
			h.Log.Error("[manga] progress save failed manga=%d page=%d err=%v", m.ID, pageIdx, err)
		}
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
		CatalogID:   m.CatalogID,
		Title:       m.Title,
		Path:        m.Path,
		PageCount:   m.PageCount,
		CurrentPage: m.Progress,
	})
}

type MangaDetailResponse struct {
	ID          int64  `json:"id"`
	CatalogID   *int64 `json:"catalogId"`
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

	user := GetUser(r.Context())
	page, err := h.DB.GetProgress(user.ID, m.ID)
	if err != nil {
		http.Error(w, "Failed to fetch progress", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/reader?manga=%d&page=%d&total=%d", m.ID, page, m.PageCount), http.StatusFound)
}
