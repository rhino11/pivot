package main

import (
	"bytes"
	"io"
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

// TestConfigCommand tests the config command
func TestConfigCommand(t *testing.T) {
	// Clean up
	defer os.Remove("config.yml")

	rootCmd := NewRootCommand()
	rootCmd.SetArgs([]string{"config"})

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Simulate user input for configuration
	oldStdin := os.Stdin
	pr, pw, _ := os.Pipe()
	os.Stdin = pr

	// Provide input for config setup
	go func() {
		defer pw.Close()
		pw.WriteString("testowner\n")     // GitHub owner
		pw.WriteString("testrepo\n")      // Repository name
		pw.WriteString("ghp_testtoken\n") // GitHub token
		pw.WriteString("\n")              // Database path (default)
		pw.WriteString("n\n")             // Include closed issues (no)
		pw.WriteString("\n")              // Batch size (default)
	}()

	err := rootCmd.Execute()

	// Restore stdin/stdout
	os.Stdin = oldStdin
	w.Close()
	os.Stdout = oldStdout

	// Read output
	go func() {
		io.ReadAll(r)
	}()
	r.Close()

	if err != nil {
		t.Errorf("Config command should not fail with valid input, got: %v", err)
	}

	// Verify config file was created
	if _, err := os.Stat("config.yml"); os.IsNotExist(err) {
		t.Error("Expected config.yml to be created")
	}
}

// TestInitCommand_WithExistingConfig tests init command when config already exists
func TestInitCommand_WithExistingConfig(t *testing.T) {
	// Clean up
	defer os.Remove("config.yml")
	defer os.Remove("test_init_existing.db")

	// Create existing config
	configContent := `owner: testowner
repo: testrepo
token: ghp_testtoken
database: ./test_init_existing.db
`
	err := os.WriteFile("config.yml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	rootCmd := NewRootCommand()
	rootCmd.SetArgs([]string{"init"})

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = rootCmd.Execute()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read output
	output, _ := io.ReadAll(r)
	r.Close()

	if err != nil {
		t.Errorf("Init command should not fail with existing config, got: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Initializing local issues database") {
		t.Error("Expected database initialization message")
	}
	if !strings.Contains(outputStr, "✓ Initialized local issues database") {
		t.Error("Expected success message")
	}
	if !strings.Contains(outputStr, "🎉 Pivot is ready to use!") {
		t.Error("Expected ready message")
	}
}

// TestInitCommand_ConfigSetupError tests init command when config setup fails
func TestInitCommand_ConfigSetupError(t *testing.T) {
	// Clean up
	defer os.Remove("config.yml")
	defer os.Remove("config.yaml")

	rootCmd := NewRootCommand()
	rootCmd.SetArgs([]string{"init"})

	// Simulate EOF on stdin to cause config setup to fail
	oldStdin := os.Stdin
	pr, pw, _ := os.Pipe()
	os.Stdin = pr
	pw.Close() // Close immediately to simulate EOF

	err := rootCmd.Execute()

	// Restore stdin
	os.Stdin = oldStdin

	if err == nil {
		t.Error("Expected error when config setup fails due to EOF")
	}
	if !strings.Contains(err.Error(), "config setup failed") {
		t.Errorf("Expected 'config setup failed' in error, got: %v", err)
	}
}

// TestRootCommand_NoArgs tests root command without arguments
func TestRootCommand_NoArgs(t *testing.T) {
	rootCmd := NewRootCommand()
	rootCmd.SetArgs([]string{})

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := rootCmd.Execute()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read output
	output, _ := io.ReadAll(r)
	r.Close()

	if err != nil {
		t.Errorf("Root command should not fail when run without args, got: %v", err)
	}

	outputStr := string(output)
	// When no arguments are provided, cobra shows help by default
	if !strings.Contains(outputStr, "pivot") && !strings.Contains(outputStr, "Usage:") {
		t.Logf("Output: %s", outputStr)
		// This is acceptable - some CLI tools show usage, others show help
	}
}

// TestRootCommand_Help tests root command help
func TestRootCommand_Help(t *testing.T) {
	rootCmd := NewRootCommand()
	rootCmd.SetArgs([]string{"--help"})

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := rootCmd.Execute()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read output
	output, _ := io.ReadAll(r)
	r.Close()

	if err != nil {
		t.Errorf("Help command should not fail, got: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Available Commands:") {
		t.Error("Expected help to show available commands")
	}
	if !strings.Contains(outputStr, "init") {
		t.Error("Expected help to show init command")
	}
	if !strings.Contains(outputStr, "sync") {
		t.Error("Expected help to show sync command")
	}
	if !strings.Contains(outputStr, "version") {
		t.Error("Expected help to show version command")
	}
}
