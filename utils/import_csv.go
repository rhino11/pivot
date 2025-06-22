package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rhino11/pivot/internal"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run cmd/import_csv.go <csv_file> [project_id]")
		fmt.Println("  csv_file    - Path to the CSV file to import")
		fmt.Println("  project_id  - Project ID to associate issues with (defaults to 1)")
		os.Exit(1)
	}

	csvFile := os.Args[1]
	
	// Default project ID to 1 if not specified
	projectID := int64(1)
	if len(os.Args) > 2 {
		if parsed, err := strconv.ParseInt(os.Args[2], 10, 64); err == nil {
			projectID = parsed
		} else {
			log.Fatalf("Invalid project_id: %s", os.Args[2])
		}
	}

	// Open database connection
	db, err := internal.InitDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Ensure sync state schema exists
	if err := internal.InitSyncStateSchema(db); err != nil {
		log.Fatalf("Failed to initialize sync state schema: %v", err)
	}

	// Check if project exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM projects WHERE id = ?", projectID).Scan(&count)
	if err != nil {
		log.Fatalf("Failed to check project existence: %v", err)
	}
	if count == 0 {
		fmt.Printf("⚠️  Project ID %d not found. Creating default project...\n", projectID)
		// Create a default project for CSV imports
		_, err = db.Exec(`
			INSERT OR REPLACE INTO projects (id, owner, repo, path, created_at, updated_at)
			VALUES (?, 'local', 'csv-import', '.', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`, projectID)
		if err != nil {
			log.Fatalf("Failed to create default project: %v", err)
		}
		fmt.Printf("✅ Created default project (ID: %d)\n", projectID)
	}

	// Open and parse CSV file
	file, err := os.Open(csvFile)
	if err != nil {
		log.Fatalf("Error opening CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Error reading CSV: %v", err)
	}

	if len(records) == 0 {
		log.Fatal("CSV file is empty")
	}

	// Parse header row to get column indices
	headers := records[0]
	headerMap := make(map[string]int)
	for i, header := range headers {
		headerMap[strings.ToLower(strings.TrimSpace(header))] = i
	}

	// Required columns
	requiredCols := []string{"title"}
	for _, col := range requiredCols {
		if _, exists := headerMap[col]; !exists {
			log.Fatalf("Required column '%s' not found in CSV", col)
		}
	}

	fmt.Printf("📋 Processing %d issues from CSV for project ID %d...\n", len(records)-1, projectID)

	// Process each data row
	inserted := 0
	for i, record := range records[1:] {
		if len(record) != len(headers) {
			fmt.Printf("Warning: Row %d has different number of columns than header\n", i+2)
			continue
		}

		// Extract fields
		var githubID *int64
		if idStr := getField(record, headerMap, "id"); idStr != "" {
			if parsed, err := strconv.ParseInt(idStr, 10, 64); err == nil {
				githubID = &parsed
			}
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
		assignees := getField(record, headerMap, "assignees")
		if assignees == "" {
			assignees = getField(record, headerMap, "assignee") // fallback to singular
		}
		
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

		// Generate a number for the issue (use row index if not provided)
		number := i + 1
		if numberStr := getField(record, headerMap, "number"); numberStr != "" {
			if parsed, err := strconv.Atoi(numberStr); err == nil {
				number = parsed
			}
		}

		// Insert issue into database
		var issueLocalID int64
		if githubID != nil {
			// Issue has GitHub ID - insert with it
			_, err = db.Exec(`
				INSERT OR REPLACE INTO issues 
				(github_id, project_id, number, title, body, state, labels, assignees, created_at, updated_at, closed_at, local_modified_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				*githubID, projectID, number, title, body, state, labels, assignees, createdAt, updatedAt, closedAt, now)
			
			if err != nil {
				fmt.Printf("Error inserting row %d: %v\n", i+2, err)
				continue
			}
			issueLocalID = *githubID
		} else {
			// Local-only issue - insert without GitHub ID
			result, err := db.Exec(`
				INSERT INTO issues 
				(project_id, number, title, body, state, labels, assignees, created_at, updated_at, closed_at, local_modified_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				projectID, number, title, body, state, labels, assignees, createdAt, updatedAt, closedAt, now)
			
			if err != nil {
				fmt.Printf("Error inserting row %d: %v\n", i+2, err)
				continue
			}
			
			// Get the auto-generated ID (for local-only issues, we need the rowid)
			lastID, err := result.LastInsertId()
			if err != nil {
				fmt.Printf("Error getting last insert ID for row %d: %v\n", i+2, err)
				continue
			}
			issueLocalID = lastID
		}

		// Create sync state record
		var syncState internal.SyncState
		if githubID != nil {
			// Issue came from GitHub or has GitHub ID - mark as synced
			syncState = internal.SyncStateSynced
		} else {
			// Local-only issue imported from CSV - mark as local only
			syncState = internal.SyncStateLocalOnly
		}

		err = internal.CreateSyncState(db, issueLocalID, syncState, githubID)
		if err != nil {
			fmt.Printf("Warning: Failed to create sync state for row %d: %v\n", i+2, err)
			// Continue anyway - the issue was inserted
		}

		inserted++
		if inserted%10 == 0 {
			fmt.Printf("✓ Imported %d issues...\n", inserted)
		}
	}

	fmt.Printf("🎉 Successfully imported %d issues to local database!\n", inserted)
	
	// Show sync state summary
	summary, err := internal.GetSyncStateSummary(db)
	if err != nil {
		fmt.Printf("Warning: Failed to get sync state summary: %v\n", err)
	} else {
		fmt.Println("\n📊 Sync State Summary:")
		for state, count := range summary {
			fmt.Printf("  %s: %d issues\n", state, count)
		}
	}
	
	fmt.Println("\nNext steps:")
	fmt.Println("  • Run 'pivot status' to see sync state summary")
	fmt.Println("  • Run 'pivot push' to push LOCAL_ONLY issues to GitHub")
	fmt.Println("  • Run 'pivot sync' to synchronize with GitHub")
}

func getField(record []string, headerMap map[string]int, fieldName string) string {
	// Try lowercase version first
	if idx, exists := headerMap[strings.ToLower(fieldName)]; exists && idx < len(record) {
		return strings.TrimSpace(record[idx])
	}
	// Try exact match as fallback
	if idx, exists := headerMap[fieldName]; exists && idx < len(record) {
		return strings.TrimSpace(record[idx])
	}
	return ""
}
