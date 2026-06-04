package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"ohara/src/internal/db"
	"ohara/src/internal/logger"
)

type AdminHandler struct {
	DB  *db.DB
	Log *logger.Logger
}

func (h *AdminHandler) HandleListPendingUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(`
		SELECT id, username, role, is_approved, pfp, created_at
		FROM user WHERE is_approved = 0
	`)
	if err != nil {
		if h.Log != nil {
			h.Log.Error("[admin] list pending users query failed err=%v", err)
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	users := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id int64
		var username, role, createdAt string
		var isApproved bool
		var pfp int
		rows.Scan(&id, &username, &role, &isApproved, &pfp, &createdAt)
		users = append(users, map[string]interface{}{
			"id":         id,
			"username":   username,
			"role":       role,
			"isApproved": isApproved,
			"pfp":        pfp,
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
		if h.Log != nil {
			h.Log.Warn("[admin] approve invalid id=%q", idStr)
		}
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	_, err = h.DB.Exec("UPDATE user SET is_approved = 1 WHERE id = ?", id)
	if err != nil {
		if h.Log != nil {
			h.Log.Error("[admin] approve failed id=%d err=%v", id, err)
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if h.Log != nil {
		h.Log.Info("[admin] user approved id=%d", id)
	}
	w.WriteHeader(http.StatusOK)
}

func (h *AdminHandler) HandleRejectUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		if h.Log != nil {
			h.Log.Warn("[admin] reject invalid id=%q", idStr)
		}
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	result, err := h.DB.Exec("DELETE FROM user WHERE id = ? AND is_approved = 0", id)
	if err != nil {
		if h.Log != nil {
			h.Log.Error("[admin] reject failed id=%d err=%v", id, err)
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		if h.Log != nil {
			h.Log.Error("[admin] reject rows affected failed id=%d err=%v", id, err)
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		if h.Log != nil {
			h.Log.Warn("[admin] reject no pending user id=%d", id)
		}
		http.Error(w, "Pending user not found", http.StatusNotFound)
		return
	}

	if h.Log != nil {
		h.Log.Info("[admin] user rejected id=%d", id)
	}
	w.WriteHeader(http.StatusOK)
}
