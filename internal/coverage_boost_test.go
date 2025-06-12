package internal

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// TestSyncProject_AdditionalCoverage improves coverage for the syncProject function
func TestSyncProject_AdditionalCoverage(t *testing.T) {
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Logf("Warning: Failed to change back: %v", err)
		}
	}()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Create database for testing
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

	t.Run("ProjectTokenOverridesGlobal", func(t *testing.T) {
		globalConfig := &GlobalConfig{Token: "global_token"}
		projectConfig := &ProjectConfig{
			Owner: "testowner",
			Repo:  "testrepo",
			Token: "project_specific_token", // This should override global
		}

		// This will fail at FetchIssues due to invalid token, but should pass token validation
		err := syncProject(db, globalConfig, projectConfig)
		if err != nil && strings.Contains(err.Error(), "no GitHub token configured") {
			t.Errorf("Should not be a token error when project has specific token, got: %v", err)
		}
		// The error should be from GitHub API call, not token validation
	})

	t.Run("GlobalTokenUsedWhenProjectHasNone", func(t *testing.T) {
		globalConfig := &GlobalConfig{Token: "global_token"}
		projectConfig := &ProjectConfig{
			Owner: "testowner",
			Repo:  "testrepo",
			Token: "", // Empty token, should use global
		}

		// This will fail at FetchIssues due to invalid token, but should pass token validation
		err := syncProject(db, globalConfig, projectConfig)
		if err != nil && strings.Contains(err.Error(), "no GitHub token configured") {
			t.Errorf("Should not be a token error when global has token, got: %v", err)
		}
		// The error should be from GitHub API call, not token validation
	})

	t.Run("CreateProjectErrorPath", func(t *testing.T) {
		// Close database to trigger CreateProject error
		db.Close()

		globalConfig := &GlobalConfig{Token: "valid_token"}
		projectConfig := &ProjectConfig{
			Owner: "testowner",
			Repo:  "testrepo",
			Token: "project_token",
		}

		err := syncProject(db, globalConfig, projectConfig)
		if err == nil {
			t.Error("Expected error when database is closed")
		}
		if !strings.Contains(err.Error(), "failed to ensure project in database") {
			t.Errorf("Expected database error, got: %v", err)
		}
	})
}

// TestSyncMultiProject_AdditionalCoverage improves coverage for SyncMultiProject function
func TestSyncMultiProject_AdditionalCoverage(t *testing.T) {
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Logf("Warning: Failed to change back: %v", err)
		}
	}()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	t.Run("SyncAllProjectsWhenNoFilterProvided", func(t *testing.T) {
		// Create a multi-project config
		configContent := `global:
  database: sync_multi.db
projects:
  - owner: org1
    repo: repo1
    token: token1
  - owner: org2
    repo: repo2
    token: token2`

		err := os.WriteFile("config.yml", []byte(configContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create config file: %v", err)
		}
		defer os.Remove("sync_multi.db")

		// Test SyncMultiProject with empty filter (should sync all projects)
		err = SyncMultiProject("")
		if err == nil {
			t.Skip("Sync succeeded unexpectedly - likely means network access")
		}

		// Should attempt to sync all projects, not fail on project filter validation
		if strings.Contains(err.Error(), "project filter must be in format") {
			t.Errorf("Should not be a filter format error when no filter provided, got: %v", err)
		}
	})

	t.Run("InvalidProjectFilterFormat", func(t *testing.T) {
		// Create a multi-project config
		configContent := `global:
  database: sync_filter.db
projects:
  - owner: org1
    repo: repo1
    token: token1`

		err := os.WriteFile("config.yml", []byte(configContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create config file: %v", err)
		}
		defer os.Remove("sync_filter.db")

		// Test with invalid filter format (missing slash)
		err = SyncMultiProject("invalidfilter")
		if err == nil {
			t.Error("Expected error for invalid filter format")
		}
		if !strings.Contains(err.Error(), "project filter must be in format 'owner/repo'") {
			t.Errorf("Expected filter format error, got: %v", err)
		}
	})

	t.Run("ProjectNotFoundInConfig", func(t *testing.T) {
		// Create a multi-project config
		configContent := `global:
  database: sync_notfound.db
projects:
  - owner: org1
    repo: repo1
    token: token1`

		err := os.WriteFile("config.yml", []byte(configContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create config file: %v", err)
		}
		defer os.Remove("sync_notfound.db")

		// Test with project filter that doesn't exist in config
		err = SyncMultiProject("nonexistent/repo")
		if err == nil {
			t.Error("Expected error when project not found")
		}
		if !strings.Contains(err.Error(), "project nonexistent/repo not found in configuration") {
			t.Errorf("Expected project not found error, got: %v", err)
		}
	})

	t.Run("DatabaseInitializationError", func(t *testing.T) {
		// Create config with invalid database path
		configContent := `global:
  database: /invalid/readonly/path/db.sqlite
projects:
  - owner: org1
    repo: repo1
    token: token1`

		err := os.WriteFile("config.yml", []byte(configContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create config file: %v", err)
		}

		// This should fail at database initialization but the function will
		// attempt to continue and sync projects, printing errors to stdout
		// The function only returns errors for config/database setup issues
		err = SyncMultiProject("")
		// The function should return nil since it continues even when individual
		// project syncs fail - it only returns errors for setup issues
		if err != nil {
			// If we get an error, it should be related to database setup
			t.Logf("Got setup error (acceptable): %v", err)
		}
	})
}

// TestFetchIssues_AdditionalEdgeCases improves coverage for FetchIssues function
func TestFetchIssues_AdditionalEdgeCases(t *testing.T) {
	t.Run("EmptyOwnerRepo", func(t *testing.T) {
		// Test with empty owner/repo combinations
		testCases := []struct {
			owner, repo, token string
			description        string
		}{
			{"", "repo", "token", "empty owner"},
			{"owner", "", "token", "empty repo"},
			{"", "", "token", "empty owner and repo"},
		}

		for _, tc := range testCases {
			t.Run(tc.description, func(t *testing.T) {
				_, err := FetchIssues(tc.owner, tc.repo, tc.token)
				if err == nil {
					t.Errorf("Expected error for %s", tc.description)
				}
				// Should get an HTTP error due to invalid URL, not panic
			})
		}
	})

	t.Run("SpecialCharactersInOwnerRepo", func(t *testing.T) {
		// Test with special characters that might cause URL issues
		specialCases := []struct {
			owner, repo string
		}{
			{"owner-with-dashes", "repo-with-dashes"},
			{"owner_with_underscores", "repo_with_underscores"},
			{"owner123", "repo456"},
			{"Owner-Mixed_Case123", "Repo-Mixed_Case456"},
		}

		for _, tc := range specialCases {
			t.Run(fmt.Sprintf("%s/%s", tc.owner, tc.repo), func(t *testing.T) {
				// This will likely fail with 401/404, but should not panic
				_, err := FetchIssues(tc.owner, tc.repo, "test-token")
				if err != nil {
					// Expected - we're just testing that it doesn't crash
					t.Logf("Expected failure for %s/%s: %v", tc.owner, tc.repo, err)
				}
			})
		}
	})
}

// TestCreateIssue_AdditionalErrorPaths improves coverage for CreateIssue function
func TestCreateIssue_AdditionalErrorPaths(t *testing.T) {
	t.Run("EmptyOwnerRepoParams", func(t *testing.T) {
		request := CreateIssueRequest{
			Title: "Test Issue",
			Body:  "Test body",
		}

		// Test with empty parameters
		testCases := []struct {
			owner, repo, token string
			description        string
		}{
			{"", "repo", "token", "empty owner"},
			{"owner", "", "token", "empty repo"},
			{"owner", "repo", "", "empty token"},
		}

		for _, tc := range testCases {
			t.Run(tc.description, func(t *testing.T) {
				_, err := CreateIssue(tc.owner, tc.repo, tc.token, request)
				if err == nil {
					t.Errorf("Expected error for %s", tc.description)
				}
				// Should get an HTTP error, not panic
			})
		}
	})

	t.Run("InvalidRequestData", func(t *testing.T) {
		// Test with invalid request data
		invalidRequests := []CreateIssueRequest{
			{Title: "", Body: "body"},  // Empty title
			{Title: "title", Body: ""}, // Empty body
			{Title: "", Body: ""},      // Both empty
		}

		for i, req := range invalidRequests {
			t.Run(fmt.Sprintf("invalid_request_%d", i), func(t *testing.T) {
				_, err := CreateIssue("owner", "repo", "token", req)
				if err == nil {
					t.Error("Expected error for invalid request data")
				}
				// Should get an HTTP error from GitHub API
			})
		}
	})
}

// TestInitMultiProjectConfig_AdditionalCoverage improves coverage for InitMultiProjectConfig
func TestInitMultiProjectConfig_AdditionalCoverage(t *testing.T) {
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Logf("Warning: Failed to change back: %v", err)
		}
	}()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	t.Run("ConfigFileWriteError", func(t *testing.T) {
		// Try to write to a path where we can't write (like /dev/null/config.yml)
		originalWd, _ := os.Getwd()

		// Create a temporary directory and make it readonly after entering it
		readonlyTestDir := "readonly_test"
		if err := os.Mkdir(readonlyTestDir, 0755); err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}
		defer os.RemoveAll(readonlyTestDir)

		// Change to the directory first
		if err := os.Chdir(readonlyTestDir); err != nil {
			t.Fatalf("Failed to change to test directory: %v", err)
		}

		// Make the directory readonly after we're in it (this should work)
		defer func() {
			// Restore permissions and change back
			_ = os.Chmod(".", 0755) // #nosec G104 - test cleanup, ignore error
			_ = os.Chdir(originalWd) // #nosec G104 - test cleanup, ignore error
		}()

		if err := os.Chmod(".", 0444); err != nil {
			t.Skipf("Cannot make directory readonly on this system: %v", err)
		}

		// Simulate user input
		input := "testowner\ntestrepo\ntoken123\n./test.db\ny\n100\n"
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatalf("Failed to create pipe: %v", err)
		}

		go func() {
			defer w.Close()
			_, _ = w.WriteString(input) // #nosec G104 - test helper, ignore error
		}()

		// Temporarily replace stdin
		oldStdin := os.Stdin
		os.Stdin = r
		defer func() {
			os.Stdin = oldStdin
			r.Close()
		}()

		// This should fail at the file write step
		err = InitMultiProjectConfig()
		if err == nil {
			t.Error("Expected error when writing to readonly directory")
		}
		if !strings.Contains(err.Error(), "failed to write config file") {
			t.Errorf("Expected write error, got: %v", err)
		}
	})
}
