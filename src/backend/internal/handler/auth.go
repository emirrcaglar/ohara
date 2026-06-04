package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"ohara/src/internal/db"
	"ohara/src/internal/logger"

	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	DB  *db.DB
	Log *logger.Logger
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UpdatePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if h.Log != nil {
			h.Log.Warn("[auth] login request decode failed err=%v", err)
		}
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	user, err := h.DB.GetUserByUsername(req.Username)
	if err != nil {
		if h.Log != nil {
			h.Log.Warn("[auth] login failed username=%s err=%v", req.Username, err)
		}
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if !user.IsApproved {
		if h.Log != nil {
			h.Log.Warn("[auth] login blocked pending approval username=%s", req.Username)
		}
		http.Error(w, "Account pending approval", http.StatusForbidden)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		if h.Log != nil {
			h.Log.Warn("[auth] login failed invalid password username=%s", req.Username)
		}
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if h.Log != nil {
		h.Log.Info("[auth] login success username=%s role=%s", user.Username, user.Role)
	}

	// TODO: JWT will be used
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    user.Username,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(24 * time.Hour),
	})

	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if h.Log != nil {
			h.Log.Warn("[auth] register request decode failed err=%v", err)
		}
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		if h.Log != nil {
			h.Log.Warn("[auth] register rejected missing fields username=%q", req.Username)
		}
		http.Error(w, "Username and password required", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		if h.Log != nil {
			h.Log.Error("[auth] register password hash failed username=%s err=%v", req.Username, err)
		}
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	_, err = h.DB.Exec(`
		INSERT INTO user (username, password_hash, role, is_approved)
		VALUES (?, ?, 'user', 0)
	`, req.Username, string(hash))

	if err != nil {
		if h.Log != nil {
			h.Log.Warn("[auth] register conflict username=%s err=%v", req.Username, err)
		}
		http.Error(w, "Username already exists", http.StatusConflict)
		return
	}

	if h.Log != nil {
		h.Log.Info("[auth] register success username=%s", req.Username)
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	if h.Log != nil {
		h.Log.Info("[auth] logout")
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) HandleUpdatePassword(w http.ResponseWriter, r *http.Request) {
	var req UpdatePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if h.Log != nil {
			h.Log.Warn("[auth] password update request decode failed err=%v", err)
		}
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.CurrentPassword == "" || req.NewPassword == "" {
		if h.Log != nil {
			h.Log.Warn("[auth] password update rejected missing fields")
		}
		http.Error(w, "Current password and new password required", http.StatusBadRequest)
		return
	}

	user := GetUser(r.Context())
	if user == nil {
		if h.Log != nil {
			h.Log.Warn("[auth] password update missing user context")
		}
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		if h.Log != nil {
			h.Log.Warn("[auth] password update failed invalid current password username=%s", user.Username)
		}
		http.Error(w, "Invalid current password", http.StatusUnauthorized)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		if h.Log != nil {
			h.Log.Error("[auth] password update hash failed username=%s err=%v", user.Username, err)
		}
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	if err := h.DB.UpdateUserPassword(user.ID, string(hash)); err != nil {
		if h.Log != nil {
			h.Log.Error("[auth] password update failed username=%s err=%v", user.Username, err)
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if h.Log != nil {
		h.Log.Info("[auth] password update success username=%s", user.Username)
	}
	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) HandleMe(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil {
		if h.Log != nil {
			h.Log.Warn("[auth] me missing session cookie")
		}
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	user, err := h.DB.GetUserByUsername(cookie.Value)
	if err != nil {
		if h.Log != nil {
			h.Log.Warn("[auth] me lookup failed username=%s err=%v", cookie.Value, err)
		}
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         user.ID,
		"username":   user.Username,
		"role":       user.Role,
		"isApproved": user.IsApproved,
	})
}
