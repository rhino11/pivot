package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/rhino11/pivot/internal"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run cmd/migrate.go <command>")
		fmt.Println("Commands:")
		fmt.Println("  apply-sync-state - Apply sync state schema migrations")
		fmt.Println("  check-schema     - Check current database schema")
		os.Exit(1)
	}

	command := os.Args[1]

	// Open database connection
	db, err := internal.InitDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	switch command {
	case "apply-sync-state":
		if err := applySyncStateMigrations(db); err != nil {
			log.Fatalf("Failed to apply sync state migrations: %v", err)
		}
		fmt.Println("✅ Successfully applied sync state migrations")
	case "check-schema":
		if err := checkSchema(db); err != nil {
			log.Fatalf("Failed to check schema: %v", err)
		}
	default:
		log.Fatalf("Unknown command: %s", command)
	}
}

func applySyncStateMigrations(db *sql.DB) error {
	fmt.Println("🔄 Applying sync state schema migrations...")
	
	// Initialize the sync state schema
	if err := internal.InitSyncStateSchema(db); err != nil {
		return fmt.Errorf("failed to initialize sync state schema: %w", err)
	}
	
	fmt.Println("✅ Created issue_sync_state table")
	fmt.Println("✅ Added sync columns to issues table")
	
	return nil
}

func checkSchema(db *sql.DB) error {
	fmt.Println("📋 Current database schema:")
	
	// Get table schemas
	query := `SELECT sql FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name`
	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query schema: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var schema string
		if err := rows.Scan(&schema); err != nil {
			return fmt.Errorf("failed to scan schema: %w", err)
		}
		fmt.Println(schema + ";")
		fmt.Println()
	}
	
	// Check if sync state table exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='issue_sync_state'").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check for sync state table: %w", err)
	}
	
	if count > 0 {
		fmt.Println("✅ issue_sync_state table exists")
	} else {
		fmt.Println("❌ issue_sync_state table missing")
	}
	
	// Check if sync columns exist in issues table
	hasLocalModified, err := hasColumn(db, "issues", "local_modified_at")
	if err != nil {
		return fmt.Errorf("failed to check for local_modified_at column: %w", err)
	}
	
	hasSyncHash, err := hasColumn(db, "issues", "sync_hash")
	if err != nil {
		return fmt.Errorf("failed to check for sync_hash column: %w", err)
	}
	
	if hasLocalModified {
		fmt.Println("✅ local_modified_at column exists in issues table")
	} else {
		fmt.Println("❌ local_modified_at column missing from issues table")
	}
	
	if hasSyncHash {
		fmt.Println("✅ sync_hash column exists in issues table")
	} else {
		fmt.Println("❌ sync_hash column missing from issues table")
	}
	
	return nil
}

// hasColumn checks if a table has a specific column (helper function)
func hasColumn(db *sql.DB, tableName, columnName string) (bool, error) {
	query := `SELECT COUNT(*) FROM pragma_table_info(?) WHERE name = ?`
	var count int
	err := db.QueryRow(query, tableName, columnName).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
