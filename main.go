package main

import (
	"etl-api/database"
	"etl-api/handlers"
	"etl-api/middleware"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	// Initialize database connection
	db, err := database.InitDB()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Run database migrations
	if err := database.RunMigrations(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Initialize handlers with database connection
	h := handlers.NewHandlers(db)

	// Create router
	r := mux.NewRouter()

	// Health check endpoint
	r.HandleFunc("/health", handlers.HealthCheck).Methods("GET")

	// Authentication routes
	auth := r.PathPrefix("/auth").Subrouter()
	auth.HandleFunc("/register", h.Register).Methods("POST")
	auth.HandleFunc("/login", h.Login).Methods("POST")

	// Protected routes - require JWT authentication
	protected := r.PathPrefix("").Subrouter()
	protected.Use(middleware.JWTAuth)

	// File upload and data management routes
	protected.HandleFunc("/upload", h.UploadFile).Methods("POST")
	protected.HandleFunc("/tables", h.ListTables).Methods("GET")
	protected.HandleFunc("/tables/{id}", h.DeleteTable).Methods("DELETE")
	protected.HandleFunc("/data/{table_id}", h.GetTableData).Methods("GET")

	// Apply CORS middleware to all routes
	handler := middleware.CORS(r)

	// Get port from environment or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 ETL API Server starting on port %s", port)
	log.Printf("📊 Database connected and migrations complete")
	log.Printf("🔐 JWT authentication enabled")
	log.Printf("📁 File upload ready (max size: 10MB)")

	// Start server
	log.Fatal(http.ListenAndServe(":"+port, handler))
}