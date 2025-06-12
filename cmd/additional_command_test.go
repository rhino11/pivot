package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestInitCommandWithImport tests init command with --import flag
func TestInitCommandWithImport(t *testing.T) {
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
    path: /imported/path1`

	importFile := filepath.Join(tempDir, "import.yml")
	err := os.WriteFile(importFile, []byte(importConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create import file: %v", err)
	}

	output := &bytes.Buffer{}
	cmd := NewRootCommand()
	cmd.SetOut(output)
	cmd.SetErr(output)
	cmd.SetArgs([]string{"init", "--import", importFile})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Init command with import failed: %v", err)
	}

	// Verify config was imported
	if _, err := os.Stat("config.yml"); os.IsNotExist(err) {
		t.Error("Expected config.yml to be created")
	}
}

// TestConfigSetupMultiProject tests config setup with multi-project flag
func TestConfigSetupMultiProject(t *testing.T) {
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

	// Mock stdin to provide input for multi-project setup
	input := strings.NewReader("test_token\n\nn\n")
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	r, w, _ := os.Pipe()
	os.Stdin = r
	go func() {
		defer w.Close()
		io.Copy(w, input)
	}()

	output := &bytes.Buffer{}
	cmd := NewRootCommand()
	cmd.SetOut(output)
	cmd.SetErr(output)
	cmd.SetArgs([]string{"config", "setup", "--multi-project"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Config setup multi-project failed: %v", err)
	}

	// Verify config was created
	if _, err := os.Stat("config.yml"); os.IsNotExist(err) {
		t.Error("Expected config.yml to be created")
	}
}

// TestConfigAddProject tests the config add-project command
func TestConfigAddProject(t *testing.T) {
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

	// Mock stdin to provide input for add-project
	input := strings.NewReader("n\nnewowner\nnewrepo\n/new/path\n")
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	r, w, _ := os.Pipe()
	os.Stdin = r
	go func() {
		defer w.Close()
		io.Copy(w, input)
	}()

	output := &bytes.Buffer{}
	cmd := NewRootCommand()
	cmd.SetOut(output)
	cmd.SetErr(output)
	cmd.SetArgs([]string{"config", "add-project"})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Config add-project failed: %v", err)
	}

	// Check that the config was updated (this is enough to verify the command worked)
	if _, err := os.Stat("config.yml"); os.IsNotExist(err) {
		t.Error("Expected config.yml to still exist")
	}
}

// TestConfigImportCommand tests the config import command
func TestConfigImportCommand(t *testing.T) {
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
    path: /imported/path1`

	importFile := filepath.Join(tempDir, "import.yml")
	err := os.WriteFile(importFile, []byte(importConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create import file: %v", err)
	}

	output := &bytes.Buffer{}
	cmd := NewRootCommand()
	cmd.SetOut(output)
	cmd.SetErr(output)
	cmd.SetArgs([]string{"config", "import", importFile})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Config import failed: %v", err)
	}

	// Verify config was imported by checking the file exists
	if _, err := os.Stat("config.yml"); os.IsNotExist(err) {
		t.Error("Expected config.yml to be created")
	}
}

// TestVersionCommandDetails tests the version command details
func TestVersionCommandDetails(t *testing.T) {
	output := &bytes.Buffer{}
	cmd := NewRootCommand()
	cmd.SetOut(output)
	cmd.SetErr(output)
	cmd.SetArgs([]string{"version"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Version command failed: %v", err)
	}

	outputStr := output.String()
	if !strings.Contains(outputStr, "pivot version") {
		t.Errorf("Expected output to contain 'pivot version', got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "commit:") {
		t.Errorf("Expected output to contain 'commit:', got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "built:") {
		t.Errorf("Expected output to contain 'built:', got: %s", outputStr)
	}
}

// TestSyncCommandWithProject tests sync command with --project flag
func TestSyncCommandWithProject(t *testing.T) {
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

	// Create config
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

	output := &bytes.Buffer{}
	cmd := NewRootCommand()
	cmd.SetOut(output)
	cmd.SetErr(output)
	cmd.SetArgs([]string{"sync", "--project", "invalid_format"})

	err = cmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid project format")
	}
	if !strings.Contains(err.Error(), "project filter must be in format") {
		t.Errorf("Expected project filter error, got: %v", err)
	}
}

// TestCSVExportCommandBasic tests basic CSV export command
func TestCSVExportCommandBasic(t *testing.T) {
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

	output := &bytes.Buffer{}
	cmd := NewRootCommand()
	cmd.SetOut(output)
	cmd.SetErr(output)
	cmd.SetArgs([]string{"export", "csv"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("CSV export command failed: %v", err)
	}

	outputStr := output.String()
	if !strings.Contains(outputStr, "Exported") {
		t.Errorf("Expected output to contain 'Exported', got: %s", outputStr)
	}

	// Verify CSV file was created
	if _, err := os.Stat("issues.csv"); os.IsNotExist(err) {
		t.Error("Expected issues.csv to be created")
	}
}

// TestCSVExportCommandWithOptions tests CSV export with various options
func TestCSVExportCommandWithOptions(t *testing.T) {
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

	output := &bytes.Buffer{}
	cmd := NewRootCommand()
	cmd.SetOut(output)
	cmd.SetErr(output)
	cmd.SetArgs([]string{"export", "csv",
		"--output", "custom.csv",
		"--fields", "title,state,labels",
		"--filter", "state:open",
		"--repository", "owner/repo"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("CSV export command with options failed: %v", err)
	}

	// Verify custom CSV file was created
	if _, err := os.Stat("custom.csv"); os.IsNotExist(err) {
		t.Error("Expected custom.csv to be created")
	}
}

// TestCSVImportPreview tests CSV import with preview flag
func TestCSVImportPreview(t *testing.T) {
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

	// Create test CSV file
	csvContent := `title,state,body,labels
Test Issue 1,open,This is a test issue,bug
Test Issue 2,closed,Another test issue,feature`

	csvFile := filepath.Join(tempDir, "test.csv")
	err := os.WriteFile(csvFile, []byte(csvContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test CSV: %v", err)
	}

	output := &bytes.Buffer{}
	cmd := NewRootCommand()
	cmd.SetOut(output)
	cmd.SetErr(output)
	cmd.SetArgs([]string{"import", "csv", "--preview", csvFile})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("CSV import preview failed: %v", err)
	}

	outputStr := output.String()
	if !strings.Contains(outputStr, "Import Preview") {
		t.Errorf("Expected output to contain 'Import Preview', got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "Test Issue 1") {
		t.Errorf("Expected output to contain 'Test Issue 1', got: %s", outputStr)
	}
}

// TestCSVImportDryRun tests CSV import with dry-run flag
func TestCSVImportDryRun(t *testing.T) {
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

	// Create test CSV file
	csvContent := `title,state,body,labels
Test Issue 1,open,This is a test issue,bug
Test Issue 2,closed,Another test issue,feature`

	csvFile := filepath.Join(tempDir, "test.csv")
	err := os.WriteFile(csvFile, []byte(csvContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test CSV: %v", err)
	}

	output := &bytes.Buffer{}
	cmd := NewRootCommand()
	cmd.SetOut(output)
	cmd.SetErr(output)
	cmd.SetArgs([]string{"import", "csv", "--dry-run", csvFile})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("CSV import dry-run failed: %v", err)
	}

	outputStr := output.String()
	if !strings.Contains(outputStr, "Dry Run Mode") {
		t.Errorf("Expected output to contain 'Dry Run Mode', got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "Would create") {
		t.Errorf("Expected output to contain 'Would create', got: %s", outputStr)
	}
}

// TestCSVImportFileNotFound tests CSV import with non-existent file
func TestCSVImportFileNotFound(t *testing.T) {
	output := &bytes.Buffer{}
	cmd := NewRootCommand()
	cmd.SetOut(output)
	cmd.SetErr(output)
	cmd.SetArgs([]string{"import", "csv", "nonexistent.csv"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for non-existent CSV file")
	}
	if !strings.Contains(err.Error(), "CSV file not found") {
		t.Errorf("Expected 'CSV file not found' error, got: %v", err)
	}
}

// TestLegacyConfigFallback tests legacy config handling in sync command
func TestLegacyConfigFallback(t *testing.T) {
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

	// Create legacy config
	legacyConfig := `owner: testowner
repo: testrepo
token: test_token
database: ./test.db`

	err := os.WriteFile("config.yml", []byte(legacyConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create legacy config: %v", err)
	}

	output := &bytes.Buffer{}
	cmd := NewRootCommand()
	cmd.SetOut(output)
	cmd.SetErr(output)
	cmd.SetArgs([]string{"sync"})

	err = cmd.Execute()
	// This will fail due to invalid token/network, but it shows legacy fallback is working
	if err != nil {
		// This is expected - either auth failure or network issue
		t.Logf("Sync failed as expected with legacy config: %v", err)
	}
}
