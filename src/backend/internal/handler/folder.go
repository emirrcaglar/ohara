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

type FolderHandler struct {
	DB  *db.DB
	Log *logger.Logger
}

type FolderResponse struct {
	ID          int64  `json:"id"`
	ParentID    *int64 `json:"parentId"`
	Name        string `json:"name"`
	ObjectCount int    `json:"objectCount"`
}

type FolderListResponse struct {
	Items []FolderResponse `json:"items"`
	Path  []FolderResponse `json:"path"`
}

type FolderCreateRequest struct {
	ParentID *int64 `json:"parentId"`
	Name     string `json:"name"`
}

type FolderUpdateRequest struct {
	ParentID *int64 `json:"parentId"`
	Name     string `json:"name"`
}

func (h *FolderHandler) HandleFolderList(w http.ResponseWriter, r *http.Request) {
	parentID, ok := parseOptionalInt64(r.URL.Query().Get("parentId"))
	if !ok {
		http.Error(w, "Invalid parentId", http.StatusBadRequest)
		return
	}

	items, err := h.DB.ListFolderChildren(parentID)
	if err != nil {
		if h.Log != nil {
			h.Log.Error("[folder] list failed parent=%v err=%v", parentID, err)
		}
		http.Error(w, "Failed to load folder", http.StatusInternalServerError)
		return
	}

	path := []db.FolderRow{}
	if parentID != nil {
		path, err = h.DB.GetFolderPath(*parentID)
		if err != nil {
			if h.Log != nil {
				h.Log.Error("[folder] path failed parent=%d err=%v", *parentID, err)
			}
			http.Error(w, "Failed to load folder path", http.StatusInternalServerError)
			return
		}
		if len(path) == 0 {
			http.Error(w, "Folder not found", http.StatusNotFound)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(FolderListResponse{
		Items: folderRowsToResponse(items),
		Path:  folderRowsToResponse(path),
	})
}

func (h *FolderHandler) HandleFolderAll(w http.ResponseWriter, r *http.Request) {
	items, err := h.DB.ListFolderAll()
	if err != nil {
		if h.Log != nil {
			h.Log.Error("[folder] list all failed err=%v", err)
		}
		http.Error(w, "Failed to load folders", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(FolderListResponse{
		Items: folderRowsToResponse(items),
		Path:  []FolderResponse{},
	})
}

func (h *FolderHandler) HandleFolderCreate(w http.ResponseWriter, r *http.Request) {
	var req FolderCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		http.Error(w, "Folder name is required", http.StatusBadRequest)
		return
	}
	if !h.validFolderParent(w, req.ParentID) {
		return
	}

	folder, err := h.DB.InsertFolder(req.ParentID, name)
	if err != nil {
		if h.Log != nil {
			h.Log.Error("[folder] create failed parent=%v name=%s err=%v", req.ParentID, name, err)
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
	json.NewEncoder(w).Encode(folderRowToResponse(*folder))
}

func (h *FolderHandler) HandleFolderGet(w http.ResponseWriter, r *http.Request) {
	id, ok := parseFolderID(w, r)
	if !ok {
		return
	}

	folder, err := h.DB.GetFolderByID(id)
	if err != nil {
		if h.Log != nil {
			h.Log.Error("[folder] get failed id=%d err=%v", id, err)
		}
		http.Error(w, "Failed to load folder", http.StatusInternalServerError)
		return
	}
	if folder == nil {
		http.Error(w, "Folder not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(folderRowToResponse(*folder))
}

func (h *FolderHandler) HandleFolderUpdate(w http.ResponseWriter, r *http.Request) {
	id, ok := parseFolderID(w, r)
	if !ok {
		return
	}

	var req FolderUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		http.Error(w, "Folder name is required", http.StatusBadRequest)
		return
	}
	if !h.validFolderParent(w, req.ParentID) {
		return
	}

	folder, err := h.DB.UpdateFolder(id, req.ParentID, name)
	if err != nil {
		if h.Log != nil {
			h.Log.Error("[folder] update failed id=%d parent=%v name=%s err=%v", id, req.ParentID, name, err)
		}
		if errors.Is(err, db.ErrFolderCycle) {
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
	json.NewEncoder(w).Encode(folderRowToResponse(*folder))
}

func (h *FolderHandler) HandleFolderDelete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseFolderID(w, r)
	if !ok {
		return
	}

	deleted, err := h.DB.DeleteFolder(id)
	if err != nil {
		if h.Log != nil {
			h.Log.Error("[folder] delete failed id=%d err=%v", id, err)
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

func parseFolderID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid folder id", http.StatusBadRequest)
		return 0, false
	}
	return id, true
}

func (h *FolderHandler) validFolderParent(w http.ResponseWriter, parentID *int64) bool {
	if parentID == nil {
		return true
	}
	parent, err := h.DB.GetFolderByID(*parentID)
	if err != nil {
		if h.Log != nil {
			h.Log.Error("[folder] parent lookup failed id=%d err=%v", *parentID, err)
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

func validMediaFolder(w http.ResponseWriter, database *db.DB, folderID *int64) bool {
	if folderID == nil {
		return true
	}
	folder, err := database.GetFolderByID(*folderID)
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

func folderRowsToResponse(rows []db.FolderRow) []FolderResponse {
	items := make([]FolderResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, folderRowToResponse(row))
	}
	return items
}

func folderRowToResponse(row db.FolderRow) FolderResponse {
	return FolderResponse{
		ID:          row.ID,
		ParentID:    row.ParentID,
		Name:        row.Name,
		ObjectCount: row.ObjectCount,
	}
}
