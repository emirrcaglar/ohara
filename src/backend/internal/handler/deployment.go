package handler

import (
	"encoding/json"
	"net/http"

	"ohara/src/internal/db"
	"ohara/src/internal/logger"
)

type DeploymentHandler struct {
	DB  *db.DB
	Log *logger.Logger
}

func (h *DeploymentHandler) HandleLatest(w http.ResponseWriter, r *http.Request) {
	deployment, err := h.DB.GetLatestDeployment()
	if err != nil {
		if h.Log != nil {
			h.Log.Error("[deployments] failed to fetch latest deployment err=%v", err)
		}
		http.Error(w, "Failed to fetch latest deployment", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"deployment": deployment})
}
