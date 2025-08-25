package models

import (
	"time"
)

// DataTable represents metadata about uploaded data tables
type DataTable struct {
	ID                string                 `json:"id" db:"id"`
	UserID            string                 `json:"user_id" db:"user_id"`
	TableName         string                 `json:"table_name" db:"table_name"`
	OriginalFilename  string                 `json:"original_filename" db:"original_filename"`
	ColumnCount       int                    `json:"column_count" db:"column_count"`
	RowCount          int                    `json:"row_count" db:"row_count"`
	TableSchema       map[string]interface{} `json:"table_schema" db:"table_schema"`
	PhysicalTableName string                 `json:"physical_table_name" db:"physical_table_name"`
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
}

// UploadResponse represents file upload response
type UploadResponse struct {
	TableID   string   `json:"table_id"`
	TableName string   `json:"table_name"`
	Filename  string   `json:"filename"`
	RowsImported int   `json:"rows_imported"`
	Columns   []string `json:"columns"`
	Message   string   `json:"message"`
}

// TableListResponse represents response for listing tables
type TableListResponse struct {
	Tables []DataTableSummary `json:"tables"`
	Total  int                `json:"total"`
}

// DataTableSummary represents summary info for table listing
type DataTableSummary struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Filename         string    `json:"filename"`
	Rows             int       `json:"rows"`
	Columns          int       `json:"columns"`
	CreatedAt        time.Time `json:"created_at"`
}

// DataResponse represents data retrieval response
type DataResponse struct {
	TableInfo  DataTableInfo            `json:"table_info"`
	Data       []map[string]interface{} `json:"data"`
	Pagination PaginationInfo           `json:"pagination"`
}

// DataTableInfo represents table information in data response
type DataTableInfo struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	TotalRows  int      `json:"total_rows"`
	Columns    []string `json:"columns"`
}

// PaginationInfo represents pagination information
type PaginationInfo struct {
	CurrentPage int  `json:"current_page"`
	PerPage     int  `json:"per_page"`
	TotalPages  int  `json:"total_pages"`
	HasNext     bool `json:"has_next"`
}