package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// Handlers holds dependencies for HTTP handlers
type Handlers struct {
	db *sql.DB
}

// NewHandlers creates a new handlers instance
func NewHandlers(db *sql.DB) *Handlers {
	return &Handlers{
		db: db,
	}
}

// HealthCheck returns API health status
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"status":  "healthy",
		"message": "ETL API is running",
		"version": "1.0.0",
	}
	json.NewEncoder(w).Encode(response)
}