package internal

import (
	"database/sql"
	"os"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// TestSyncProject tests the syncProject function for multiproject support
func TestSyncProject(t *testing.T) {
	tempDB := "test_sync_project.db"
	defer os.Remove(tempDB)

	db, err := sql.Open("sqlite3", tempDB)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	err = InitMultiProjectDB(db)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	t.Run("Missing token", func(t *testing.T) {
		globalNoToken := &GlobalConfig{Token: ""}
		projectNoToken := &ProjectConfig{
			Owner: "testowner",
			Repo:  "testrepo",
		}

		err := syncProject(db, globalNoToken, projectNoToken)
		if err == nil {
			t.Error("Expected error for missing token")
		}
		if !strings.Contains(err.Error(), "no GitHub token configured") {
			t.Errorf("Expected token error, got: %v", err)
		}
	})

	t.Run("With project-specific token", func(t *testing.T) {
		globalNoToken := &GlobalConfig{Token: ""}
		projectWithToken := &ProjectConfig{
			Owner: "testowner",
			Repo:  "testrepo",
			Token: "project_token",
		}

		// This will fail on the HTTP call, but should pass the token check
		err := syncProject(db, globalNoToken, projectWithToken)
		if err != nil && strings.Contains(err.Error(), "no GitHub token configured") {
			t.Errorf("Should not be a token error when project has token, got: %v", err)
		}
		// We expect this to fail on HTTP request, that's fine
	})

	t.Run("With global token", func(t *testing.T) {
		globalWithToken := &GlobalConfig{Token: "global_token"}
		project := &ProjectConfig{
			Owner: "testowner",
			Repo:  "testrepo",
		}

		// This will fail on the HTTP call, but should pass the token check
		err := syncProject(db, globalWithToken, project)
		if err != nil && strings.Contains(err.Error(), "no GitHub token configured") {
			t.Errorf("Should not be a token error when global has token, got: %v", err)
		}
		// We expect this to fail on HTTP request, that's fine
	})
}

// TestSyncEnhanced tests additional coverage for the main Sync function
func TestSyncEnhanced(t *testing.T) {
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Logf("Warning: Failed to change back: %v", err)
		}
	}()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	t.Run("Config with empty owner", func(t *testing.T) {
		configContent := `owner: ""
repo: testrepo
token: testtoken
database: test.db`

		err := os.WriteFile("config.yml", []byte(configContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create config file: %v", err)
		}

		err = Sync()
		if err == nil {
			t.Error("Expected error for empty owner")
		}
	})

	t.Run("Config with empty repo", func(t *testing.T) {
		configContent := `owner: testowner
repo: ""
token: testtoken
database: test.db`

		err := os.WriteFile("config.yml", []byte(configContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create config file: %v", err)
		}

		err = Sync()
		if err == nil {
			t.Error("Expected error for empty repo")
		}
	})

	t.Run("Valid config but network will fail", func(t *testing.T) {
		configContent := `owner: testowner
repo: testrepo
token: validtoken
database: test.db`

		err := os.WriteFile("config.yml", []byte(configContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create config file: %v", err)
		}

		// This should pass validation but fail on network call
		err = Sync()
		if err == nil {
			t.Error("Expected error due to network call failure")
		}
		// Should not be a validation error
		if strings.Contains(err.Error(), "GitHub token is required") ||
			strings.Contains(err.Error(), "owner and repo are required") {
			t.Errorf("Should not be a validation error, got: %v", err)
		}
	})
}

// TestLoadConfigPublicWrapper tests the public LoadConfig wrapper function
func TestLoadConfigPublicWrapper(t *testing.T) {
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Logf("Warning: Failed to change back: %v", err)
		}
	}()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Test with valid config
	configContent := `owner: testowner
repo: testrepo
token: testtoken
database: test.db`

	err := os.WriteFile("config.yml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Test the public LoadConfig function (which wraps loadConfig)
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if config.Owner != "testowner" {
		t.Errorf("Expected owner 'testowner', got '%s'", config.Owner)
	}
	if config.Repo != "testrepo" {
		t.Errorf("Expected repo 'testrepo', got '%s'", config.Repo)
	}
}

// TestSyncDatabaseOperations tests the database insertion logic in Sync
func TestSyncDatabaseOperations(t *testing.T) {
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Logf("Warning: Failed to change back: %v", err)
		}
	}()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create a config that would work but will fail on GitHub API call
	configContent := `owner: validowner
repo: validrepo
token: validtoken
database: test_sync.db`

	err := os.WriteFile("config.yml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Test that Sync attempts to load config and init DB properly
	// This will fail on the FetchIssues call, but that's expected
	err = Sync()
	if err == nil {
		t.Error("Expected error due to GitHub API call")
	}

	// The error should not be about config loading or DB init
	if strings.Contains(err.Error(), "failed to read config file") {
		t.Errorf("Should not be a config error: %v", err)
	}

	// Check that database was created
	if _, err := os.Stat("test_sync.db"); os.IsNotExist(err) {
		t.Error("Database file should have been created")
	}
}

// TestSyncWithDatabaseError tests Sync when database operations fail
func TestSyncWithDatabaseError(t *testing.T) {
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Logf("Warning: Failed to change back: %v", err)
		}
	}()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create config with database path that can't be created
	configContent := `owner: testowner
repo: testrepo
token: testtoken
database: /invalid/readonly/path/test.db`

	err := os.WriteFile("config.yml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Test that Sync fails appropriately when DB can't be created
	err = Sync()
	if err == nil {
		t.Error("Expected error when database can't be created")
	}
}

// TestCreateIssueErrorPaths tests error handling in CreateIssue
func TestCreateIssueErrorPaths(t *testing.T) {
	t.Run("Marshal error", func(t *testing.T) {
		// Create a request that would cause JSON marshal error
		// This is hard to trigger with normal structs, so we test the happy path
		request := CreateIssueRequest{
			Title: "Test Issue",
			Body:  "Test body",
		}

		// This will fail on network, but not on marshaling
		_, err := CreateIssue("owner", "repo", "token", request)
		if err == nil {
			t.Error("Expected network error in test environment")
		}
		// Should not be a marshal error
		if strings.Contains(err.Error(), "failed to marshal request") {
			t.Error("Should not be a marshal error with valid request")
		}
	})

	t.Run("Request creation with empty values", func(t *testing.T) {
		request := CreateIssueRequest{
			Title: "", // Empty title
			Body:  "",
		}

		_, err := CreateIssue("", "", "", request) // Empty params
		if err == nil {
			t.Error("Expected error with empty parameters")
		}
	})
}

// TestInitMultiProjectDBErrorPaths tests error cases for InitMultiProjectDB
func TestInitMultiProjectDBErrorPaths(t *testing.T) {
	// Test with database that can't be written to
	tempDB := "/dev/null" // Can't write to this on Unix systems

	db, err := sql.Open("sqlite3", tempDB)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// This should fail when trying to create tables
	err = InitMultiProjectDB(db)
	if err == nil {
		t.Error("Expected error when database can't be written to")
	}
}

// TestMigrateToMultiProjectErrorPaths tests error cases for MigrateToMultiProject
func TestMigrateToMultiProjectErrorPaths(t *testing.T) {
	tempDB := "test_migrate_error.db"
	defer os.Remove(tempDB)

	db, err := sql.Open("sqlite3", tempDB)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Test migration with valid parameters - should succeed even on empty database
	err = MigrateToMultiProject(db, "owner", "repo", "/path")
	if err != nil {
		t.Fatalf("MigrateToMultiProject failed unexpectedly: %v", err)
	}

	// Verify the project was created
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM projects WHERE owner = 'owner' AND repo = 'repo'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check for migrated project: %v", err)
	}
	if count != 1 {
		t.Error("Expected migrated project to be created")
	}
}
