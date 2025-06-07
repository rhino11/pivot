package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestRootCommand(t *testing.T) {
	// Capture output
	output := &bytes.Buffer{}

	// Create a new root command for testing
	cmd := NewRootCommand()
	cmd.SetOut(output)
	cmd.SetErr(output)

	// Test help command
	cmd.SetArgs([]string{"--help"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error for help command, got: %v", err)
	}

	helpOutput := output.String()
	if !strings.Contains(helpOutput, "Pivot is a CLI tool for managing GitHub issues") {
		t.Errorf("Help output should contain app description. Got: %s", helpOutput)
	}
	if !strings.Contains(helpOutput, "sync") {
		t.Errorf("Help output should contain sync command. Got: %s", helpOutput)
	}
}

func TestSyncCommand(t *testing.T) {
	// Create a test config file
	configContent := `owner: testowner
repo: testrepo
token: testtoken123
`
	configFile := "config.yaml"
	defer os.Remove(configFile)

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Also clean up any test database
	defer os.Remove("pivot.db")

	// Capture output
	output := &bytes.Buffer{}

	// Create a new root command for testing
	cmd := NewRootCommand()
	cmd.SetOut(output)
	cmd.SetErr(output)

	// Test sync command (this will fail due to invalid token, but we can test the command setup)
	cmd.SetArgs([]string{"sync"})
	err = cmd.Execute()

	// We expect this to fail since we're using a fake token, but the command should be recognized
	if err == nil {
		t.Log("Sync command executed (may have failed due to fake token, which is expected)")
	} else {
		// Check if it's a network/auth error (expected) vs command not found (unexpected)
		errorStr := err.Error()
		if strings.Contains(errorStr, "unknown command") {
			t.Errorf("Sync command not recognized: %v", err)
		} else {
			t.Logf("Sync command recognized but failed as expected (fake token): %v", err)
		}
	}
}

func TestVersionCommand(t *testing.T) {
	// Test version command by checking it doesn't error and runs successfully
	cmd := NewRootCommand()

	// Test version command (not flag)
	cmd.SetArgs([]string{"version"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error for version command, got: %v", err)
	}

	// The version command prints directly to stdout via fmt.Printf
	// For a more comprehensive test, we could redirect stdout, but this ensures the command works
	t.Log("Version command executed successfully")
}

func TestInvalidCommand(t *testing.T) {
	// Capture output
	output := &bytes.Buffer{}

	// Create a new root command for testing
	cmd := NewRootCommand()
	cmd.SetOut(output)
	cmd.SetErr(output)

	// Test invalid command
	cmd.SetArgs([]string{"invalidcommand"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("Expected error for invalid command")
	}

	if !strings.Contains(err.Error(), "unknown command") {
		t.Errorf("Expected 'unknown command' error, got: %v", err)
	}
}

func TestInitCommand(t *testing.T) {
	// Clean up any existing database and config files
	defer os.Remove("pivot.db")
	defer os.Remove("config.yml")
	defer os.Remove("config.yaml")

	// Create a test config file first to avoid interactive prompts
	configContent := `owner: testowner
repo: testrepo
token: testtoken123
database: pivot.db
sync:
  include_closed: true
  batch_size: 100
`
	err := os.WriteFile("config.yml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Capture output
	output := &bytes.Buffer{}

	// Create a new root command for testing
	cmd := NewRootCommand()
	cmd.SetOut(output)
	cmd.SetErr(output)

	// Test init command
	cmd.SetArgs([]string{"init"})
	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error for init command, got: %v", err)
	}

	// Check if database file was created
	if _, err := os.Stat("pivot.db"); os.IsNotExist(err) {
		t.Error("Expected database file to be created")
	}

	t.Log("Init command executed successfully")
}

func TestRun_Success(t *testing.T) {
	// Test successful execution
	// Create a temporary config file to avoid errors
	configContent := `owner: testowner
repo: testrepo
token: testtoken123
`
	configFile := "config.yaml"
	defer os.Remove(configFile)

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Clean up any database
	defer os.Remove("pivot.db")
	// Capture stderr
	originalStderr := os.Stderr
	_, w, _ := os.Pipe()
	os.Stderr = w

	// Test with version command (should succeed)
	os.Args = []string{"pivot", "version"}

	exitCode := Run()

	// Restore stderr
	w.Close()
	os.Stderr = originalStderr

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestRun_InvalidCommand(t *testing.T) {
	// Test with invalid command
	// Capture stderr
	originalStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Save original args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Set invalid command
	os.Args = []string{"pivot", "invalidcommand"}

	exitCode := Run()

	// Read stderr output
	w.Close()
	stderr := make([]byte, 1024)
	n, _ := r.Read(stderr)
	stderrStr := string(stderr[:n])

	// Restore stderr
	os.Stderr = originalStderr

	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}

	if !strings.Contains(stderrStr, "Error:") {
		t.Errorf("Expected error message in stderr, got: %s", stderrStr)
	}
}
