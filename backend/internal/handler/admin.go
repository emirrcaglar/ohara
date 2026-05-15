package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"ohara/src/internal/db"
)

type AdminHandler struct {
	DB *db.DB
}

func (h *AdminHandler) HandleListPendingUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(`
		SELECT id, username, role, is_approved, created_at
		FROM user WHERE is_approved = 0
	`)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []map[string]interface{}
	for rows.Next() {
		var id int64
		var username, role, createdAt string
		var isApproved bool
		rows.Scan(&id, &username, &role, &isApproved, &createdAt)
		users = append(users, map[string]interface{}{
			"id":         id,
			"username":   username,
			"role":       role,
			"isApproved": isApproved,
			"createdAt":  createdAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (h *AdminHandler) HandleApproveUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	_, err = h.DB.Exec("UPDATE user SET is_approved = 1 WHERE id = ?", id)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
