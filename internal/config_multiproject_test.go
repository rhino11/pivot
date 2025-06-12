package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestMultiProjectConfigStructure tests the new multi-project config structure
func TestMultiProjectConfigStructure(t *testing.T) {
	// Test that we can load a multi-project config
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

	// Create a multi-project config
	configContent := `# Multi-project configuration
global:
  database: ~/.pivot/pivot.db
  token: ghp_global_token123

projects:
  - owner: rhino11
    repo: pivot
    path: /Users/ryan/code/github.com/rhino11/pivot
  - owner: myorg
    repo: myproject
    path: /Users/ryan/code/github.com/myorg/myproject
    token: ghp_project_specific_token
`
	err := os.WriteFile("config.yml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Test loading multi-project config
	config, err := LoadMultiProjectConfig()
	if err != nil {
		t.Fatalf("LoadMultiProjectConfig failed: %v", err)
	}

	// Verify global settings
	if config.Global.Database != "~/.pivot/pivot.db" {
		t.Errorf("Expected database '~/.pivot/pivot.db', got '%s'", config.Global.Database)
	}
	if config.Global.Token != "ghp_global_token123" {
		t.Errorf("Expected token 'ghp_global_token123', got '%s'", config.Global.Token)
	}

	// Verify projects
	if len(config.Projects) != 2 {
		t.Fatalf("Expected 2 projects, got %d", len(config.Projects))
	}

	project1 := config.Projects[0]
	if project1.Owner != "rhino11" || project1.Repo != "pivot" {
		t.Errorf("Unexpected project 1: %+v", project1)
	}

	project2 := config.Projects[1]
	if project2.Owner != "myorg" || project2.Repo != "myproject" {
		t.Errorf("Unexpected project 2: %+v", project2)
	}
	if project2.Token != "ghp_project_specific_token" {
		t.Errorf("Expected project-specific token, got '%s'", project2.Token)
	}
}

// TestLegacyConfigBackwardCompatibility tests that old single-project configs still work
func TestLegacyConfigBackwardCompatibility(t *testing.T) {
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

	// Create a legacy single-project config
	configContent := `owner: testowner
repo: testrepo
token: ghp_testtoken
database: ./pivot.db
`
	err := os.WriteFile("config.yml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Test that we can still load it as a multi-project config
	config, err := LoadMultiProjectConfig()
	if err != nil {
		t.Fatalf("LoadMultiProjectConfig failed with legacy config: %v", err)
	}

	// Should convert to multi-project format with single project
	if len(config.Projects) != 1 {
		t.Fatalf("Expected 1 project from legacy config, got %d", len(config.Projects))
	}

	project := config.Projects[0]
	if project.Owner != "testowner" || project.Repo != "testrepo" {
		t.Errorf("Legacy config not properly converted: %+v", project)
	}
}

// TestDetectProjectFromGit tests automatic project detection from .git directory
func TestDetectProjectFromGit(t *testing.T) {
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Logf("Warning: Failed to change back to original directory: %v", err)
		}
	}()

	// Create a fake .git directory with config
	gitDir := filepath.Join(tempDir, ".git")
	err := os.Mkdir(gitDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	// Create a fake git config with remote origin
	gitConfig := `[core]
	repositoryformatversion = 0
[remote "origin"]
	url = https://github.com/rhino11/pivot.git
	fetch = +refs/heads/*:refs/remotes/origin/*
`
	err = os.WriteFile(filepath.Join(gitDir, "config"), []byte(gitConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create git config: %v", err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Test project detection
	project, err := DetectProjectFromGit()
	if err != nil {
		t.Fatalf("DetectProjectFromGit failed: %v", err)
	}

	if project.Owner != "rhino11" || project.Repo != "pivot" {
		t.Errorf("Expected owner 'rhino11' and repo 'pivot', got owner '%s' and repo '%s'",
			project.Owner, project.Repo)
	}

	// Check that paths match (accounting for macOS /private prefix)
	expectedPath := tempDir
	actualPath := project.Path
	// On macOS, resolve both paths to handle /private symlinks
	if actualPath != expectedPath {
		actualAbs, _ := filepath.EvalSymlinks(actualPath)
		expectedAbs, _ := filepath.EvalSymlinks(expectedPath)
		if actualAbs != expectedAbs {
			t.Errorf("Expected path '%s' (resolved: '%s'), got '%s' (resolved: '%s')",
				expectedPath, expectedAbs, actualPath, actualAbs)
		}
	}
}

// TestCentralDatabasePath tests that we can resolve central database paths
func TestCentralDatabasePath(t *testing.T) {
	tests := []struct {
		name     string
		dbPath   string
		expected string
	}{
		{
			name:     "home directory expansion",
			dbPath:   "~/.pivot/pivot.db",
			expected: "/.pivot/pivot.db", // Will be expanded to actual home
		},
		{
			name:     "absolute path",
			dbPath:   "/var/lib/pivot/pivot.db",
			expected: "/var/lib/pivot/pivot.db",
		},
		{
			name:     "relative path",
			dbPath:   "./pivot.db",
			expected: "./pivot.db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved, err := ResolveDatabasePath(tt.dbPath)
			if err != nil {
				t.Fatalf("ResolveDatabasePath failed: %v", err)
			}

			if tt.name == "home directory expansion" {
				// For home expansion, just check it contains the suffix
				if !strings.HasSuffix(resolved, tt.expected) {
					t.Errorf("Expected path to end with '%s', got '%s'", tt.expected, resolved)
				}
			} else {
				if resolved != tt.expected {
					t.Errorf("Expected '%s', got '%s'", tt.expected, resolved)
				}
			}
		})
	}
}

// TestConfigImportFromFile tests importing configuration from a file
func TestConfigImportFromFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create a config file to import
	importConfigContent := `global:
  database: ~/.pivot/pivot.db
  token: ghp_imported_token

projects:
  - owner: imported
    repo: project
    path: /path/to/imported/project
`
	importFile := filepath.Join(tempDir, "import.yml")
	err := os.WriteFile(importFile, []byte(importConfigContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create import file: %v", err)
	}

	// Test importing
	config, err := ImportConfigFromFile(importFile)
	if err != nil {
		t.Fatalf("ImportConfigFromFile failed: %v", err)
	}

	if config.Global.Token != "ghp_imported_token" {
		t.Errorf("Expected imported token, got '%s'", config.Global.Token)
	}

	if len(config.Projects) != 1 {
		t.Fatalf("Expected 1 imported project, got %d", len(config.Projects))
	}

	project := config.Projects[0]
	if project.Owner != "imported" || project.Repo != "project" {
		t.Errorf("Imported project incorrect: %+v", project)
	}
}

// TestFindGitDirectory tests the findGitDirectory function
func TestFindGitDirectory(t *testing.T) {
	// Create a temporary directory structure with .git
	tempDir := t.TempDir()
	gitDir := filepath.Join(tempDir, ".git")

	err := os.Mkdir(gitDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	// Change to the temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		_ = os.Chdir(originalDir) // #nosec G104 - test cleanup, ignore error
	}()

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Test finding .git directory
	foundGitDir, err := findGitDirectory()
	if err != nil {
		t.Fatalf("findGitDirectory failed: %v", err)
	}

	// Resolve symlinks for proper comparison on macOS
	expectedGitDir, _ := filepath.EvalSymlinks(gitDir)
	foundGitDir, _ = filepath.EvalSymlinks(foundGitDir)

	if foundGitDir != expectedGitDir {
		t.Errorf("Expected git directory '%s', got '%s'", expectedGitDir, foundGitDir)
	}

	// Test from subdirectory
	subDir := filepath.Join(tempDir, "subdir")
	err = os.Mkdir(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	err = os.Chdir(subDir)
	if err != nil {
		t.Fatalf("Failed to change to subdirectory: %v", err)
	}

	foundGitDir, err = findGitDirectory()
	if err != nil {
		t.Fatalf("findGitDirectory failed from subdirectory: %v", err)
	}

	// Resolve symlinks for proper comparison
	foundGitDir, _ = filepath.EvalSymlinks(foundGitDir)
	if foundGitDir != expectedGitDir {
		t.Errorf("Expected git directory '%s' from subdirectory, got '%s'", expectedGitDir, foundGitDir)
	}
}

// TestFindGitDirectory_NotFound tests findGitDirectory when no .git directory exists
func TestFindGitDirectory_NotFound(t *testing.T) {
	// Create a temporary directory without .git
	tempDir := t.TempDir()

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		_ = os.Chdir(originalDir) // #nosec G104 - test cleanup, ignore error
	}()

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	_, err = findGitDirectory()
	if err == nil {
		t.Error("Expected error when .git directory not found, got nil")
	}

	if !strings.Contains(err.Error(), ".git directory not found") {
		t.Errorf("Expected '.git directory not found' error, got: %v", err)
	}
}

// TestParseGitRemoteOrigin tests the parseGitRemoteOrigin function
func TestParseGitRemoteOrigin(t *testing.T) {
	testCases := []struct {
		name      string
		url       string
		owner     string
		repo      string
		shouldErr bool
	}{
		{
			name:  "HTTPS URL",
			url:   "https://github.com/owner/repo.git",
			owner: "owner",
			repo:  "repo",
		},
		{
			name:  "SSH URL",
			url:   "git@github.com:owner/repo.git",
			owner: "owner",
			repo:  "repo",
		},
		{
			name:  "HTTPS URL without .git",
			url:   "https://github.com/owner/repo",
			owner: "owner",
			repo:  "repo",
		},
		{
			name:  "SSH URL without .git",
			url:   "git@github.com:owner/repo",
			owner: "owner",
			repo:  "repo",
		},
		{
			name:  "Complex owner and repo names",
			url:   "https://github.com/complex-owner/complex-repo-name.git",
			owner: "complex-owner",
			repo:  "complex-repo-name",
		},
		{
			name:      "Invalid URL",
			url:       "not-a-git-url",
			shouldErr: true,
		},
		{
			name:      "Non-GitHub URL",
			url:       "https://gitlab.com/owner/repo.git",
			shouldErr: true,
		},
		{
			name:      "Empty URL",
			url:       "",
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock git config content
			var gitConfig string
			if tc.shouldErr && tc.url == "" {
				gitConfig = ""
			} else if tc.shouldErr {
				gitConfig = fmt.Sprintf(`[remote "origin"]
	url = %s`, tc.url)
			} else {
				gitConfig = fmt.Sprintf(`[core]
	repositoryformatversion = 0
[remote "origin"]
	url = %s
	fetch = +refs/heads/*:refs/remotes/origin/*`, tc.url)
			}

			owner, repo, err := parseGitRemoteOrigin(gitConfig)

			if tc.shouldErr {
				if err == nil {
					t.Errorf("Expected error for URL '%s', got nil", tc.url)
				}
				return
			}

			if err != nil {
				t.Fatalf("parseGitRemoteOrigin failed for URL '%s': %v", tc.url, err)
			}

			if owner != tc.owner {
				t.Errorf("Expected owner '%s', got '%s'", tc.owner, owner)
			}

			if repo != tc.repo {
				t.Errorf("Expected repo '%s', got '%s'", tc.repo, repo)
			}
		})
	}
}

// TestSetDefaults tests the setDefaults function
func TestSetDefaults(t *testing.T) {
	testCases := []struct {
		name     string
		input    MultiProjectConfig
		expected MultiProjectConfig
	}{
		{
			name:  "Empty config gets defaults",
			input: MultiProjectConfig{},
			expected: MultiProjectConfig{
				Global: GlobalConfig{
					Database: "~/.pivot/pivot.db",
				},
			},
		},
		{
			name: "Existing database preserved",
			input: MultiProjectConfig{
				Global: GlobalConfig{
					Database: "/custom/path/db.sqlite",
					Token:    "custom-token",
				},
			},
			expected: MultiProjectConfig{
				Global: GlobalConfig{
					Database: "/custom/path/db.sqlite",
					Token:    "custom-token",
				},
			},
		},
		{
			name: "Empty database gets default",
			input: MultiProjectConfig{
				Global: GlobalConfig{
					Database: "",
					Token:    "preserve-token",
				},
			},
			expected: MultiProjectConfig{
				Global: GlobalConfig{
					Database: "~/.pivot/pivot.db",
					Token:    "preserve-token",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.input
			setDefaults(&result)

			if result.Global.Database != tc.expected.Global.Database {
				t.Errorf("Expected database '%s', got '%s'",
					tc.expected.Global.Database, result.Global.Database)
			}

			if result.Global.Token != tc.expected.Global.Token {
				t.Errorf("Expected token '%s', got '%s'",
					tc.expected.Global.Token, result.Global.Token)
			}
		})
	}
}
