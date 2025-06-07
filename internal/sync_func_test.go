package internal

import (
	"os"
	"strings"
	"testing"
)

// TestSync_ConfigError tests sync behavior when config is missing
func TestSync_ConfigError(t *testing.T) {
	// Clean up
	defer os.Remove("config.yml")
	defer os.Remove("config.yaml")

	// Run sync without config file
	err := Sync()
	if err == nil {
		t.Error("Expected error when config file is missing")
	}
	if !strings.Contains(err.Error(), "no such file or directory") && !strings.Contains(err.Error(), "cannot find") {
		t.Errorf("Expected file not found error, got: %v", err)
	}
}

// TestSync_DatabaseError tests sync behavior when database initialization fails
func TestSync_DatabaseError(t *testing.T) {
	// Clean up
	defer os.Remove("config.yml")

	// Create config with invalid database path
	configContent := `owner: testowner
repo: testrepo
token: ghp_testtoken
database: /invalid/path/test.db
`
	err := os.WriteFile("config.yml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Run sync
	err = Sync()
	if err == nil {
		t.Error("Expected error when database path is invalid")
	}
}

// TestSync_LoadsConfig tests that Sync properly loads the configuration
func TestSync_LoadsConfig(t *testing.T) {
	// Clean up
	defer os.Remove("config.yml")
	defer os.Remove("test_sync_loads.db")

	// Create test config
	configContent := `owner: testowner
repo: testrepo
token: ghp_testtoken
database: ./test_sync_loads.db
sync:
  include_closed: false
  batch_size: 50
`
	err := os.WriteFile("config.yml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Initialize database first to avoid database errors
	db, err := InitDB()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	db.Close()

	// Test that loadConfig is called correctly by trying to run sync
	// This will fail on FetchIssues (GitHub API call) but should succeed in loading config
	err = Sync()
	// We expect an error from the GitHub API call, but not from config loading
	if err != nil && strings.Contains(err.Error(), "no such file or directory") {
		t.Errorf("Config loading failed, got: %v", err)
	}
	// Any other error (like API errors) is expected for this test
}

// TestSync_DatabaseInsertError tests sync behavior when database insertion fails
func TestSync_DatabaseInsertError(t *testing.T) {
	// Clean up
	defer os.Remove("config.yml")
	defer os.Remove("test_sync_insert_error.db")

	// Create test config
	configContent := `owner: testowner
repo: testrepo
token: ghp_testtoken
database: ./test_sync_insert_error.db
`
	err := os.WriteFile("config.yml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Override FetchIssues to return a problematic issue
	originalFetchIssues := FetchIssues
	FetchIssues = func(owner, repo, token string) ([]Issue, error) {
		return []Issue{
			{
				ID:        999999999999999, // Very large ID that might cause issues
				Number:    1,
				Title:     "Test Issue",
				Body:      "Test body",
				State:     "open",
				Labels:    []struct{ Name string }{{Name: "test"}},
				Assignees: []struct{ Login string }{{Login: "testuser"}},
				CreatedAt: "2023-01-01T00:00:00Z",
				UpdatedAt: "2023-01-02T00:00:00Z",
				ClosedAt:  "",
			},
		}, nil
	}
	defer func() { FetchIssues = originalFetchIssues }()

	// Initialize database first
	db, err := InitDB()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	db.Close()

	// Run sync - should handle insert errors gracefully
	err = Sync()
	if err != nil {
		t.Fatalf("Sync should handle insert errors gracefully, got: %v", err)
	}
}

// TestSync_WithMultipleIssues tests sync with multiple issues
func TestSync_WithMultipleIssues(t *testing.T) {
	// Clean up
	defer os.Remove("config.yml")
	defer os.Remove("test_sync_multiple.db")

	// Create test config
	configContent := `owner: testowner
repo: testrepo
token: ghp_testtoken
database: ./test_sync_multiple.db
`
	err := os.WriteFile("config.yml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Override FetchIssues to return multiple issues
	originalFetchIssues := FetchIssues
	FetchIssues = func(owner, repo, token string) ([]Issue, error) {
		return []Issue{
			{
				ID:        1,
				Number:    1,
				Title:     "First Issue",
				Body:      "First body",
				State:     "open",
				Labels:    []struct{ Name string }{{Name: "bug"}},
				Assignees: []struct{ Login string }{{Login: "user1"}},
				CreatedAt: "2023-01-01T00:00:00Z",
				UpdatedAt: "2023-01-02T00:00:00Z",
				ClosedAt:  "",
			},
			{
				ID:        2,
				Number:    2,
				Title:     "Second Issue",
				Body:      "Second body",
				State:     "closed",
				Labels:    []struct{ Name string }{{Name: "feature"}, {Name: "enhancement"}},
				Assignees: []struct{ Login string }{{Login: "user2"}, {Login: "user3"}},
				CreatedAt: "2023-01-03T00:00:00Z",
				UpdatedAt: "2023-01-04T00:00:00Z",
				ClosedAt:  "2023-01-04T00:00:00Z",
			},
		}, nil
	}
	defer func() { FetchIssues = originalFetchIssues }()

	// Initialize database first
	db, err := InitDB()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	db.Close()

	// Run sync
	err = Sync()
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	// Verify both issues were synced
	db, err = sql.Open("sqlite3", "./test_sync_multiple.db")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM issues").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count issues: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 issues in database, got %d", count)
	}

	// Check first issue
	var title1, labels1, assignees1 string
	err = db.QueryRow("SELECT title, labels, assignees FROM issues WHERE number = 1").Scan(&title1, &labels1, &assignees1)
	if err != nil {
		t.Fatalf("Failed to fetch first issue: %v", err)
	}

	if title1 != "First Issue" {
		t.Errorf("Expected first issue title 'First Issue', got '%s'", title1)
	}
	if labels1 != "bug" {
		t.Errorf("Expected labels 'bug', got '%s'", labels1)
	}
	if assignees1 != "user1" {
		t.Errorf("Expected assignees 'user1', got '%s'", assignees1)
	}

	// Check second issue
	var title2, labels2, assignees2 string
	err = db.QueryRow("SELECT title, labels, assignees FROM issues WHERE number = 2").Scan(&title2, &labels2, &assignees2)
	if err != nil {
		t.Fatalf("Failed to fetch second issue: %v", err)
	}

	if title2 != "Second Issue" {
		t.Errorf("Expected second issue title 'Second Issue', got '%s'", title2)
	}
	if labels2 != "feature,enhancement" {
		t.Errorf("Expected labels 'feature,enhancement', got '%s'", labels2)
	}
	if assignees2 != "user2,user3" {
		t.Errorf("Expected assignees 'user2,user3', got '%s'", assignees2)
	}
}
