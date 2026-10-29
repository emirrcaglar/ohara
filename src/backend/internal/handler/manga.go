package handler

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"ohara/src/internal/db"
	"ohara/src/internal/logger"
	"ohara/src/internal/media/manga"
)

type MangaHandler struct {
	DB  *db.DB
	Log *logger.Logger
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
	records, err := h.DB.ListManga(user.ID)
	if err != nil {
		http.Error(w, "Failed to load library", http.StatusInternalServerError)
		return
	}

	items := make([]MangaResponse, 0, len(records))
	for _, m := range records {
		items = append(items, MangaResponse{
			ID:            m.ID,
			CatalogID:     m.CatalogID,
			Title:         m.Title,
			SeriesTitle:   m.SeriesTitle,
			FileName:      filepath.Base(m.Path),
			Metadata:      m.Metadata,
			ParserVersion: m.ParserVersion,
			PageCount:     m.PageCount,
			CurrentPage:   m.Progress,
			UpdatedAt:     m.UpdatedAt,
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
	SeriesTitle   string          `json:"seriesTitle"`
	FileName      string          `json:"fileName"`
	Metadata      *manga.Metadata `json:"metadata"`
	ParserVersion string          `json:"parserVersion"`
	ID            int64           `json:"id"`
	CatalogID     *int64          `json:"catalogId"`
	Title         string          `json:"title"`
	PageCount     int             `json:"pageCount"`
	CurrentPage   int             `json:"currentPage"`
	FileExtension string          `json:"fileExtension"`
	UpdatedAt     *time.Time      `json:"updatedAt"`
}

type MangaLibraryResponse struct {
	Items []MangaResponse `json:"items"`
	Total int             `json:"total"`
}

type MediaCatalogUpdateRequest struct {
	CatalogID *int64 `json:"catalogId"`
}

func (h *MangaHandler) HandleMangaPage(w http.ResponseWriter, r *http.Request) {

	m, status, err := h.mangaByID(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}

	pageIndex, err := strconv.Atoi(r.PathValue("page"))
	if err != nil || pageIndex < 0 {
		http.Error(w, "Invalid page number", http.StatusBadRequest)
		return
	}

	archive, err := manga.Open(m.Path)
	if err != nil {
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}
	defer archive.Close()
	if pageIndex >= len(archive.Pages) {
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}
	rc, err := archive.OpenPage(pageIndex)
	if err != nil {
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}
	defer rc.Close()

	// Sniff a bounded prefix without decoding image pixels or buffering the
	// whole entry. Peek leaves these bytes available for the subsequent copy.
	reader := bufio.NewReader(rc)
	prefix, err := reader.Peek(512)
	if err != nil && err != io.EOF {
		http.Error(w, "Failed to read page", http.StatusInternalServerError)
		return
	}
	contentType := http.DetectContentType(prefix)
	if contentType != "image/png" && contentType != "image/jpeg" && contentType != "image/webp" {
		http.Error(w, "Unsupported page image", http.StatusUnsupportedMediaType)
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "private, no-store, no-transform")
	w.Header().Set("X-Ohara-Representation", "original")
	w.Header().Set("Content-Length", strconv.FormatUint(archive.Pages[pageIndex].Size, 10))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if r.Method == http.MethodHead || r.Context().Err() != nil {
		return
	}
	if _, err := io.Copy(w, reader); err != nil && r.Context().Err() == nil && h.Log != nil {
		h.Log.Error("[manga] original page stream failed manga=%d page=%d err=%v", m.ID, pageIndex, err)
	}
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

	pageIndex, err := strconv.Atoi(r.PathValue("page"))
	if err != nil || pageIndex < 0 {
		http.Error(w, "Invalid page number", http.StatusBadRequest)
		return
	}

	user := GetUser(r.Context())
	if err := h.DB.UpsertMangaProgress(user.ID, m.ID, pageIndex); err != nil {
		if h.Log != nil {
			h.Log.Error("[manga] progress save failed manga=%d page=%d err=%v", m.ID, pageIndex, err)
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
		ID:            m.ID,
		CatalogID:     m.CatalogID,
		Title:         m.Title,
		SeriesTitle:   m.SeriesTitle,
		FileName:      filepath.Base(m.Path),
		Metadata:      m.Metadata,
		ParserVersion: m.ParserVersion,
		Path:          m.Path,
		PageCount:     m.PageCount,
		CurrentPage:   m.Progress,
	})
}

type MangaDetailResponse struct {
	SeriesTitle   string          `json:"seriesTitle"`
	FileName      string          `json:"fileName"`
	Metadata      *manga.Metadata `json:"metadata"`
	ParserVersion string          `json:"parserVersion"`
	ID            int64           `json:"id"`
	CatalogID     *int64          `json:"catalogId"`
	Title         string          `json:"title"`
	Path          string          `json:"path"`
	PageCount     int             `json:"pageCount"`
	CurrentPage   int             `json:"currentPage"`
}

func (h *MangaHandler) HandleMangaResume(w http.ResponseWriter, r *http.Request) {
	m, status, err := h.mangaByID(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}

	user := GetUser(r.Context())
	page, err := h.DB.GetMangaProgress(user.ID, m.ID)
	if err != nil {
		http.Error(w, "Failed to fetch progress", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/reader?manga=%d&page=%d&total=%d", m.ID, page, m.PageCount), http.StatusFound)
}
