package handlers

import (
	"encoding/json"
	"net/http"
)

// Health answers GET /v1/health with liveness only (specs/api/openapi.yaml, getHealth).
func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "OK"})
}
