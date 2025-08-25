package handlers

import (
	"encoding/json"
	"etl-api/middleware"
	"etl-api/models"
	"etl-api/utils"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// UploadFile handles CSV file upload and processing
func (h *Handlers) UploadFile(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user ID
	userID := middleware.GetUserIDFromContext(r)
	if userID == "" {
		http.Error(w, `{"error": "Authentication required"}`, http.StatusUnauthorized)
		return
	}

	// Parse multipart form
	maxFileSize := int64(10 << 20) // 10MB default
	if sizeStr := os.Getenv("MAX_FILE_SIZE"); sizeStr != "" {
		if size, err := strconv.ParseInt(sizeStr, 10, 64); err == nil {
			maxFileSize = size
		}
	}

	if err := r.ParseMultipartForm(maxFileSize); err != nil {
		http.Error(w, `{"error": "File too large or invalid form data"}`, http.StatusBadRequest)
		return
	}

	// Get file from form
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, `{"error": "No file provided"}`, http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file type
	if !strings.HasSuffix(strings.ToLower(fileHeader.Filename), ".csv") {
		http.Error(w, `{"error": "Only CSV files are supported"}`, http.StatusBadRequest)
		return
	}

	// Get table name from form
	tableName := r.FormValue("table_name")
	if err := utils.ValidateTableName(tableName); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	// Parse CSV file
	csvData, err := utils.ParseCSV(file)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to parse CSV: %s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	// Generate physical table name
	physicalTableName := utils.SanitizeTableName(userID, tableName)

	// Create dynamic table
	if err := h.createDynamicTable(physicalTableName, csvData.Headers); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to create table: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	// Insert data into dynamic table
	rowsInserted, err := h.insertCSVData(physicalTableName, csvData)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to insert data: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	// Store table metadata
	tableSchema := make(map[string]interface{})
	for _, col := range csvData.Headers {
		tableSchema[col.Name] = map[string]string{
			"type":   col.DataType,
			"sample": col.Sample,
		}
	}

	schemaJSON, err := json.Marshal(tableSchema)
	if err != nil {
		http.Error(w, `{"error": "Failed to serialize table schema"}`, http.StatusInternalServerError)
		return
	}

	var tableID string
	err = h.db.QueryRow(`
		INSERT INTO data_tables (user_id, table_name, original_filename, column_count, row_count, table_schema, physical_table_name)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, userID, tableName, fileHeader.Filename, len(csvData.Headers), rowsInserted, schemaJSON, physicalTableName).Scan(&tableID)

	if err != nil {
		http.Error(w, `{"error": "Failed to store table metadata"}`, http.StatusInternalServerError)
		return
	}

	// Return success response
	columnNames := make([]string, len(csvData.Headers))
	for i, col := range csvData.Headers {
		columnNames[i] = col.Name
	}

	response := models.UploadResponse{
		TableID:      tableID,
		TableName:    tableName,
		Filename:     fileHeader.Filename,
		RowsImported: rowsInserted,
		Columns:      columnNames,
		Message:      "Data imported successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// createDynamicTable creates a PostgreSQL table based on CSV structure
func (h *Handlers) createDynamicTable(tableName string, columns []utils.CSVColumn) error {
	var columnDefs []string
	columnDefs = append(columnDefs, "id SERIAL PRIMARY KEY")
	
	for _, col := range columns {
		columnDef := fmt.Sprintf(`"%s" %s`, col.Name, col.DataType)
		columnDefs = append(columnDefs, columnDef)
	}
	columnDefs = append(columnDefs, "created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP")

	query := fmt.Sprintf(`CREATE TABLE "%s" (%s)`, tableName, strings.Join(columnDefs, ", "))
	
	_, err := h.db.Exec(query)
	return err
}

// insertCSVData inserts CSV data into the dynamic table
func (h *Handlers) insertCSVData(tableName string, csvData *utils.CSVData) (int, error) {
	if len(csvData.Rows) == 0 {
		return 0, nil
	}

	// Build column names for INSERT
	columnNames := make([]string, len(csvData.Headers))
	for i, col := range csvData.Headers {
		columnNames[i] = fmt.Sprintf(`"%s"`, col.Name)
	}

	// Build placeholders
	placeholders := make([]string, len(csvData.Headers))
	for i := range placeholders {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	query := fmt.Sprintf(`INSERT INTO "%s" (%s) VALUES (%s)`,
		tableName,
		strings.Join(columnNames, ", "),
		strings.Join(placeholders, ", "))

	// Insert rows in batches
	insertedCount := 0
	for _, row := range csvData.Rows {
		// Ensure row has the correct number of columns
		values := make([]interface{}, len(csvData.Headers))
		for i := 0; i < len(csvData.Headers); i++ {
			if i < len(row) {
				values[i] = row[i]
			} else {
				values[i] = nil
			}
		}

		_, err := h.db.Exec(query, values...)
		if err != nil {
			return insertedCount, fmt.Errorf("failed to insert row %d: %v", insertedCount+1, err)
		}
		insertedCount++
	}

	return insertedCount, nil
}