package handlers

import (
	"encoding/json"
	"net/http"
)

// UploadFile handles CSV file upload and processing
func (h *Handlers) UploadFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"message": "File upload endpoint - coming soon",
		"status":  "not_implemented",
	}
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(response)
}