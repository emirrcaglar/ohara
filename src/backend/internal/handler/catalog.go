package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"ohara/src/internal/db"
	"ohara/src/internal/logger"
)

type CatalogHandler struct {
	DB  *db.DB
	Log *logger.Logger
}

type CatalogResponse struct {
	ID          int64  `json:"id"`
	ParentID    *int64 `json:"parentId"`
	Name        string `json:"name"`
	ObjectCount int    `json:"objectCount"`
}

type CatalogListResponse struct {
	Items []CatalogResponse `json:"items"`
	Path  []CatalogResponse `json:"path"`
}

type CatalogCreateRequest struct {
	ParentID *int64 `json:"parentId"`
	Name     string `json:"name"`
}

type CatalogUpdateRequest struct {
	ParentID *int64 `json:"parentId"`
	Name     string `json:"name"`
}

func (h *CatalogHandler) HandleCatalogList(w http.ResponseWriter, r *http.Request) {
	parentID, ok := parseOptionalInt64(r.URL.Query().Get("parentId"))
	if !ok {
		http.Error(w, "Invalid parentId", http.StatusBadRequest)
		return
	}

	items, err := h.DB.ListCatalogChildren(parentID)
	if err != nil {
		if h.Log != nil {
			h.Log.Error("[catalog] list failed parent=%v err=%v", parentID, err)
		}
		http.Error(w, "Failed to load catalog", http.StatusInternalServerError)
		return
	}

	path := []db.CatalogRow{}
	if parentID != nil {
		path, err = h.DB.GetCatalogPath(*parentID)
		if err != nil {
			if h.Log != nil {
				h.Log.Error("[catalog] path failed parent=%d err=%v", *parentID, err)
			}
			http.Error(w, "Failed to load catalog path", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CatalogListResponse{
		Items: catalogRowsToResponse(items),
		Path:  catalogRowsToResponse(path),
	})
}

func (h *CatalogHandler) HandleCatalogAll(w http.ResponseWriter, r *http.Request) {
	items, err := h.DB.ListCatalogAll()
	if err != nil {
		if h.Log != nil {
			h.Log.Error("[catalog] list all failed err=%v", err)
		}
		http.Error(w, "Failed to load catalogs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CatalogListResponse{
		Items: catalogRowsToResponse(items),
		Path:  []CatalogResponse{},
	})
}

func (h *CatalogHandler) HandleCatalogCreate(w http.ResponseWriter, r *http.Request) {
	var req CatalogCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		http.Error(w, "Folder name is required", http.StatusBadRequest)
		return
	}
	if !h.validCatalogParent(w, req.ParentID) {
		return
	}

	folder, err := h.DB.InsertCatalog(req.ParentID, name)
	if err != nil {
		if h.Log != nil {
			h.Log.Error("[catalog] create failed parent=%v name=%s err=%v", req.ParentID, name, err)
		}
		if isConstraintError(err) {
			http.Error(w, "Folder already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to create folder", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(catalogRowToResponse(*folder))
}

func (h *CatalogHandler) HandleCatalogGet(w http.ResponseWriter, r *http.Request) {
	id, ok := parseCatalogID(w, r)
	if !ok {
		return
	}

	folder, err := h.DB.GetCatalogByID(id)
	if err != nil {
		if h.Log != nil {
			h.Log.Error("[catalog] get failed id=%d err=%v", id, err)
		}
		http.Error(w, "Failed to load folder", http.StatusInternalServerError)
		return
	}
	if folder == nil {
		http.Error(w, "Folder not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(catalogRowToResponse(*folder))
}

func (h *CatalogHandler) HandleCatalogUpdate(w http.ResponseWriter, r *http.Request) {
	id, ok := parseCatalogID(w, r)
	if !ok {
		return
	}

	var req CatalogUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		http.Error(w, "Folder name is required", http.StatusBadRequest)
		return
	}
	if !h.validCatalogParent(w, req.ParentID) {
		return
	}

	folder, err := h.DB.UpdateCatalog(id, req.ParentID, name)
	if err != nil {
		if h.Log != nil {
			h.Log.Error("[catalog] update failed id=%d parent=%v name=%s err=%v", id, req.ParentID, name, err)
		}
		if errors.Is(err, db.ErrCatalogCycle) {
			http.Error(w, "Folder cannot be moved into itself", http.StatusBadRequest)
			return
		}
		if isConstraintError(err) {
			http.Error(w, "Folder already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to update folder", http.StatusInternalServerError)
		return
	}
	if folder == nil {
		http.Error(w, "Folder not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(catalogRowToResponse(*folder))
}

func (h *CatalogHandler) HandleCatalogDelete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseCatalogID(w, r)
	if !ok {
		return
	}

	deleted, err := h.DB.DeleteCatalog(id)
	if err != nil {
		if h.Log != nil {
			h.Log.Error("[catalog] delete failed id=%d err=%v", id, err)
		}
		http.Error(w, "Failed to delete folder", http.StatusInternalServerError)
		return
	}
	if !deleted {
		http.Error(w, "Folder not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseCatalogID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid folder id", http.StatusBadRequest)
		return 0, false
	}
	return id, true
}

func (h *CatalogHandler) validCatalogParent(w http.ResponseWriter, parentID *int64) bool {
	if parentID == nil {
		return true
	}
	parent, err := h.DB.GetCatalogByID(*parentID)
	if err != nil {
		if h.Log != nil {
			h.Log.Error("[catalog] parent lookup failed id=%d err=%v", *parentID, err)
		}
		http.Error(w, "Failed to validate parent folder", http.StatusInternalServerError)
		return false
	}
	if parent == nil {
		http.Error(w, "Parent folder not found", http.StatusBadRequest)
		return false
	}
	return true
}

func validMediaCatalog(w http.ResponseWriter, database *db.DB, catalogID *int64) bool {
	if catalogID == nil {
		return true
	}
	folder, err := database.GetCatalogByID(*catalogID)
	if err != nil {
		http.Error(w, "Failed to validate destination folder", http.StatusInternalServerError)
		return false
	}
	if folder == nil {
		http.Error(w, "Destination folder not found", http.StatusBadRequest)
		return false
	}
	return true
}

func isConstraintError(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "constraint")
}

func parseOptionalInt64(value string) (*int64, bool) {
	if value == "" {
		return nil, true
	}
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return nil, false
	}
	return &id, true
}

func catalogRowsToResponse(rows []db.CatalogRow) []CatalogResponse {
	items := make([]CatalogResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, catalogRowToResponse(row))
	}
	return items
}

func catalogRowToResponse(row db.CatalogRow) CatalogResponse {
	return CatalogResponse{
		ID:          row.ID,
		ParentID:    row.ParentID,
		Name:        row.Name,
		ObjectCount: row.ObjectCount,
	}
}
