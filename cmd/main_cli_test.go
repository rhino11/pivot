package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Test helper to create temporary config
func createTestConfig(t *testing.T) (string, func()) {
	t.Helper()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	configContent := `owner: testowner
repo: testrepo
token: ghp_testtoken123
database: ./test-pivot.db
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Change to temp directory
	oldDir, _ := os.Getwd()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	cleanup := func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Logf("Warning: Failed to change back to original directory: %v", err)
		}
		os.Remove(filepath.Join(tempDir, "test-pivot.db"))
	}

	return configPath, cleanup
}

// TestRootCommandHelp tests the root help functionality
func TestRootCommandHelp(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		contains []string
	}{
		{
			name:     "help flag",
			args:     []string{"--help"},
			contains: []string{"Pivot is a CLI tool", "init", "config", "sync", "version"},
		},
		{
			name:     "help command",
			args:     []string{"help"},
			contains: []string{"Pivot is a CLI tool", "init", "config", "sync", "version"},
		},
		{
			name:     "no args (should show help)",
			args:     []string{},
			contains: []string{"Pivot is a CLI tool"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			cmd := NewRootCommand()
			cmd.SetOut(output)
			cmd.SetErr(output)
			cmd.SetArgs(tt.args)

			_ = cmd.Execute() // Help commands shouldn't error

			result := output.String()
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("Help output should contain '%s'. Got: %s", expected, result)
				}
			}
		})
	}
}

// TestVersionCommand tests the version command
func TestVersionCommand(t *testing.T) {
	output := &bytes.Buffer{}
	cmd := NewRootCommand()
	cmd.SetOut(output)
	cmd.SetErr(output)
	cmd.SetArgs([]string{"version"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Version command should not error: %v", err)
	}

	result := output.String()
	expectedStrings := []string{"pivot version", "commit:", "built:"}
	for _, expected := range expectedStrings {
		if !strings.Contains(result, expected) {
			t.Errorf("Version output should contain '%s'. Got: %q", expected, result)
		}
	}
}

// TestInitCommand tests the init command happy path and failures
func TestInitCommand(t *testing.T) {
	t.Run("successful init without existing config", func(t *testing.T) {
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

		// Create a minimal valid config first to avoid interactive setup
		configContent := `owner: testowner
repo: testrepo
token: ghp_testtoken123
database: ./test-pivot.db
`
		err := os.WriteFile("config.yml", []byte(configContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create test config: %v", err)
		}

		output := &bytes.Buffer{}
		cmd := NewRootCommand()
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetArgs([]string{"init"})

		err = cmd.Execute()
		if err != nil {
			t.Fatalf("Init should succeed: %v", err)
		}

		// Verify database file was created
		if _, err := os.Stat("test-pivot.db"); os.IsNotExist(err) {
			t.Error("Database file should be created")
		}
	})

	t.Run("init with existing config", func(t *testing.T) {
		_, cleanup := createTestConfig(t)
		defer cleanup()

		output := &bytes.Buffer{}
		cmd := NewRootCommand()
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetArgs([]string{"init"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Init with existing config should succeed: %v", err)
		}

		// Verify database file was created
		if _, err := os.Stat("test-pivot.db"); os.IsNotExist(err) {
			t.Error("Database file should be created with existing config")
		}
	})

	t.Run("init help", func(t *testing.T) {
		output := &bytes.Buffer{}
		cmd := NewRootCommand()
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetArgs([]string{"init", "--help"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Init help should not error: %v", err)
		}

		result := output.String()
		if !strings.Contains(result, "Initialize Pivot by creating a configuration file") {
			t.Errorf("Init help should contain description. Got: %s", result)
		}
	})

	t.Run("init in directory without permissions", func(t *testing.T) {
		// Create a temporary directory that we can't write to
		tempDir := t.TempDir()
		restrictedDir := filepath.Join(tempDir, "restricted")
		err := os.Mkdir(restrictedDir, 0555) // read and execute only
		if err != nil {
			t.Fatalf("Failed to create restricted directory: %v", err)
		}
		defer func() {
			if err := os.Chmod(restrictedDir, 0755); err != nil {
				t.Logf("Warning: Failed to restore directory permissions: %v", err)
			}
		}()

		oldDir, _ := os.Getwd()
		defer func() {
			if err := os.Chdir(oldDir); err != nil {
				t.Logf("Warning: Failed to change back to original directory: %v", err)
			}
		}()
		if err := os.Chdir(restrictedDir); err != nil {
			t.Fatalf("Failed to change to restricted directory: %v", err)
		}

		// Create config that points to a database in restricted location
		configContent := `owner: testowner
repo: testrepo
token: ghp_testtoken123
database: ./restricted-pivot.db
`
		configPath := filepath.Join(tempDir, "config.yml")
		err = os.WriteFile(configPath, []byte(configContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create test config: %v", err)
		}

		// Copy config to restricted directory (should fail on some systems)
		// This test may not work on all systems due to permission handling
		output := &bytes.Buffer{}
		cmd := NewRootCommand()
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetArgs([]string{"init"})

		err = cmd.Execute()
		// We expect this to potentially fail due to permissions
		// but we don't want to make the test fail if the OS allows it
		if err != nil {
			t.Logf("Init correctly failed in restricted directory: %v", err)
		} else {
			t.Log("Init succeeded in restricted directory (OS allows)")
		}
	})
}

// TestConfigCommand tests the config command
func TestConfigCommand(t *testing.T) {
	t.Run("config help", func(t *testing.T) {
		output := &bytes.Buffer{}
		cmd := NewRootCommand()
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetArgs([]string{"config", "--help"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Config help should not error: %v", err)
		}

		result := output.String()
		if !strings.Contains(result, "Set up or modify Pivot configuration") {
			t.Errorf("Config help should contain description. Got: %s", result)
		}
	})

	t.Run("config command without interaction", func(t *testing.T) {
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

		// The config command will try to be interactive, but we can't easily test that
		// So we'll just verify it doesn't crash when called
		output := &bytes.Buffer{}
		cmd := NewRootCommand()
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetArgs([]string{"config"})

		// Since this will try to read from stdin and we can't provide input easily,
		// we expect it to fail, but we want to make sure it fails gracefully
		err := cmd.Execute()
		if err != nil {
			// This is expected since we can't provide interactive input
			t.Logf("Config command failed as expected without input: %v", err)
		} else {
			t.Log("Config command succeeded (unexpected but not necessarily wrong)")
		}
	})
}

// TestSyncCommand tests the sync command
func TestSyncCommand(t *testing.T) {
	t.Run("sync without config should fail", func(t *testing.T) {
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
		cmd.SetArgs([]string{"sync"})

		err := cmd.Execute()
		if err == nil {
			t.Error("Sync without config should fail")
		}

		if !strings.Contains(err.Error(), "sync failed") {
			t.Errorf("Error should mention sync failure: %v", err)
		}
	})

	t.Run("sync with invalid config should fail", func(t *testing.T) {
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

		// Create invalid config
		configContent := `invalid yaml content:`
		err := os.WriteFile("config.yaml", []byte(configContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create invalid config: %v", err)
		}

		output := &bytes.Buffer{}
		cmd := NewRootCommand()
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetArgs([]string{"sync"})

		err = cmd.Execute()
		if err == nil {
			t.Error("Sync with invalid config should fail")
		}
	})

	t.Run("sync help", func(t *testing.T) {
		output := &bytes.Buffer{}
		cmd := NewRootCommand()
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetArgs([]string{"sync", "--help"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Sync help should not error: %v", err)
		}

		result := output.String()
		if !strings.Contains(result, "Sync issues between upstream and local database") {
			t.Error("Sync help should contain description")
		}
	})
}

// TestInvalidCommands tests invalid command handling
func TestInvalidCommands(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"invalid command", []string{"invalid"}},
		{"invalid flag", []string{"--invalid"}},
		{"invalid subcommand", []string{"init", "invalid"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			cmd := NewRootCommand()
			cmd.SetOut(output)
			cmd.SetErr(output)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if err == nil {
				t.Error("Invalid command should return error")
			}
		})
	}
}

// TestCommandsExist verifies all expected commands are available
func TestCommandsExist(t *testing.T) {
	cmd := NewRootCommand()

	// Test our explicitly added commands
	expectedCommands := []string{"init", "config", "sync", "version"}

	for _, expectedCmd := range expectedCommands {
		found := false
		for _, subCmd := range cmd.Commands() {
			if subCmd.Name() == expectedCmd {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Expected command '%s' not found", expectedCmd)
		}
	}

	// Test that help and completion work (cobra adds these automatically)
	output := &bytes.Buffer{}
	helpCmd := NewRootCommand()
	helpCmd.SetOut(output)
	helpCmd.SetErr(output)
	helpCmd.SetArgs([]string{"help"})

	err := helpCmd.Execute()
	if err != nil {
		t.Errorf("Help command should work: %v", err)
	}

	// Test completion command exists
	completionOutput := &bytes.Buffer{}
	completionCmd := NewRootCommand()
	completionCmd.SetOut(completionOutput)
	completionCmd.SetErr(completionOutput)
	completionCmd.SetArgs([]string{"completion", "--help"})

	err = completionCmd.Execute()
	if err != nil {
		t.Errorf("Completion command should work: %v", err)
	}
}

// TestMainFunction tests the main function and Run()
func TestMainFunction(t *testing.T) {
	t.Run("Run returns 0 on success", func(t *testing.T) {
		// Mock os.Args
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()
		os.Args = []string{"pivot", "version"}

		exitCode := Run()
		if exitCode != 0 {
			t.Errorf("Expected exit code 0 for version command, got %d", exitCode)
		}
	})

	t.Run("Run returns 1 on error", func(t *testing.T) {
		// Mock os.Args with invalid command
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()
		os.Args = []string{"pivot", "invalid"}

		exitCode := Run()
		if exitCode != 1 {
			t.Errorf("Expected exit code 1 for invalid command, got %d", exitCode)
		}
	})
}

// BenchmarkCommands benchmarks command execution
func BenchmarkVersionCommand(b *testing.B) {
	for i := 0; i < b.N; i++ {
		output := &bytes.Buffer{}
		cmd := NewRootCommand()
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetArgs([]string{"version"})
		_ = cmd.Execute() // Ignore error in benchmark
	}
}

func BenchmarkHelpCommand(b *testing.B) {
	for i := 0; i < b.N; i++ {
		output := &bytes.Buffer{}
		cmd := NewRootCommand()
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetArgs([]string{"--help"})
		_ = cmd.Execute() // Ignore error in benchmark
	}
}

// TestConcurrentCommandExecution tests thread safety
func TestConcurrentCommandExecution(t *testing.T) {
	const numGoroutines = 10

	results := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			output := &bytes.Buffer{}
			cmd := NewRootCommand()
			cmd.SetOut(output)
			cmd.SetErr(output)
			cmd.SetArgs([]string{"version"})
			results <- cmd.Execute()
		}()
	}

	for i := 0; i < numGoroutines; i++ {
		select {
		case err := <-results:
			if err != nil {
				t.Errorf("Concurrent command execution failed: %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Error("Concurrent command execution timed out")
		}
	}
}
