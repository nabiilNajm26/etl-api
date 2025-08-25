package handlers

import (
	"database/sql"
	"encoding/json"
	"etl-api/middleware"
	"etl-api/models"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

// ListTables returns all tables for the authenticated user
func (h *Handlers) ListTables(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user ID
	userID := middleware.GetUserIDFromContext(r)
	if userID == "" {
		http.Error(w, `{"error": "Authentication required"}`, http.StatusUnauthorized)
		return
	}

	// Query user's tables
	rows, err := h.db.Query(`
		SELECT id, table_name, original_filename, column_count, row_count, created_at
		FROM data_tables 
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)

	if err != nil {
		http.Error(w, `{"error": "Failed to retrieve tables"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tables []models.DataTableSummary
	for rows.Next() {
		var table models.DataTableSummary
		if err := rows.Scan(&table.ID, &table.Name, &table.Filename, &table.Columns, &table.Rows, &table.CreatedAt); err != nil {
			http.Error(w, `{"error": "Failed to scan table data"}`, http.StatusInternalServerError)
			return
		}
		tables = append(tables, table)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}

	response := models.TableListResponse{
		Tables: tables,
		Total:  len(tables),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteTable removes a table and its data
func (h *Handlers) DeleteTable(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user ID
	userID := middleware.GetUserIDFromContext(r)
	if userID == "" {
		http.Error(w, `{"error": "Authentication required"}`, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	tableID := vars["id"]

	// Get table info to verify ownership and get physical table name
	var physicalTableName string
	err := h.db.QueryRow(`
		SELECT physical_table_name 
		FROM data_tables 
		WHERE id = $1 AND user_id = $2
	`, tableID, userID).Scan(&physicalTableName)

	if err == sql.ErrNoRows {
		http.Error(w, `{"error": "Table not found"}`, http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}

	// Drop the physical table
	_, err = h.db.Exec(fmt.Sprintf(`DROP TABLE IF EXISTS "%s"`, physicalTableName))
	if err != nil {
		http.Error(w, `{"error": "Failed to drop table"}`, http.StatusInternalServerError)
		return
	}

	// Delete metadata
	_, err = h.db.Exec(`DELETE FROM data_tables WHERE id = $1 AND user_id = $2`, tableID, userID)
	if err != nil {
		http.Error(w, `{"error": "Failed to delete table metadata"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"message":  "Table deleted successfully",
		"table_id": tableID,
	}
	json.NewEncoder(w).Encode(response)
}

// GetTableData retrieves data from a specific table with pagination
func (h *Handlers) GetTableData(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user ID
	userID := middleware.GetUserIDFromContext(r)
	if userID == "" {
		http.Error(w, `{"error": "Authentication required"}`, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	tableID := vars["table_id"]

	// Get pagination parameters
	page := 1
	limit := 100 // default limit

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	// Get table metadata
	var table models.DataTable
	var schemaJSON []byte
	err := h.db.QueryRow(`
		SELECT id, table_name, original_filename, column_count, row_count, table_schema, physical_table_name
		FROM data_tables 
		WHERE id = $1 AND user_id = $2
	`, tableID, userID).Scan(&table.ID, &table.TableName, &table.OriginalFilename, 
		&table.ColumnCount, &table.RowCount, &schemaJSON, &table.PhysicalTableName)

	if err == sql.ErrNoRows {
		http.Error(w, `{"error": "Table not found"}`, http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}

	// Parse table schema
	if err := json.Unmarshal(schemaJSON, &table.TableSchema); err != nil {
		http.Error(w, `{"error": "Failed to parse table schema"}`, http.StatusInternalServerError)
		return
	}

	// Get column names from schema
	var columns []string
	for columnName := range table.TableSchema {
		columns = append(columns, columnName)
	}

	// Build column list for query (with quotes for safety)
	quotedColumns := make([]string, len(columns))
	for i, col := range columns {
		quotedColumns[i] = fmt.Sprintf(`"%s"`, col)
	}

	// Calculate offset
	offset := (page - 1) * limit

	// Query data with pagination
	query := fmt.Sprintf(`
		SELECT %s 
		FROM "%s" 
		ORDER BY id 
		LIMIT $1 OFFSET $2
	`, strings.Join(quotedColumns, ", "), table.PhysicalTableName)

	rows, err := h.db.Query(query, limit, offset)
	if err != nil {
		http.Error(w, `{"error": "Failed to query table data"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Prepare data structure
	var data []map[string]interface{}
	
	for rows.Next() {
		// Create slice for values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		// Scan row
		if err := rows.Scan(valuePtrs...); err != nil {
			http.Error(w, `{"error": "Failed to scan row data"}`, http.StatusInternalServerError)
			return
		}

		// Create row map
		rowData := make(map[string]interface{})
		for i, column := range columns {
			value := values[i]
			// Handle null values and convert byte slices to strings
			if value == nil {
				rowData[column] = nil
			} else if b, ok := value.([]byte); ok {
				rowData[column] = string(b)
			} else {
				rowData[column] = value
			}
		}
		data = append(data, rowData)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}

	// Calculate pagination info
	totalPages := (table.RowCount + limit - 1) / limit
	hasNext := page < totalPages

	// Build response
	response := models.DataResponse{
		TableInfo: models.DataTableInfo{
			ID:        table.ID,
			Name:      table.TableName,
			TotalRows: table.RowCount,
			Columns:   columns,
		},
		Data: data,
		Pagination: models.PaginationInfo{
			CurrentPage: page,
			PerPage:     limit,
			TotalPages:  totalPages,
			HasNext:     hasNext,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}