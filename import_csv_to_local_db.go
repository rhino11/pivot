package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Simple function to import CSV directly to local database without GitHub API
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run import_csv_to_local_db.go <csv_file>")
		os.Exit(1)
	}

	csvFile := os.Args[1]
	
	// Open the CSV file
	file, err := os.Open(csvFile)
	if err != nil {
		fmt.Printf("Error opening CSV file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Parse CSV
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Printf("Error reading CSV: %v\n", err)
		os.Exit(1)
	}

	if len(records) == 0 {
		fmt.Println("CSV file is empty")
		os.Exit(1)
	}

	// Open local database
	db, err := sql.Open("sqlite3", "./pivot.db")
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Create table if it doesn't exist
	schema := `
	CREATE TABLE IF NOT EXISTS issues (
		github_id INTEGER PRIMARY KEY,
		number INTEGER,
		title TEXT,
		body TEXT,
		state TEXT,
		labels TEXT,
		assignees TEXT,
		created_at TEXT,
		updated_at TEXT,
		closed_at TEXT
	);`
	
	_, err = db.Exec(schema)
	if err != nil {
		fmt.Printf("Error creating schema: %v\n", err)
		os.Exit(1)
	}

	// Parse header row to get column indices
	headers := records[0]
	headerMap := make(map[string]int)
	for i, header := range headers {
		headerMap[strings.TrimSpace(header)] = i
	}

	// Required columns
	requiredCols := []string{"id", "title", "state"}
	for _, col := range requiredCols {
		if _, exists := headerMap[col]; !exists {
			fmt.Printf("Required column '%s' not found in CSV\n", col)
			os.Exit(1)
		}
	}

	fmt.Printf("📋 Processing %d issues from CSV...\n", len(records)-1)

	// Process each data row
	inserted := 0
	for i, record := range records[1:] {
		if len(record) != len(headers) {
			fmt.Printf("Warning: Row %d has different number of columns than header\n", i+2)
			continue
		}

		// Extract fields
		var githubID int64
		if idStr := getField(record, headerMap, "id"); idStr != "" {
			if parsed, err := strconv.ParseInt(idStr, 10, 64); err == nil {
				githubID = parsed
			} else {
				githubID = int64(i + 1) // Use row number as fallback
			}
		} else {
			githubID = int64(i + 1)
		}

		title := getField(record, headerMap, "title")
		if title == "" {
			fmt.Printf("Warning: Row %d missing title, skipping\n", i+2)
			continue
		}

		state := getField(record, headerMap, "state")
		if state == "" {
			state = "open"
		}

		body := getField(record, headerMap, "body")
		labels := getField(record, headerMap, "labels")
		assignee := getField(record, headerMap, "assignee")
		
		// Use current time if timestamps not provided
		now := time.Now().Format(time.RFC3339)
		createdAt := getField(record, headerMap, "created_at")
		if createdAt == "" {
			createdAt = now
		}
		updatedAt := getField(record, headerMap, "updated_at")
		if updatedAt == "" {
			updatedAt = now
		}
		closedAt := getField(record, headerMap, "closed_at")

		// Insert into database
		_, err := db.Exec(`
			INSERT OR REPLACE INTO issues (github_id, number, title, body, state, labels, assignees, created_at, updated_at, closed_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			githubID, githubID, title, body, state, labels, assignee, createdAt, updatedAt, closedAt)
		
		if err != nil {
			fmt.Printf("Error inserting row %d: %v\n", i+2, err)
			continue
		}
		
		inserted++
		if inserted%5 == 0 {
			fmt.Printf("✓ Imported %d issues...\n", inserted)
		}
	}

	fmt.Printf("🎉 Successfully imported %d issues to local database!\n", inserted)
	fmt.Println("Run 'pivot export csv' to verify the import")
}

func getField(record []string, headerMap map[string]int, fieldName string) string {
	if idx, exists := headerMap[fieldName]; exists && idx < len(record) {
		return strings.TrimSpace(record[idx])
	}
	return ""
}
