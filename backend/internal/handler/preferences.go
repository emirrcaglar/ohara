package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"ohara/src/internal/db"
	"ohara/src/internal/logger"
)

type PreferencesHandler struct {
	DB  *db.DB
	Log *logger.Logger
}

type preferenceRequest struct {
	Value string `json:"value"`
}

type preferenceResponse struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type preferencesResponse struct {
	Preferences map[string]string `json:"preferences"`
}

func (h *PreferencesHandler) HandlePreferencesList(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r.Context())
	preferences, err := h.DB.ListPreferences(user.ID)
	if err != nil {
		if h.Log != nil {
			h.Log.Error("[preferences] list failed user=%d err=%v", user.ID, err)
		}
		http.Error(w, "Failed to load preferences", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(preferencesResponse{Preferences: preferences})
}

func (h *PreferencesHandler) HandlePreferenceGet(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimSpace(r.PathValue("key"))
	if key == "" {
		http.Error(w, "Preference key is required", http.StatusBadRequest)
		return
	}

	user := GetUser(r.Context())
	value, ok, err := h.DB.GetPreference(user.ID, key)
	if err != nil {
		if h.Log != nil {
			h.Log.Error("[preferences] get failed user=%d key=%s err=%v", user.ID, key, err)
		}
		http.Error(w, "Failed to load preference", http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Error(w, "Preference not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(preferenceResponse{Key: key, Value: value})
}

func (h *PreferencesHandler) HandlePreferenceUpsert(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimSpace(r.PathValue("key"))
	if key == "" {
		http.Error(w, "Preference key is required", http.StatusBadRequest)
		return
	}

	var req preferenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	user := GetUser(r.Context())
	if err := h.DB.UpsertPreference(user.ID, key, req.Value); err != nil {
		if h.Log != nil {
			h.Log.Error("[preferences] save failed user=%d key=%s err=%v", user.ID, key, err)
		}
		http.Error(w, "Failed to save preference", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(preferenceResponse{Key: key, Value: req.Value})
}

func (h *PreferencesHandler) HandlePreferenceDelete(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimSpace(r.PathValue("key"))
	if key == "" {
		http.Error(w, "Preference key is required", http.StatusBadRequest)
		return
	}

	user := GetUser(r.Context())
	if err := h.DB.DeletePreference(user.ID, key); err != nil {
		if h.Log != nil {
			h.Log.Error("[preferences] delete failed user=%d key=%s err=%v", user.ID, key, err)
		}
		http.Error(w, "Failed to delete preference", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
