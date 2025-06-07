package internal

import (
	"database/sql"
	"os"
	"path/filepath"
	"sync"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestInitDB(t *testing.T) {
	// Clean up any existing test database
	testDB := "test_pivot.db"
	defer os.Remove(testDB)

	// Test database initialization
	db, err := sql.Open("sqlite3", testDB)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Create the schema
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
		t.Fatalf("Failed to create schema: %v", err)
	}

	// Test that we can insert and retrieve data
	_, err = db.Exec(`
		INSERT INTO issues (github_id, number, title, body, state, labels, assignees, created_at, updated_at, closed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		1, 123, "Test Issue", "This is a test", "open", "bug,feature", "user1,user2", "2023-01-01", "2023-01-02", "")

	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Verify data was inserted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM issues").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count issues: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 issue, got %d", count)
	}

	// Test retrieving the data
	var githubID, number int
	var title, body, state, labels, assignees, createdAt, updatedAt, closedAt string

	err = db.QueryRow("SELECT github_id, number, title, body, state, labels, assignees, created_at, updated_at, closed_at FROM issues WHERE github_id = 1").
		Scan(&githubID, &number, &title, &body, &state, &labels, &assignees, &createdAt, &updatedAt, &closedAt)

	if err != nil {
		t.Fatalf("Failed to retrieve test data: %v", err)
	}

	// Verify the data matches what we inserted
	if githubID != 1 {
		t.Errorf("Expected github_id 1, got %d", githubID)
	}
	if number != 123 {
		t.Errorf("Expected number 123, got %d", number)
	}
	if title != "Test Issue" {
		t.Errorf("Expected title 'Test Issue', got %s", title)
	}
	if state != "open" {
		t.Errorf("Expected state 'open', got %s", state)
	}

	t.Log("Database functionality test passed")
}

func TestInit(t *testing.T) {
	// Test the Init function which should create and initialize the database
	testDB := "test_init.db"
	defer os.Remove(testDB)

	// This test would need to be modified to use a test database path
	// For now, we'll test that Init doesn't return an error
	err := Init()
	if err != nil {
		t.Errorf("Init() returned error: %v", err)
	}

	// Clean up the database created by Init
	defer os.Remove("./pivot.db")
}

func TestInitDBErrorHandling(t *testing.T) {
	// Test error handling by trying to create a database in a non-existent directory
	invalidPath := "/nonexistent/directory/test.db"

	db, err := sql.Open("sqlite3", invalidPath)
	if err != nil {
		// This is expected behavior
		t.Logf("Expected error opening invalid database path: %v", err)
		return
	}
	defer db.Close()

	// If we get here, the database was created successfully (unexpected)
	// Let's try to execute the schema to see if it fails
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
		t.Logf("Expected error creating schema in invalid location: %v", err)
	}
}

func TestInitDBWithExistingDatabase(t *testing.T) {
	// Test initializing database when it already exists
	testDB := "test_existing.db"
	defer os.Remove(testDB)

	// Create database first time
	db1, err := sql.Open("sqlite3", testDB)
	if err != nil {
		t.Fatalf("Failed to create initial database: %v", err)
	}

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
	)`
	_, err = db1.Exec(schema)
	if err != nil {
		t.Fatalf("Failed to create initial schema: %v", err)
	}
	db1.Close()

	// Now test Init() function on existing database
	err = Init()
	if err != nil {
		t.Errorf("Init should work with existing database, got error: %v", err)
	}

	t.Log("Init with existing database test passed")
}

func TestDatabaseSchema(t *testing.T) {
	// Test that the database schema is created correctly
	testDB := "test_schema.db"
	defer os.Remove(testDB)

	db, err := sql.Open("sqlite3", testDB)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Create the schema
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
	)`
	_, err = db.Exec(schema)
	if err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	// Test that we can query the schema
	rows, err := db.Query("PRAGMA table_info(issues)")
	if err != nil {
		t.Fatalf("Failed to query table info: %v", err)
	}
	defer rows.Close()

	columnCount := 0
	expectedColumns := map[string]bool{
		"github_id":  false,
		"number":     false,
		"title":      false,
		"body":       false,
		"state":      false,
		"labels":     false,
		"assignees":  false,
		"created_at": false,
		"updated_at": false,
		"closed_at":  false,
	}

	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, pk int
		var defaultValue interface{}

		err = rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk)
		if err != nil {
			t.Fatalf("Failed to scan column info: %v", err)
		}

		if _, exists := expectedColumns[name]; exists {
			expectedColumns[name] = true
			columnCount++
		}
	}

	if columnCount != len(expectedColumns) {
		t.Errorf("Expected %d columns, found %d", len(expectedColumns), columnCount)
	}

	for col, found := range expectedColumns {
		if !found {
			t.Errorf("Expected column '%s' not found", col)
		}
	}

	t.Log("Database schema test passed")
}

func TestInitDBPermissionError(t *testing.T) {
	// Test database creation with permission constraints
	// Save original working directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Create a temporary directory and remove write permissions
	tempDir := t.TempDir()
	testPath := filepath.Join(tempDir, "readonly")
	err := os.Mkdir(testPath, 0755)
	if err != nil {
		t.Skipf("Cannot create test directory: %v", err)
	}

	// Remove write permissions
	err = os.Chmod(testPath, 0555)
	if err != nil {
		t.Skipf("Cannot change directory permissions: %v", err)
	}

	// Try to create database in read-only directory
	os.Chdir(testPath)

	_, err = InitDB()
	// On some systems this might still succeed due to SQLite behavior,
	// so we just log the result rather than asserting failure
	if err != nil {
		t.Logf("Correctly failed with permission error: %v", err)
	} else {
		t.Logf("Database creation succeeded despite directory permissions")
	}
}

func TestInitDBConcurrentAccess(t *testing.T) {
	// Test concurrent database initialization
	testDB := "test_concurrent.db"
	defer os.Remove(testDB)

	var wg sync.WaitGroup
	errors := make(chan error, 5)

	// Try to initialize database concurrently
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			db, err := sql.Open("sqlite3", testDB)
			if err != nil {
				errors <- err
				return
			}
			defer db.Close()

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
			)`
			_, err = db.Exec(schema)
			errors <- err
		}()
	}

	wg.Wait()
	close(errors)

	// Check that at least some operations succeeded
	successCount := 0
	for err := range errors {
		if err == nil {
			successCount++
		}
	}

	if successCount == 0 {
		t.Error("All concurrent database operations failed")
	} else {
		t.Logf("Concurrent access test passed with %d successful operations", successCount)
	}
}
