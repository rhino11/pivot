package internal

import (
	"os"
	"strings"
	"testing"
)

// TestLoadMultiProjectConfig_BackupFileHandling tests the config.yaml fallback behavior
func TestLoadMultiProjectConfig_BackupFileHandling(t *testing.T) {
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

	t.Run("Load from config.yaml when config.yml missing", func(t *testing.T) {
		// Create config.yaml (not config.yml)
		configContent := `global:
  database: ~/.pivot/test.db
  token: test_token
projects:
  - owner: testowner
    repo: testrepo
    path: /test/path`

		err := os.WriteFile("config.yaml", []byte(configContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create config.yaml: %v", err)
		}

		config, err := LoadMultiProjectConfig()
		if err != nil {
			t.Fatalf("LoadMultiProjectConfig failed: %v", err)
		}

		if config.Global.Token != "test_token" {
			t.Errorf("Expected token 'test_token', got '%s'", config.Global.Token)
		}

		if len(config.Projects) != 1 {
			t.Errorf("Expected 1 project, got %d", len(config.Projects))
		}

		// Clean up for next test
		os.Remove("config.yaml")
	})

	t.Run("Both config files missing", func(t *testing.T) {
		// Ensure no config files exist
		os.Remove("config.yml")
		os.Remove("config.yaml")

		_, err := LoadMultiProjectConfig()
		if err == nil {
			t.Error("Expected error when both config files are missing")
		}
	})

	t.Run("Invalid legacy config - missing required fields", func(t *testing.T) {
		// Create config that will fail legacy validation
		invalidLegacyContent := `owner: ""
repo: ""
database: test.db`

		err := os.WriteFile("config.yml", []byte(invalidLegacyContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create invalid config: %v", err)
		}

		_, err = LoadMultiProjectConfig()
		if err == nil {
			t.Error("Expected error for invalid legacy config with missing owner/repo")
		}
		if err != nil && err.Error() != "invalid configuration: missing required fields" {
			t.Errorf("Expected specific error message, got: %v", err)
		}
	})

	t.Run("Invalid YAML syntax", func(t *testing.T) {
		// Create invalid YAML that will fail to parse as both multi-project and legacy
		invalidYAML := `owner: test
repo: test
invalid: yaml: syntax: [missing bracket
database: test.db`

		err := os.WriteFile("config.yml", []byte(invalidYAML), 0644)
		if err != nil {
			t.Fatalf("Failed to create invalid YAML: %v", err)
		}

		_, err = LoadMultiProjectConfig()
		if err == nil {
			t.Error("Expected error for invalid YAML syntax")
		}
		if err != nil && !strings.Contains(err.Error(), "failed to parse config as either multi-project or legacy format") {
			t.Errorf("Expected parse error, got: %v", err)
		}
	})
}

// TestLoadMultiProjectConfig_MultiProjectDetection tests the conditions for detecting multi-project format
func TestLoadMultiProjectConfig_MultiProjectDetection(t *testing.T) {
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

	t.Run("Detect via global config with database and token", func(t *testing.T) {
		// Test detection via global.database and global.token without projects array
		configContent := `global:
  database: test.db
  token: test_token`

		err := os.WriteFile("config.yml", []byte(configContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create config: %v", err)
		}

		config, err := LoadMultiProjectConfig()
		if err != nil {
			t.Fatalf("LoadMultiProjectConfig failed: %v", err)
		}

		if config.Global.Database != "test.db" {
			t.Errorf("Expected database 'test.db', got '%s'", config.Global.Database)
		}

		// Clean up
		os.Remove("config.yml")
	})

	t.Run("Detect via 'global:' keyword in content", func(t *testing.T) {
		// Test detection via string content check
		configContent := `# This config has global: keyword
global:
  database: ""  # Empty but keyword present
  token: ""     # Empty but keyword present
projects: []   # Empty projects array`

		err := os.WriteFile("config.yml", []byte(configContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create config: %v", err)
		}

		config, err := LoadMultiProjectConfig()
		if err != nil {
			t.Fatalf("LoadMultiProjectConfig failed: %v", err)
		}

		// Should be detected as multi-project format even with empty values
		if len(config.Projects) != 0 {
			t.Errorf("Expected 0 projects, got %d", len(config.Projects))
		}

		// Clean up
		os.Remove("config.yml")
	})

	t.Run("Detect via 'projects:' keyword in content", func(t *testing.T) {
		// Test detection via string content check for projects
		configContent := `# This config has projects: keyword
global:
  database: ""
  token: ""
projects:   # Keyword present
  - owner: testowner
    repo: testrepo`

		err := os.WriteFile("config.yml", []byte(configContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create config: %v", err)
		}

		config, err := LoadMultiProjectConfig()
		if err != nil {
			t.Fatalf("LoadMultiProjectConfig failed: %v", err)
		}

		if len(config.Projects) != 1 {
			t.Errorf("Expected 1 project, got %d", len(config.Projects))
		}

		// Clean up
		os.Remove("config.yml")
	})
}

// TestLoadMultiProjectConfig_LegacyConversion tests edge cases in legacy config conversion
func TestLoadMultiProjectConfig_LegacyConversion(t *testing.T) {
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

	t.Run("Valid legacy config conversion", func(t *testing.T) {
		// Test successful legacy config conversion
		legacyContent := `owner: legacyowner
repo: legacyrepo
token: legacy_token
database: legacy.db`

		err := os.WriteFile("config.yml", []byte(legacyContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create legacy config: %v", err)
		}

		config, err := LoadMultiProjectConfig()
		if err != nil {
			t.Fatalf("LoadMultiProjectConfig failed: %v", err)
		}

		// Check conversion to multi-project format
		if config.Global.Token != "legacy_token" {
			t.Errorf("Expected global token 'legacy_token', got '%s'", config.Global.Token)
		}

		if config.Global.Database != "legacy.db" {
			t.Errorf("Expected global database 'legacy.db', got '%s'", config.Global.Database)
		}

		if len(config.Projects) != 1 {
			t.Fatalf("Expected 1 project from legacy conversion, got %d", len(config.Projects))
		}

		project := config.Projects[0]
		if project.Owner != "legacyowner" || project.Repo != "legacyrepo" {
			t.Errorf("Legacy project not properly converted: owner=%s, repo=%s", project.Owner, project.Repo)
		}

		// Path should be set to current directory
		currentDir, _ := os.Getwd()
		if project.Path != currentDir {
			t.Errorf("Expected path '%s', got '%s'", currentDir, project.Path)
		}

		// Clean up
		os.Remove("config.yml")
	})

	t.Run("Legacy config with partial fields", func(t *testing.T) {
		// Test legacy config with only required fields
		legacyContent := `owner: minimalowner
repo: minimalrepo`

		err := os.WriteFile("config.yml", []byte(legacyContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create minimal legacy config: %v", err)
		}

		config, err := LoadMultiProjectConfig()
		if err != nil {
			t.Fatalf("LoadMultiProjectConfig failed: %v", err)
		}

		// Global should have default database
		if config.Global.Database != "~/.pivot/pivot.db" {
			t.Errorf("Expected default database '~/.pivot/pivot.db', got '%s'", config.Global.Database)
		}

		// Global token should be empty
		if config.Global.Token != "" {
			t.Errorf("Expected empty global token, got '%s'", config.Global.Token)
		}

		if len(config.Projects) != 1 {
			t.Fatalf("Expected 1 project, got %d", len(config.Projects))
		}

		project := config.Projects[0]
		if project.Owner != "minimalowner" || project.Repo != "minimalrepo" {
			t.Errorf("Minimal legacy project incorrect: owner=%s, repo=%s", project.Owner, project.Repo)
		}

		// Clean up
		os.Remove("config.yml")
	})
}
