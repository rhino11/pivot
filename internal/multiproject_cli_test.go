package internal

import (
	"database/sql"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// TestSaveMultiProjectConfig tests saving configuration to file
func TestSaveMultiProjectConfig(t *testing.T) {
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Logf("Warning: Failed to change back to original directory: %v", err)
		}
	}()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	config := &MultiProjectConfig{
		Global: GlobalConfig{
			Database: "~/.pivot/test.db",
			Token:    "ghp_test_token",
		},
		Projects: []ProjectConfig{
			{
				Owner: "testowner",
				Repo:  "testrepo",
				Path:  "/test/path",
				Token: "project_token",
			},
		},
	}

	err := SaveMultiProjectConfig(config)
	if err != nil {
		t.Fatalf("SaveMultiProjectConfig failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat("config.yml"); os.IsNotExist(err) {
		t.Error("config.yml was not created")
	}

	// Verify file contents by loading it back
	loadedConfig, err := LoadMultiProjectConfig()
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if loadedConfig.Global.Database != config.Global.Database {
		t.Errorf("Expected database %s, got %s", config.Global.Database, loadedConfig.Global.Database)
	}

	if loadedConfig.Global.Token != config.Global.Token {
		t.Errorf("Expected token %s, got %s", config.Global.Token, loadedConfig.Global.Token)
	}

	if len(loadedConfig.Projects) != len(config.Projects) {
		t.Errorf("Expected %d projects, got %d", len(config.Projects), len(loadedConfig.Projects))
	}
}

// TestGetEffectiveToken tests the GetEffectiveToken method
func TestGetEffectiveToken(t *testing.T) {
	global := &GlobalConfig{Token: "global_token"}

	testCases := []struct {
		name          string
		projectToken  string
		expectedToken string
	}{
		{
			name:          "Project has specific token",
			projectToken:  "project_token",
			expectedToken: "project_token",
		},
		{
			name:          "Project has no token, use global",
			projectToken:  "",
			expectedToken: "global_token",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			project := &ProjectConfig{Token: tc.projectToken}
			result := project.GetEffectiveToken(global)
			if result != tc.expectedToken {
				t.Errorf("Expected %s, got %s", tc.expectedToken, result)
			}
		})
	}
}

// TestGetEffectiveDatabase tests the GetEffectiveDatabase method
func TestGetEffectiveDatabase(t *testing.T) {
	global := &GlobalConfig{Database: "global.db"}

	testCases := []struct {
		name             string
		projectDatabase  string
		expectedDatabase string
	}{
		{
			name:             "Project has specific database",
			projectDatabase:  "project.db",
			expectedDatabase: "project.db",
		},
		{
			name:             "Project has no database, use global",
			projectDatabase:  "",
			expectedDatabase: "global.db",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			project := &ProjectConfig{Database: tc.projectDatabase}
			result := project.GetEffectiveDatabase(global)
			if result != tc.expectedDatabase {
				t.Errorf("Expected %s, got %s", tc.expectedDatabase, result)
			}
		})
	}
}

// TestInitMultiProjectConfig tests the InitMultiProjectConfig function with simulated input
func TestInitMultiProjectConfig(t *testing.T) {
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Logf("Warning: Failed to change back to original directory: %v", err)
		}
	}()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Test case: config doesn't exist, simulate minimal input
	t.Run("New config with minimal input", func(t *testing.T) {
		// Create a mock stdin for testing
		input := strings.NewReader("test_token\n\nn\n")
		oldStdin := os.Stdin
		defer func() { os.Stdin = oldStdin }()

		// Create a pipe to capture stdin
		r, w, _ := os.Pipe()
		os.Stdin = r
		go func() {
			defer w.Close()
			_, _ = io.Copy(w, input) // #nosec G104 - test helper, ignore error
		}()

		err := InitMultiProjectConfig()
		if err != nil {
			t.Fatalf("InitMultiProjectConfig failed: %v", err)
		}

		// Verify config file was created
		if _, err := os.Stat("config.yml"); os.IsNotExist(err) {
			t.Error("config.yml was not created")
		}
	})

	// Test case: config already exists, user cancels
	t.Run("Existing config, user cancels", func(t *testing.T) {
		// config.yml already exists from previous test
		input := strings.NewReader("n\n")
		oldStdin := os.Stdin
		defer func() { os.Stdin = oldStdin }()

		r, w, _ := os.Pipe()
		os.Stdin = r
		go func() {
			defer w.Close()
			_, _ = io.Copy(w, input) // #nosec G104 - test helper, ignore error
		}()

		err := InitMultiProjectConfig()
		if err != nil {
			t.Fatalf("InitMultiProjectConfig should not fail when user cancels: %v", err)
		}
	})

	// Test case: config exists, user overwrites
	t.Run("Existing config, user overwrites", func(t *testing.T) {
		input := strings.NewReader("y\ntest_token2\n\nn\n")
		oldStdin := os.Stdin
		defer func() { os.Stdin = oldStdin }()

		r, w, _ := os.Pipe()
		os.Stdin = r
		go func() {
			defer w.Close()
			_, _ = io.Copy(w, input) // #nosec G104 - test helper, ignore error
		}()

		err := InitMultiProjectConfig()
		if err != nil {
			t.Fatalf("InitMultiProjectConfig failed: %v", err)
		}
	})
}

// TestInitMultiProjectDatabase tests the InitMultiProjectDatabase function
func TestInitMultiProjectDatabase(t *testing.T) {
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Logf("Warning: Failed to change back to original directory: %v", err)
		}
	}()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create a test config file
	configContent := `global:
  database: test.db
  token: test_token
projects:
  - owner: testowner
    repo: testrepo
    path: /test/path`

	err := os.WriteFile("config.yml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	err = InitMultiProjectDatabase()
	if err != nil {
		t.Fatalf("InitMultiProjectDatabase failed: %v", err)
	}

	// Verify database was created
	if _, err := os.Stat("test.db"); os.IsNotExist(err) {
		t.Error("Database file was not created")
	}

	// Verify database has proper schema
	db, err := sql.Open("sqlite3", "test.db")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Check for projects table
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='projects'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check for projects table: %v", err)
	}
	if count != 1 {
		t.Error("Projects table was not created")
	}

	// Check that project was registered
	err = db.QueryRow("SELECT COUNT(*) FROM projects WHERE owner='testowner' AND repo='testrepo'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check for project registration: %v", err)
	}
	if count != 1 {
		t.Error("Project was not registered in database")
	}
}

// TestSyncMultiProject tests the SyncMultiProject function
func TestSyncMultiProject(t *testing.T) {
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Logf("Warning: Failed to change back to original directory: %v", err)
		}
	}()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create a test config file
	configContent := `global:
  database: test.db
  token: test_token
projects:
  - owner: testowner
    repo: testrepo
    path: /test/path`

	err := os.WriteFile("config.yml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Initialize database first
	err = InitMultiProjectDatabase()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Test with no projects configured
	emptyConfigContent := `global:
  database: test.db
  token: test_token
projects: []`

	err = os.WriteFile("config.yml", []byte(emptyConfigContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create empty config: %v", err)
	}

	err = SyncMultiProject("")
	if err == nil {
		t.Error("Expected error when no projects configured")
	}
	if !strings.Contains(err.Error(), "no projects configured") {
		t.Errorf("Expected 'no projects configured' error, got: %v", err)
	}

	// Restore original config
	err = os.WriteFile("config.yml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to restore config: %v", err)
	}

	// Test with invalid project filter
	err = SyncMultiProject("invalid")
	if err == nil {
		t.Error("Expected error for invalid project filter")
	}
	if !strings.Contains(err.Error(), "project filter must be in format 'owner/repo'") {
		t.Errorf("Expected project filter error, got: %v", err)
	}

	// Test with project filter that doesn't exist
	err = SyncMultiProject("nonexistent/repo")
	if err == nil {
		t.Error("Expected error for non-existent project")
	}
	if !strings.Contains(err.Error(), "project nonexistent/repo not found") {
		t.Errorf("Expected project not found error, got: %v", err)
	}

	// Note: We can't easily test successful sync without a real GitHub API or extensive mocking
	// The function attempts to call GitHub API which would fail in tests without network mocking
}

// TestShowMultiProjectConfig tests the ShowMultiProjectConfig function
func TestShowMultiProjectConfig(t *testing.T) {
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Logf("Warning: Failed to change back to original directory: %v", err)
		}
	}()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create a test config file
	configContent := `global:
  database: ~/.pivot/test.db
  token: ghp_test12345678
projects:
  - owner: testowner
    repo: testrepo
    path: /test/path
    token: project_token12345
  - owner: anotherowner
    repo: anotherrepo
    path: /another/path`

	err := os.WriteFile("config.yml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Capture stdout to verify output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = ShowMultiProjectConfig()

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("ShowMultiProjectConfig failed: %v", err)
	}

	// Read captured output
	output := make([]byte, 1024)
	n, _ := r.Read(output)
	outputStr := string(output[:n])

	// Verify expected content in output
	expectedStrings := []string{
		"Multi-Project Configuration",
		"Global Settings",
		"Database: ~/.pivot/test.db",
		"Token: ghp_test***",
		"Projects (2 configured)",
		"testowner/testrepo",
		"anotherowner/anotherrepo",
		"project_*** (project-specific)", // Adjusted to match actual output
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(outputStr, expected) {
			t.Errorf("Expected output to contain '%s', but it didn't. Output: %s", expected, outputStr)
		}
	}
}

// TestAddProject tests the AddProject function
func TestAddProject(t *testing.T) {
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Logf("Warning: Failed to change back to original directory: %v", err)
		}
	}()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create initial config
	configContent := `global:
  database: test.db
  token: global_token
projects:
  - owner: existing
    repo: repo
    path: /existing/path`

	err := os.WriteFile("config.yml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	t.Run("Add new project manually", func(t *testing.T) {
		// Simulate user input: manual entry, new project
		input := strings.NewReader("n\nnewowner\nnewrepo\n/new/path\n")
		oldStdin := os.Stdin
		defer func() { os.Stdin = oldStdin }()

		r, w, _ := os.Pipe()
		os.Stdin = r
		go func() {
			defer w.Close()
			_, _ = io.Copy(w, input) // #nosec G104 - test helper, ignore error
		}()

		err := AddProject()
		if err != nil {
			t.Fatalf("AddProject failed: %v", err)
		}

		// Verify project was added
		config, err := LoadMultiProjectConfig()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		if len(config.Projects) != 2 {
			t.Errorf("Expected 2 projects, got %d", len(config.Projects))
		}

		// Find the new project
		found := false
		for _, project := range config.Projects {
			if project.Owner == "newowner" && project.Repo == "newrepo" {
				found = true
				if project.Path != "/new/path" {
					t.Errorf("Expected path '/new/path', got '%s'", project.Path)
				}
				break
			}
		}

		if !found {
			t.Error("New project was not added to configuration")
		}
	})

	t.Run("Try to add duplicate project", func(t *testing.T) {
		// Try to add the same project again
		input := strings.NewReader("n\nexisting\nrepo\n/some/path\n")
		oldStdin := os.Stdin
		defer func() { os.Stdin = oldStdin }()

		r, w, _ := os.Pipe()
		os.Stdin = r
		go func() {
			defer w.Close()
			_, _ = io.Copy(w, input) // #nosec G104 - test helper, ignore error
		}()

		err := AddProject()
		if err == nil {
			t.Error("Expected error when adding duplicate project")
		}

		if !strings.Contains(err.Error(), "already exists in configuration") {
			t.Errorf("Expected duplicate project error, got: %v", err)
		}
	})
}

// TestImportConfigFile tests the ImportConfigFile function
func TestImportConfigFile(t *testing.T) {
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Logf("Warning: Failed to change back to original directory: %v", err)
		}
	}()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create import file
	importConfig := `global:
  database: imported.db
  token: imported_token
projects:
  - owner: imported
    repo: repo1
    path: /imported/path1
  - owner: imported
    repo: repo2
    path: /imported/path2`

	importFile := filepath.Join(tempDir, "import.yml")
	err := os.WriteFile(importFile, []byte(importConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create import file: %v", err)
	}

	t.Run("Import into empty configuration", func(t *testing.T) {
		err := ImportConfigFile(importFile)
		if err != nil {
			t.Fatalf("ImportConfigFile failed: %v", err)
		}

		// Verify configuration was imported
		config, err := LoadMultiProjectConfig()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		if config.Global.Database != "imported.db" {
			t.Errorf("Expected database 'imported.db', got '%s'", config.Global.Database)
		}

		if config.Global.Token != "imported_token" {
			t.Errorf("Expected token 'imported_token', got '%s'", config.Global.Token)
		}

		if len(config.Projects) != 2 {
			t.Errorf("Expected 2 projects, got %d", len(config.Projects))
		}
	})

	t.Run("Import into existing configuration (merge)", func(t *testing.T) {
		// Create existing config
		existingConfig := `global:
  database: existing.db
  token: existing_token
projects:
  - owner: existing
    repo: repo
    path: /existing/path`

		err := os.WriteFile("config.yml", []byte(existingConfig), 0644)
		if err != nil {
			t.Fatalf("Failed to create existing config: %v", err)
		}

		// Mock stdin to answer "y" for merge question
		input := strings.NewReader("y\n")
		oldStdin := os.Stdin
		defer func() { os.Stdin = oldStdin }()

		r, w, _ := os.Pipe()
		os.Stdin = r
		go func() {
			defer w.Close()
			_, _ = io.Copy(w, input) // #nosec G104 - test helper, ignore error
		}()

		err = ImportConfigFile(importFile)
		if err != nil {
			t.Fatalf("ImportConfigFile failed: %v", err)
		}

		// Verify configuration was merged
		config, err := LoadMultiProjectConfig()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		// Should prefer imported values for global config
		if config.Global.Database != "imported.db" {
			t.Errorf("Expected database 'imported.db', got '%s'", config.Global.Database)
		}

		if config.Global.Token != "imported_token" {
			t.Errorf("Expected token 'imported_token', got '%s'", config.Global.Token)
		}

		// Should have all projects (existing + imported)
		if len(config.Projects) != 3 {
			t.Errorf("Expected 3 projects after merge, got %d", len(config.Projects))
		}
	})

	t.Run("Import non-existent file", func(t *testing.T) {
		err := ImportConfigFile("/non/existent/file.yml")
		if err == nil {
			t.Error("Expected error when importing non-existent file")
		}
	})

	t.Run("Import invalid YAML file", func(t *testing.T) {
		invalidFile := filepath.Join(tempDir, "invalid.yml")
		err := os.WriteFile(invalidFile, []byte("invalid: yaml: content: ["), 0644)
		if err != nil {
			t.Fatalf("Failed to create invalid file: %v", err)
		}

		err = ImportConfigFile(invalidFile)
		if err == nil {
			t.Error("Expected error when importing invalid YAML")
		}
	})
}

// TestSaveMultiProjectConfigErrorCases tests error handling in SaveMultiProjectConfig
func TestSaveMultiProjectConfigErrorCases(t *testing.T) {
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

	// Test saving to read-only directory
	readonlyDir := "readonly"
	err := os.Mkdir(readonlyDir, 0444) // Read-only directory
	if err != nil {
		t.Fatalf("Failed to create readonly directory: %v", err)
	}

	err = os.Chdir(readonlyDir)
	if err != nil {
		t.Skipf("Cannot test readonly directory on this system: %v", err)
	}

	config := &MultiProjectConfig{
		Global: GlobalConfig{
			Database: "test.db",
			Token:    "test_token",
		},
		Projects: []ProjectConfig{
			{Owner: "test", Repo: "repo"},
		},
	}

	// This should fail because we can't write to read-only directory
	err = SaveMultiProjectConfig(config)
	if err == nil {
		t.Error("Expected error when saving to read-only directory")
	}
}

// TestInitMultiProjectDatabaseErrorCases tests error handling in InitMultiProjectDatabase
func TestInitMultiProjectDatabaseErrorCases(t *testing.T) {
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

	// Test with invalid database path in config
	configContent := `global:
  database: /invalid/readonly/path/test.db
  token: test_token
projects:
  - owner: testowner
    repo: testrepo`

	err := os.WriteFile("config.yml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	err = InitMultiProjectDatabase()
	if err == nil {
		t.Error("Expected error with invalid database path")
	}
}

// TestAddProjectEdgeCases tests edge cases for AddProject function
func TestAddProjectEdgeCases(t *testing.T) {
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

	// Test with invalid config file
	err := os.WriteFile("config.yml", []byte("invalid: yaml: ["), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid config: %v", err)
	}

	err = AddProject()
	if err == nil {
		t.Error("Expected error with invalid config file")
	}
	if !strings.Contains(err.Error(), "failed to load config") {
		t.Errorf("Expected config load error, got: %v", err)
	}
}
