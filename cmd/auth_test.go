package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestAuthCommand(t *testing.T) {
	t.Run("auth help", func(t *testing.T) {
		output := &bytes.Buffer{}
		cmd := NewRootCommand()
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetArgs([]string{"auth", "--help"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Auth help should not error: %v", err)
		}

		result := output.String()
		if !strings.Contains(result, "Manage GitHub authentication tokens and verify access") {
			t.Error("Auth help should contain description")
		}
		if !strings.Contains(result, "verify") {
			t.Error("Auth help should show verify subcommand")
		}
	})

	t.Run("auth verify help", func(t *testing.T) {
		output := &bytes.Buffer{}
		cmd := NewRootCommand()
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetArgs([]string{"auth", "verify", "--help"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Auth verify help should not error: %v", err)
		}

		result := output.String()
		if !strings.Contains(result, "Verify that your GitHub token is valid") {
			t.Error("Auth verify help should contain description")
		}
		if !strings.Contains(result, "--owner") {
			t.Error("Auth verify help should show --owner flag")
		}
		if !strings.Contains(result, "--repo") {
			t.Error("Auth verify help should show --repo flag")
		}
		if !strings.Contains(result, "--token") {
			t.Error("Auth verify help should show --token flag")
		}
	})

	t.Run("auth verify with invalid token", func(t *testing.T) {
		output := &bytes.Buffer{}
		cmd := NewRootCommand()
		cmd.SetOut(output)
		cmd.SetErr(output)
		cmd.SetArgs([]string{"auth", "verify", "--token", "invalid_token"})

		err := cmd.Execute()
		if err == nil {
			t.Error("Auth verify with invalid token should fail")
		}

		result := output.String()
		if !strings.Contains(result, "🧪 Testing GitHub token validity") {
			t.Error("Auth verify should show validation message")
		}
		if !strings.Contains(result, "❌ Token validation failed") {
			t.Error("Auth verify should show token validation failure")
		}
	})
}
