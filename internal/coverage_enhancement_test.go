package internal

import (
	"database/sql"
	"os"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// TestSync_EdgeCases tests additional sync scenarios
func TestSync_EdgeCases(t *testing.T) {
	// Test sync with no config file
	t.Run("no_config_file", func(t *testing.T) {
		// Remove any existing config files
		os.Remove("config.yml")
		os.Remove("config.yaml")

		err := Sync()
		if err == nil {
			t.Error("Expected error when no config file exists")
		}
		if !strings.Contains(err.Error(), "config") {
			t.Errorf("Expected config-related error, got: %v", err)
		}
	})

	// Test sync with invalid config file
	t.Run("invalid_config_file", func(t *testing.T) {
		// Create invalid config file
		invalidConfig := "invalid: yaml: content: ["
		err := os.WriteFile("config.yml", []byte(invalidConfig), 0644)
		if err != nil {
			t.Fatalf("Failed to create invalid config: %v", err)
		}
		defer os.Remove("config.yml")

		err = Sync()
		if err == nil {
			t.Error("Expected error with invalid config file")
		}
	})
}

// TestAddSyncColumnsToIssues_Coverage tests database migration scenarios
func TestAddSyncColumnsToIssues_Coverage(t *testing.T) {
	// Create temporary database
	tmpDB, err := os.CreateTemp("", "test-sync-*.db")
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

	// Test adding sync columns to table without existing columns
	t.Run("fresh_table", func(t *testing.T) {
		// Create issues table without sync columns
		_, err := db.Exec(`
			DROP TABLE IF EXISTS issues;
			CREATE TABLE issues (
				github_id INTEGER PRIMARY KEY,
				title TEXT,
				body TEXT,
				state TEXT
			);
		`)
		if err != nil {
			t.Fatalf("Failed to create test table: %v", err)
		}

		err = AddSyncColumnsToIssues(db)
		if err != nil {
			t.Errorf("AddSyncColumnsToIssues failed: %v", err)
		}

		// Verify columns were added
		rows, err := db.Query("PRAGMA table_info(issues)")
		if err != nil {
			t.Fatalf("Failed to query table info: %v", err)
		}
		defer rows.Close()

		foundLocalModified := false
		foundSyncHash := false

		for rows.Next() {
			var cid int
			var name, dataType string
			var notNull int
			var defaultValue interface{}
			var pk int

			err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk)
			if err != nil {
				t.Fatalf("Failed to scan column info: %v", err)
			}

			if name == "local_modified_at" {
				foundLocalModified = true
			}
			if name == "sync_hash" {
				foundSyncHash = true
			}
		}

		if !foundLocalModified {
			t.Error("Expected local_modified_at column to be added")
		}
		if !foundSyncHash {
			t.Error("Expected sync_hash column to be added")
		}
	})

	// Test adding sync columns to table that already has them
	t.Run("columns_already_exist", func(t *testing.T) {
		// Columns should already exist from previous test
		err = AddSyncColumnsToIssues(db)
		if err != nil {
			t.Errorf("AddSyncColumnsToIssues failed on existing columns: %v", err)
		}
	})

	// Test with invalid database
	t.Run("invalid_database", func(t *testing.T) {
		// Close the database connection
		db.Close()

		err = AddSyncColumnsToIssues(db)
		if err == nil {
			t.Error("Expected error with closed database")
		}
	})
}

// TestEnsureGitHubCredentials_AdditionalCases tests more credential scenarios
func TestEnsureGitHubCredentials_AdditionalCases(t *testing.T) {
	tests := []struct {
		name        string
		owner       string
		repo        string
		token       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "whitespace_only_owner",
			owner:       "   ",
			repo:        "testrepo",
			token:       "token",
			expectError: true,
			errorMsg:    "Invalid GitHub token", // GitHub API will return invalid token error
		},
		{
			name:        "whitespace_only_repo",
			owner:       "testowner",
			repo:        "   ",
			token:       "token",
			expectError: true,
			errorMsg:    "Invalid GitHub token", // GitHub API will return invalid token error
		},
		{
			name:        "whitespace_only_token",
			owner:       "testowner",
			repo:        "testrepo",
			token:       "   ",
			expectError: true,
			errorMsg:    "token",
		},
		{
			name:        "newlines_in_values",
			owner:       "test\nowner",
			repo:        "test\nrepo",
			token:       "test\ntoken",
			expectError: true,
			errorMsg:    "",
		},
		{
			name:        "very_long_values",
			owner:       strings.Repeat("a", 1000),
			repo:        strings.Repeat("b", 1000),
			token:       strings.Repeat("c", 1000),
			expectError: true, // Will fail due to invalid token
			errorMsg:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := EnsureGitHubCredentials(tt.owner, tt.repo, tt.token)

			if tt.expectError && err == nil {
				t.Errorf("Expected error for %s, but got none", tt.name)
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error for %s, but got: %v", tt.name, err)
			}
			if tt.errorMsg != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorMsg, err)
				}
			}
		})
	}
}

// TestValidateRepositoryAccess_EdgeCases tests repository access validation edge cases
func TestValidateRepositoryAccess_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		owner       string
		repo        string
		token       string
		expectError bool
	}{
		{
			name:        "special_characters_in_owner",
			owner:       "test-owner_123",
			repo:        "test-repo",
			token:       "dummy-token",
			expectError: true, // Will fail with invalid token but tests path
		},
		{
			name:        "special_characters_in_repo",
			owner:       "test-owner",
			repo:        "test-repo_123.git",
			token:       "dummy-token",
			expectError: true, // Will fail with invalid token but tests path
		},
		{
			name:        "case_sensitive_values",
			owner:       "TestOwner",
			repo:        "TestRepo",
			token:       "dummy-token",
			expectError: true, // Will fail with invalid token but tests path
		},
		{
			name:        "unicode_characters",
			owner:       "test-owner-ñ",
			repo:        "test-repo-ü",
			token:       "dummy-token",
			expectError: true, // Will fail with invalid token but tests path
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRepositoryAccess(tt.owner, tt.repo, tt.token)

			if tt.expectError && err == nil {
				t.Errorf("Expected error for %s, but got none", tt.name)
			}
			// We don't check for no error since these will all fail with dummy tokens
		})
	}
}

// TestFetchIssues_AdditionalCoverage tests more FetchIssues scenarios
func TestFetchIssues_AdditionalCoverage(t *testing.T) {
	tests := []struct {
		name        string
		owner       string
		repo        string
		token       string
		expectError bool
	}{
		{
			name:        "numeric_owner_repo",
			owner:       "123",
			repo:        "456",
			token:       "dummy-token",
			expectError: true,
		},
		{
			name:        "mixed_case_values",
			owner:       "TestOwner",
			repo:        "TestRepo",
			token:       "dummy-token",
			expectError: true,
		},
		{
			name:        "hyphenated_values",
			owner:       "test-owner",
			repo:        "test-repo",
			token:       "dummy-token",
			expectError: true,
		},
		{
			name:        "underscore_values",
			owner:       "test_owner",
			repo:        "test_repo",
			token:       "dummy-token",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := FetchIssues(tt.owner, tt.repo, tt.token)

			if tt.expectError && err == nil {
				t.Errorf("Expected error for %s, but got none", tt.name)
			}
			// Log the error type for debugging
			if err != nil {
				t.Logf("FetchIssues error for %s: %v", tt.name, err)
			}
		})
	}
}

// TestCreateIssue_AdditionalCoverage tests more CreateIssue scenarios
func TestCreateIssue_AdditionalCoverage(t *testing.T) {
	tests := []struct {
		name        string
		owner       string
		repo        string
		token       string
		request     CreateIssueRequest
		expectError bool
	}{
		{
			name:  "issue_with_all_fields",
			owner: "test-owner",
			repo:  "test-repo",
			token: "dummy-token",
			request: CreateIssueRequest{
				Title:     "Test Issue",
				Body:      "Test body with detailed description",
				Labels:    []string{"bug", "urgent", "backend"},
				Assignees: []string{"user1", "user2"},
			},
			expectError: true, // Will fail with dummy token
		},
		{
			name:  "issue_with_empty_body",
			owner: "test-owner",
			repo:  "test-repo",
			token: "dummy-token",
			request: CreateIssueRequest{
				Title:  "Test Issue No Body",
				Labels: []string{"feature"},
			},
			expectError: true, // Will fail with dummy token
		},
		{
			name:  "issue_with_special_characters",
			owner: "test-owner",
			repo:  "test-repo",
			token: "dummy-token",
			request: CreateIssueRequest{
				Title: "Test Issue with émojis 🚀 and spëcial chars",
				Body:  "Body with ñúmérös and symbols: @#$%^&*()",
			},
			expectError: true, // Will fail with dummy token
		},
		{
			name:  "issue_with_long_content",
			owner: "test-owner",
			repo:  "test-repo",
			token: "dummy-token",
			request: CreateIssueRequest{
				Title: strings.Repeat("Very long title ", 20),
				Body:  strings.Repeat("Very long body content with lots of text. ", 100),
			},
			expectError: true, // Will fail with dummy token
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CreateIssue(tt.owner, tt.repo, tt.token, tt.request)

			if tt.expectError && err == nil {
				t.Errorf("Expected error for %s, but got none", tt.name)
			}
			// Log the error for debugging
			if err != nil {
				t.Logf("CreateIssue error for %s: %v", tt.name, err)
			}
		})
	}
}
