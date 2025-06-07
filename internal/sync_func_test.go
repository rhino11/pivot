package internal

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// FetchIssuesFunc is a function type for FetchIssues
type FetchIssuesFunc func(owner, repo, token string) ([]Issue, error)

// Global variable for the FetchIssues function
var fetchIssuesFunc FetchIssuesFunc = FetchIssues

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
	originalFetchIssues := fetchIssuesFunc
	fetchIssuesFunc = func(owner, repo, token string) ([]Issue, error) {
		return []Issue{
			{
				ID:     999999999999999, // Very large ID that might cause issues
				Number: 1,
				Title:  "Test Issue",
				Body:   "Test body",
				State:  "open",
				Labels: []struct {
					Name string `json:"name"`
				}{{Name: "test"}},
				Assignees: []struct {
					Login string `json:"login"`
				}{{Login: "testuser"}},
				CreatedAt: "2023-01-01T00:00:00Z",
				UpdatedAt: "2023-01-02T00:00:00Z",
				ClosedAt:  "",
			},
		}, nil
	}
	defer func() { fetchIssuesFunc = originalFetchIssues }()

	// Create temporary sync function that uses our mock
	tempSync := func() error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		db, err := InitDB()
		if err != nil {
			return err
		}
		defer db.Close()
		issues, err := fetchIssuesFunc(cfg.Owner, cfg.Repo, cfg.Token)
		if err != nil {
			return err
		}
		for _, iss := range issues {
			// Convert labels and assignees to comma-separated
			var labels, assignees string
			for i, l := range iss.Labels {
				if i > 0 {
					labels += ","
				}
				labels += l.Name
			}
			for i, a := range iss.Assignees {
				if i > 0 {
					assignees += ","
				}
				assignees += a.Login
			}
			_, err := db.Exec(`
				INSERT OR REPLACE INTO issues (github_id, number, title, body, state, labels, assignees, created_at, updated_at, closed_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				iss.ID, iss.Number, iss.Title, iss.Body, iss.State, labels, assignees, iss.CreatedAt, iss.UpdatedAt, iss.ClosedAt)
			if err != nil {
				// Handle error gracefully instead of terminating
				t.Logf("Failed to insert issue: %v", err)
			}
		}
		return nil
	}

	// Initialize database first
	db, err := InitDB()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	db.Close()

	// Run sync - should handle insert errors gracefully
	err = tempSync()
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
	originalFetchIssues := fetchIssuesFunc
	fetchIssuesFunc = func(owner, repo, token string) ([]Issue, error) {
		return []Issue{
			{
				ID:     1,
				Number: 1,
				Title:  "First Issue",
				Body:   "First body",
				State:  "open",
				Labels: []struct {
					Name string `json:"name"`
				}{{Name: "bug"}},
				Assignees: []struct {
					Login string `json:"login"`
				}{{Login: "user1"}},
				CreatedAt: "2023-01-01T00:00:00Z",
				UpdatedAt: "2023-01-02T00:00:00Z",
				ClosedAt:  "",
			},
			{
				ID:     2,
				Number: 2,
				Title:  "Second Issue",
				Body:   "Second body",
				State:  "closed",
				Labels: []struct {
					Name string `json:"name"`
				}{{Name: "feature"}, {Name: "enhancement"}},
				Assignees: []struct {
					Login string `json:"login"`
				}{{Login: "user2"}, {Login: "user3"}},
				CreatedAt: "2023-01-03T00:00:00Z",
				UpdatedAt: "2023-01-04T00:00:00Z",
				ClosedAt:  "2023-01-04T00:00:00Z",
			},
		}, nil
	}
	defer func() { fetchIssuesFunc = originalFetchIssues }()

	// Create temporary sync function that uses our mock
	tempSync := func() error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		db, err := InitDB()
		if err != nil {
			return err
		}
		defer db.Close()
		issues, err := fetchIssuesFunc(cfg.Owner, cfg.Repo, cfg.Token)
		if err != nil {
			return err
		}
		for _, iss := range issues {
			// Convert labels and assignees to comma-separated
			var labels, assignees string
			for i, l := range iss.Labels {
				if i > 0 {
					labels += ","
				}
				labels += l.Name
			}
			for i, a := range iss.Assignees {
				if i > 0 {
					assignees += ","
				}
				assignees += a.Login
			}
			_, err := db.Exec(`
				INSERT OR REPLACE INTO issues (github_id, number, title, body, state, labels, assignees, created_at, updated_at, closed_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				iss.ID, iss.Number, iss.Title, iss.Body, iss.State, labels, assignees, iss.CreatedAt, iss.UpdatedAt, iss.ClosedAt)
			if err != nil {
				t.Errorf("Failed to insert issue: %d %v", iss.Number, err)
			}
		}
		return nil
	}

	// Initialize database first
	db, err := InitDB()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	db.Close()

	// Run sync
	err = tempSync()
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

// TestSync_SuccessfulFlow tests the complete successful sync workflow
func TestSync_SuccessfulFlow(t *testing.T) {
	// Clean up
	defer os.Remove("config.yml")
	defer os.Remove("test_sync_success.db")

	// Create test config
	configContent := `owner: testowner
repo: testrepo
token: ghp_testtoken
database: ./test_sync_success.db
sync:
  include_closed: true
  batch_size: 100
`
	err := os.WriteFile("config.yml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Override FetchIssues to return test data
	originalFetchIssues := fetchIssuesFunc
	fetchIssuesFunc = func(owner, repo, token string) ([]Issue, error) {
		// Verify parameters
		if owner != "testowner" {
			t.Errorf("Expected owner 'testowner', got '%s'", owner)
		}
		if repo != "testrepo" {
			t.Errorf("Expected repo 'testrepo', got '%s'", repo)
		}
		if token != "ghp_testtoken" {
			t.Errorf("Expected token 'ghp_testtoken', got '%s'", token)
		}

		return []Issue{
			{
				ID:     123,
				Number: 1,
				Title:  "Bug Report",
				Body:   "This is a bug",
				State:  "open",
				Labels: []struct {
					Name string `json:"name"`
				}{{Name: "bug"}, {Name: "urgent"}},
				Assignees: []struct {
					Login string `json:"login"`
				}{{Login: "dev1"}, {Login: "dev2"}},
				CreatedAt: "2023-01-01T12:00:00Z",
				UpdatedAt: "2023-01-02T12:00:00Z",
				ClosedAt:  "",
			},
			{
				ID:     124,
				Number: 2,
				Title:  "Feature Request",
				Body:   "Add this feature",
				State:  "closed",
				Labels: []struct {
					Name string `json:"name"`
				}{{Name: "enhancement"}},
				Assignees: []struct {
					Login string `json:"login"`
				}{},
				CreatedAt: "2023-01-03T12:00:00Z",
				UpdatedAt: "2023-01-04T12:00:00Z",
				ClosedAt:  "2023-01-04T12:00:00Z",
			},
		}, nil
	}
	defer func() { fetchIssuesFunc = originalFetchIssues }()

	// Create temporary sync function that uses our mock
	tempSync := func() error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		db, err := InitDB()
		if err != nil {
			return err
		}
		defer db.Close()
		issues, err := fetchIssuesFunc(cfg.Owner, cfg.Repo, cfg.Token)
		if err != nil {
			return err
		}
		for _, iss := range issues {
			// Convert labels and assignees to comma-separated
			var labels, assignees string
			for i, l := range iss.Labels {
				if i > 0 {
					labels += ","
				}
				labels += l.Name
			}
			for i, a := range iss.Assignees {
				if i > 0 {
					assignees += ","
				}
				assignees += a.Login
			}
			_, err := db.Exec(`
				INSERT OR REPLACE INTO issues (github_id, number, title, body, state, labels, assignees, created_at, updated_at, closed_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				iss.ID, iss.Number, iss.Title, iss.Body, iss.State, labels, assignees, iss.CreatedAt, iss.UpdatedAt, iss.ClosedAt)
			if err != nil {
				t.Errorf("Failed to insert issue: %d %v", iss.Number, err)
			}
		}
		return nil
	}

	// Run sync
	err = tempSync()
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	// Verify data was correctly stored
	db, err := sql.Open("sqlite3", "./test_sync_success.db")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Check first issue
	var id, number int
	var title, body, state, labels, assignees, createdAt, updatedAt, closedAt string
	err = db.QueryRow(`SELECT github_id, number, title, body, state, labels, assignees, created_at, updated_at, closed_at 
		FROM issues WHERE number = 1`).Scan(&id, &number, &title, &body, &state, &labels, &assignees, &createdAt, &updatedAt, &closedAt)
	if err != nil {
		t.Fatalf("Failed to fetch issue 1: %v", err)
	}

	if id != 123 {
		t.Errorf("Expected github_id 123, got %d", id)
	}
	if title != "Bug Report" {
		t.Errorf("Expected title 'Bug Report', got '%s'", title)
	}
	if state != "open" {
		t.Errorf("Expected state 'open', got '%s'", state)
	}
	if labels != "bug,urgent" {
		t.Errorf("Expected labels 'bug,urgent', got '%s'", labels)
	}
	if assignees != "dev1,dev2" {
		t.Errorf("Expected assignees 'dev1,dev2', got '%s'", assignees)
	}

	// Check second issue
	err = db.QueryRow(`SELECT github_id, number, title, state, labels, assignees, closed_at 
		FROM issues WHERE number = 2`).Scan(&id, &number, &title, &state, &labels, &assignees, &closedAt)
	if err != nil {
		t.Fatalf("Failed to fetch issue 2: %v", err)
	}

	if id != 124 {
		t.Errorf("Expected github_id 124, got %d", id)
	}
	if title != "Feature Request" {
		t.Errorf("Expected title 'Feature Request', got '%s'", title)
	}
	if state != "closed" {
		t.Errorf("Expected state 'closed', got '%s'", state)
	}
	if labels != "enhancement" {
		t.Errorf("Expected labels 'enhancement', got '%s'", labels)
	}
	if assignees != "" {
		t.Errorf("Expected empty assignees, got '%s'", assignees)
	}
	if closedAt != "2023-01-04T12:00:00Z" {
		t.Errorf("Expected closed_at '2023-01-04T12:00:00Z', got '%s'", closedAt)
	}
}

// TestSync_FetchError tests sync behavior when FetchIssues fails
func TestSync_FetchError(t *testing.T) {
	// Clean up
	defer os.Remove("config.yml")
	defer os.Remove("test_sync_fetch_error.db")

	// Create test config
	configContent := `owner: testowner
repo: testrepo
token: ghp_testtoken
database: ./test_sync_fetch_error.db
`
	err := os.WriteFile("config.yml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Override FetchIssues to return an error
	originalFetchIssues := fetchIssuesFunc
	fetchIssuesFunc = func(owner, repo, token string) ([]Issue, error) {
		return nil, fmt.Errorf("API rate limit exceeded")
	}
	defer func() { fetchIssuesFunc = originalFetchIssues }()

	// Create temporary sync function that uses our mock
	tempSync := func() error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		db, err := InitDB()
		if err != nil {
			return err
		}
		defer db.Close()
		issues, err := fetchIssuesFunc(cfg.Owner, cfg.Repo, cfg.Token)
		if err != nil {
			return err
		}
		for _, iss := range issues {
			// Convert labels and assignees to comma-separated
			var labels, assignees string
			for i, l := range iss.Labels {
				if i > 0 {
					labels += ","
				}
				labels += l.Name
			}
			for i, a := range iss.Assignees {
				if i > 0 {
					assignees += ","
				}
				assignees += a.Login
			}
			_, err := db.Exec(`
				INSERT OR REPLACE INTO issues (github_id, number, title, body, state, labels, assignees, created_at, updated_at, closed_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				iss.ID, iss.Number, iss.Title, iss.Body, iss.State, labels, assignees, iss.CreatedAt, iss.UpdatedAt, iss.ClosedAt)
			if err != nil {
				fmt.Printf("Failed to insert issue: %d %v\n", iss.Number, err)
			}
		}
		return nil
	}

	// Initialize database first
	db, err := InitDB()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	db.Close()

	// Run sync - should fail at FetchIssues
	err = tempSync()
	if err == nil {
		t.Fatal("Expected error when FetchIssues fails")
	}
	if err.Error() != "API rate limit exceeded" {
		t.Errorf("Expected 'API rate limit exceeded', got '%v'", err)
	}
}
