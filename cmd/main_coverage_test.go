package main

import (
	"os"
	"testing"
)

// TestMain_ExitCodes tests the main function exit behavior
func TestMain_ExitCodes(t *testing.T) {
	// Save original args and defer restoration
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	tests := []struct {
		name     string
		args     []string
		expected int
	}{
		{
			name:     "successful_version_command",
			args:     []string{"pivot", "version"},
			expected: 0,
		},
		{
			name:     "help_command",
			args:     []string{"pivot", "--help"},
			expected: 0,
		},
		{
			name:     "invalid_command",
			args:     []string{"pivot", "nonexistent"},
			expected: 1,
		},
		{
			name:     "invalid_flag",
			args:     []string{"pivot", "--invalid-flag"},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set test args
			os.Args = tt.args

			// Test Run function
			result := Run()
			if result != tt.expected {
				t.Errorf("Expected exit code %d for %s, got %d", tt.expected, tt.name, result)
			}
		})
	}
}

// TestRun_EdgeCases tests additional edge cases for the Run function
func TestRun_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "empty_args",
			args:        []string{"pivot"},
			expectError: false, // Should show help
		},
		{
			name:        "multiple_help_flags",
			args:        []string{"pivot", "--help", "--help"},
			expectError: false,
		},
		{
			name:        "command_with_help",
			args:        []string{"pivot", "init", "--help"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore args
			originalArgs := os.Args
			defer func() { os.Args = originalArgs }()

			os.Args = tt.args
			result := Run()

			if tt.expectError && result == 0 {
				t.Errorf("Expected non-zero exit code for %s", tt.name)
			}
			if !tt.expectError && result != 0 {
				t.Errorf("Expected zero exit code for %s, got %d", tt.name, result)
			}
		})
	}
}

// TestNewRootCommand_Coverage tests additional command structure
func TestNewRootCommand_Coverage(t *testing.T) {
	rootCmd := NewRootCommand()

	// Test command metadata
	if rootCmd.Use != "pivot" {
		t.Errorf("Expected root command use 'pivot', got '%s'", rootCmd.Use)
	}

	if rootCmd.Short == "" {
		t.Error("Expected root command to have short description")
	}

	if rootCmd.Long == "" {
		t.Error("Expected root command to have long description")
	}

	// Test that all expected commands are present
	expectedCommands := []string{
		"init", "config", "sync", "status", "push", "resolve",
		"auth", "import", "export", "version",
	}

	commands := rootCmd.Commands()
	commandNames := make(map[string]bool)
	for _, cmd := range commands {
		commandNames[cmd.Name()] = true
	}

	for _, expected := range expectedCommands {
		if !commandNames[expected] {
			t.Errorf("Expected command '%s' not found", expected)
		}
	}
}

// TestCommandFlags_Coverage tests command flags exist
func TestCommandFlags_Coverage(t *testing.T) {
	rootCmd := NewRootCommand()

	// Test init command flags
	initCmd, _, err := rootCmd.Find([]string{"init"})
	if err != nil {
		t.Fatalf("Init command not found: %v", err)
	}

	// Check for expected flags
	if initCmd.Flags().Lookup("import") == nil {
		t.Error("Expected init command to have --import flag")
	}

	if initCmd.Flags().Lookup("multi-project") == nil {
		t.Error("Expected init command to have --multi-project flag")
	}

	// Test sync command flags
	syncCmd, _, err := rootCmd.Find([]string{"sync"})
	if err != nil {
		t.Fatalf("Sync command not found: %v", err)
	}

	if syncCmd.Flags().Lookup("project") == nil {
		t.Error("Expected sync command to have --project flag")
	}

	// Test status command flags
	statusCmd, _, err := rootCmd.Find([]string{"status"})
	if err != nil {
		t.Fatalf("Status command not found: %v", err)
	}

	if statusCmd.Flags().Lookup("verbose") == nil {
		t.Error("Expected status command to have --verbose flag")
	}

	// Test push command flags
	pushCmd, _, err := rootCmd.Find([]string{"push"})
	if err != nil {
		t.Fatalf("Push command not found: %v", err)
	}

	if pushCmd.Flags().Lookup("dry-run") == nil {
		t.Error("Expected push command to have --dry-run flag")
	}

	if pushCmd.Flags().Lookup("limit") == nil {
		t.Error("Expected push command to have --limit flag")
	}
}
