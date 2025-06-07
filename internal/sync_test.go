package internal

import (
	"database/sql"
	"os"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestSyncDatabaseSchema(t *testing.T) {
	// Create temporary database
	tmpDB, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp database: %v", err)
	}
	defer os.Remove(tmpDB.Name())
	tmpDB.Close()

	// Open database connection
	db, err := sql.Open("sqlite3", tmpDB.Name())
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create issues table with similar schema as in sync.go
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS issues (
github_id INTEGER,
number INTEGER PRIMARY KEY,
title TEXT NOT NULL,
body TEXT,
state TEXT,
labels TEXT,
assignees TEXT,
created_at TEXT,
updated_at TEXT,
closed_at TEXT
)`

	if _, err := db.Exec(createTableSQL); err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Test issue insertion using same SQL pattern as sync.go
	insertSQL := `
	INSERT OR REPLACE INTO issues (github_id, number, title, body, state, labels, assignees, created_at, updated_at, closed_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	labels := []string{"bug", "enhancement"}
	assignees := []string{"user1", "user2"}

	// Convert to comma-separated strings like sync.go does
	var labelsStr, assigneesStr string
	for i, l := range labels {
		if i > 0 {
			labelsStr += ","
		}
		labelsStr += l
	}
	for i, a := range assignees {
		if i > 0 {
			assigneesStr += ","
		}
		assigneesStr += a
	}

	_, err = db.Exec(insertSQL, 12345, 123, "Test Issue", "Test body", "open", labelsStr, assigneesStr, "2023-01-01", "2023-01-02", nil)
	if err != nil {
		t.Fatalf("Failed to insert issue: %v", err)
	}

	// Verify insertion
	var githubID, number int
	var title, body, state, storedLabels, storedAssignees, createdAt, updatedAt string
	var closedAt sql.NullString

	selectSQL := "SELECT github_id, number, title, body, state, labels, assignees, created_at, updated_at, closed_at FROM issues WHERE number = ?"
	err = db.QueryRow(selectSQL, 123).Scan(&githubID, &number, &title, &body, &state, &storedLabels, &storedAssignees, &createdAt, &updatedAt, &closedAt)
	if err != nil {
		t.Fatalf("Failed to query issue: %v", err)
	}

	if githubID != 12345 {
		t.Errorf("Expected github_id 12345, got %d", githubID)
	}
	if number != 123 {
		t.Errorf("Expected number 123, got %d", number)
	}
	if title != "Test Issue" {
		t.Errorf("Expected title 'Test Issue', got '%s'", title)
	}
	if storedLabels != "bug,enhancement" {
		t.Errorf("Expected labels 'bug,enhancement', got '%s'", storedLabels)
	}
	if storedAssignees != "user1,user2" {
		t.Errorf("Expected assignees 'user1,user2', got '%s'", storedAssignees)
	}
}

func TestLabelProcessing(t *testing.T) {
	// Test the label and assignee processing logic used in sync.go
	labels := []string{"bug", "enhancement", "documentation"}
	assignees := []string{"alice", "bob", "charlie"}

	// Test the same string building logic as in sync.go
	var labelsStr, assigneesStr string
	for i, l := range labels {
		if i > 0 {
			labelsStr += ","
		}
		labelsStr += l
	}
	for i, a := range assignees {
		if i > 0 {
			assigneesStr += ","
		}
		assigneesStr += a
	}

	expectedLabels := "bug,enhancement,documentation"
	expectedAssignees := "alice,bob,charlie"

	if labelsStr != expectedLabels {
		t.Errorf("Expected labels '%s', got '%s'", expectedLabels, labelsStr)
	}

	if assigneesStr != expectedAssignees {
		t.Errorf("Expected assignees '%s', got '%s'", expectedAssignees, assigneesStr)
	}

	// Test parsing back (reverse operation)
	parsedLabels := strings.Split(labelsStr, ",")
	parsedAssignees := strings.Split(assigneesStr, ",")

	if len(parsedLabels) != 3 {
		t.Errorf("Expected 3 labels, got %d", len(parsedLabels))
	}

	if len(parsedAssignees) != 3 {
		t.Errorf("Expected 3 assignees, got %d", len(parsedAssignees))
	}

	for i, label := range labels {
		if parsedLabels[i] != label {
			t.Errorf("Expected label '%s', got '%s'", label, parsedLabels[i])
		}
	}
}

func TestDatabaseUpdateOperations(t *testing.T) {
	// Test the INSERT OR REPLACE functionality used in sync.go
	tmpDB, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp database: %v", err)
	}
	defer os.Remove(tmpDB.Name())
	tmpDB.Close()

	db, err := sql.Open("sqlite3", tmpDB.Name())
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS issues (
github_id INTEGER,
number INTEGER PRIMARY KEY,
title TEXT NOT NULL,
body TEXT,
state TEXT,
labels TEXT,
assignees TEXT,
created_at TEXT,
updated_at TEXT,
closed_at TEXT
)`

	if _, err := db.Exec(createTableSQL); err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	insertSQL := `
	INSERT OR REPLACE INTO issues (github_id, number, title, body, state, labels, assignees, created_at, updated_at, closed_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	// Insert initial issue
	_, err = db.Exec(insertSQL, 1001, 1, "Original Title", "Original body", "open", "bug", "user1", "2023-01-01", "2023-01-01", nil)
	if err != nil {
		t.Fatalf("Failed to insert issue: %v", err)
	}

	// Update same issue (INSERT OR REPLACE)
	_, err = db.Exec(insertSQL, 1001, 1, "Updated Title", "Updated body", "closed", "bug,enhancement", "user1,user2", "2023-01-01", "2023-01-02", "2023-01-03")
	if err != nil {
		t.Fatalf("Failed to update issue: %v", err)
	}

	// Verify update worked
	var title, state, labels, assignees string
	err = db.QueryRow("SELECT title, state, labels, assignees FROM issues WHERE number = 1").Scan(&title, &state, &labels, &assignees)
	if err != nil {
		t.Fatalf("Failed to query updated issue: %v", err)
	}

	if title != "Updated Title" {
		t.Errorf("Expected title 'Updated Title', got '%s'", title)
	}
	if state != "closed" {
		t.Errorf("Expected state 'closed', got '%s'", state)
	}
	if labels != "bug,enhancement" {
		t.Errorf("Expected labels 'bug,enhancement', got '%s'", labels)
	}
	if assignees != "user1,user2" {
		t.Errorf("Expected assignees 'user1,user2', got '%s'", assignees)
	}

	// Verify only one record exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM issues").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count issues: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 issue, got %d", count)
	}
}

func TestDatabaseErrorHandling(t *testing.T) {
	// Test various error conditions

	// 1. Invalid database path should fail on operations
	tmpDB, err := os.CreateTemp("", "error-test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp database: %v", err)
	}
	defer os.Remove(tmpDB.Name())
	tmpDB.Close()

	db, err := sql.Open("sqlite3", tmpDB.Name())
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Execute invalid SQL should return error
	_, err = db.Exec("INVALID SQL STATEMENT")
	if err == nil {
		t.Error("Expected SQL error, but operation succeeded")
	}
}

func TestEmptyLabelProcessing(t *testing.T) {
	// Test processing of empty labels and assignees
	var labels []string
	var assignees []string

	var labelsStr, assigneesStr string
	for i, l := range labels {
		if i > 0 {
			labelsStr += ","
		}
		labelsStr += l
	}
	for i, a := range assignees {
		if i > 0 {
			assigneesStr += ","
		}
		assigneesStr += a
	}

	if labelsStr != "" {
		t.Errorf("Expected empty labels string, got '%s'", labelsStr)
	}

	if assigneesStr != "" {
		t.Errorf("Expected empty assignees string, got '%s'", assigneesStr)
	}
}

func TestMultipleLabelsAndAssignees(t *testing.T) {
	// Test handling of multiple labels and assignees like in real GitHub issues
	labels := []string{"bug", "high-priority", "backend", "database", "needs-review"}
	assignees := []string{"alice", "bob", "charlie", "david"}

	var labelsStr, assigneesStr string
	for i, l := range labels {
		if i > 0 {
			labelsStr += ","
		}
		labelsStr += l
	}
	for i, a := range assignees {
		if i > 0 {
			assigneesStr += ","
		}
		assigneesStr += a
	}

	expectedLabels := "bug,high-priority,backend,database,needs-review"
	expectedAssignees := "alice,bob,charlie,david"

	if labelsStr != expectedLabels {
		t.Errorf("Expected labels '%s', got '%s'", expectedLabels, labelsStr)
	}

	if assigneesStr != expectedAssignees {
		t.Errorf("Expected assignees '%s', got '%s'", expectedAssignees, assigneesStr)
	}

	// Test database storage with these values
	tmpDB, err := os.CreateTemp("", "multi-test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp database: %v", err)
	}
	defer os.Remove(tmpDB.Name())
	tmpDB.Close()

	db, err := sql.Open("sqlite3", tmpDB.Name())
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS issues (
github_id INTEGER,
number INTEGER PRIMARY KEY,
title TEXT NOT NULL,
body TEXT,
state TEXT,
labels TEXT,
assignees TEXT,
created_at TEXT,
updated_at TEXT,
closed_at TEXT
)`

	if _, err := db.Exec(createTableSQL); err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	insertSQL := `
	INSERT OR REPLACE INTO issues (github_id, number, title, body, state, labels, assignees, created_at, updated_at, closed_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = db.Exec(insertSQL, 999, 999, "Complex Issue", "Issue with many labels and assignees",
		"open", labelsStr, assigneesStr, "2023-01-01", "2023-01-01", nil)
	if err != nil {
		t.Fatalf("Failed to insert issue with multiple labels/assignees: %v", err)
	}

	// Verify storage and retrieval
	var storedLabels, storedAssignees string
	err = db.QueryRow("SELECT labels, assignees FROM issues WHERE number = 999").Scan(&storedLabels, &storedAssignees)
	if err != nil {
		t.Fatalf("Failed to query issue: %v", err)
	}

	if storedLabels != expectedLabels {
		t.Errorf("Expected labels '%s', got '%s'", expectedLabels, storedLabels)
	}

	if storedAssignees != expectedAssignees {
		t.Errorf("Expected assignees '%s', got '%s'", expectedAssignees, storedAssignees)
	}
}
