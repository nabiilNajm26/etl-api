package utils

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// CSVColumn represents a column in CSV data
type CSVColumn struct {
	Name     string
	DataType string
	Sample   string
}

// CSVData represents parsed CSV data
type CSVData struct {
	Headers []CSVColumn
	Rows    [][]string
}

// ParseCSV reads and parses CSV data from a reader
func ParseCSV(reader io.Reader) (*CSVData, error) {
	csvReader := csv.NewReader(reader)
	csvReader.FieldsPerRecord = -1 // Allow variable number of fields

	// Read all records
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %v", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("CSV file is empty")
	}

	// Extract headers
	headers := records[0]
	if len(headers) == 0 {
		return nil, fmt.Errorf("CSV file has no headers")
	}

	// Clean and validate headers
	cleanHeaders := make([]string, len(headers))
	for i, header := range headers {
		cleaned := strings.TrimSpace(header)
		if cleaned == "" {
			cleaned = fmt.Sprintf("column_%d", i+1)
		}
		// Make header PostgreSQL-safe
		cleaned = SanitizeColumnName(cleaned)
		cleanHeaders[i] = cleaned
	}

	// Get data rows (skip header)
	dataRows := records[1:]
	if len(dataRows) == 0 {
		return nil, fmt.Errorf("CSV file contains only headers, no data")
	}

	// Infer column data types
	columns := make([]CSVColumn, len(cleanHeaders))
	for i, header := range cleanHeaders {
		dataType := inferColumnType(dataRows, i)
		sample := ""
		if len(dataRows) > 0 && i < len(dataRows[0]) {
			sample = dataRows[0][i]
		}
		
		columns[i] = CSVColumn{
			Name:     header,
			DataType: dataType,
			Sample:   sample,
		}
	}

	return &CSVData{
		Headers: columns,
		Rows:    dataRows,
	}, nil
}

// SanitizeColumnName creates PostgreSQL-safe column name
func SanitizeColumnName(name string) string {
	// Replace spaces with underscores and convert to lowercase
	sanitized := strings.ReplaceAll(name, " ", "_")
	sanitized = strings.ToLower(sanitized)
	
	// Remove special characters, keep only alphanumeric and underscores
	var result strings.Builder
	for _, char := range sanitized {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '_' {
			result.WriteRune(char)
		}
	}
	
	sanitized = result.String()
	
	// Ensure it starts with a letter or underscore
	if len(sanitized) > 0 && sanitized[0] >= '0' && sanitized[0] <= '9' {
		sanitized = "col_" + sanitized
	}
	
	// Ensure it's not empty
	if sanitized == "" {
		sanitized = "unnamed_column"
	}
	
	return sanitized
}

// inferColumnType analyzes sample data to determine PostgreSQL data type
func inferColumnType(rows [][]string, columnIndex int) string {
	if len(rows) == 0 {
		return "TEXT"
	}

	samples := 0
	intCount := 0
	floatCount := 0
	dateCount := 0
	totalValues := 0

	// Sample up to 100 rows for type inference
	sampleSize := len(rows)
	if sampleSize > 100 {
		sampleSize = 100
	}

	for i := 0; i < sampleSize; i++ {
		if columnIndex >= len(rows[i]) {
			continue
		}
		
		value := strings.TrimSpace(rows[i][columnIndex])
		if value == "" {
			continue
		}
		
		totalValues++
		
		// Check if it's an integer
		if _, err := strconv.Atoi(value); err == nil {
			intCount++
		} else if _, err := strconv.ParseFloat(value, 64); err == nil {
			floatCount++
		}
		
		// Check if it's a date
		if isDateString(value) {
			dateCount++
		}
		
		samples++
	}

	if samples == 0 {
		return "TEXT"
	}

	// If 80% or more are integers, use INTEGER
	if float64(intCount)/float64(samples) >= 0.8 {
		return "INTEGER"
	}

	// If 80% or more are numbers (int or float), use NUMERIC
	if float64(intCount+floatCount)/float64(samples) >= 0.8 {
		return "NUMERIC"
	}

	// If 80% or more are dates, use DATE
	if float64(dateCount)/float64(samples) >= 0.8 {
		return "DATE"
	}

	// Default to TEXT
	return "TEXT"
}

// isDateString checks if a string could be a date
func isDateString(value string) bool {
	dateFormats := []string{
		"2006-01-02",
		"01/02/2006",
		"1/2/2006",
		"2006/01/02",
		"2006-1-2",
		"01-02-2006",
		"2 Jan 2006",
		"January 2, 2006",
	}

	for _, format := range dateFormats {
		if _, err := time.Parse(format, value); err == nil {
			return true
		}
	}

	return false
}