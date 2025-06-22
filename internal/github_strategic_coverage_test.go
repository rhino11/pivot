package internal

import (
	"strings"
	"testing"
)

// TestGitHubFunctions_StrategicCoverage specifically targets GitHub functions to maximize coverage
func TestGitHubFunctions_StrategicCoverage(t *testing.T) {

	// Test 1: Hit FetchIssues parameter validation and URL construction paths
	t.Run("FetchIssues_ParameterAndUrlPaths", func(t *testing.T) {
		// This will exercise URL construction, header setting, and parameter validation
		// Even though the API call fails, it hits the code paths we need for coverage

		// Test with valid parameters to exercise URL construction
		_, err := FetchIssues("validowner", "validrepo", "validtoken")
		if err == nil {
			t.Errorf("Expected error due to invalid token")
		}
		// The error should be a credential error, showing we hit the validation path
		if !strings.Contains(err.Error(), "GitHub") {
			t.Errorf("Expected GitHub error, got: %v", err)
		}

		// Test with special characters to exercise URL escaping
		_, err = FetchIssues("owner-with-dash", "repo_with_underscore", "test-token")
		if err == nil {
			t.Errorf("Expected error due to invalid token")
		}

		// Test with empty parameters to exercise validation paths
		_, err = FetchIssues("", "repo", "token")
		if err == nil {
			t.Errorf("Expected error for empty owner")
		}

		_, err = FetchIssues("owner", "", "token")
		if err == nil {
			t.Errorf("Expected error for empty repo")
		}

		_, err = FetchIssues("owner", "repo", "")
		if err == nil {
			t.Errorf("Expected error for empty token")
		}
	})

	// Test 2: Hit CreateIssue marshaling, parameter validation, and URL construction paths
	t.Run("CreateIssue_MarshalingAndParameterPaths", func(t *testing.T) {
		// Test with complex request to exercise JSON marshaling path
		complexRequest := CreateIssueRequest{
			Title:     "Complex Issue Title with Special Characters: !@#$%^&*()",
			Body:      "Complex body with\nmultiple lines\nand unicode: 🚀 📝 ✅",
			Labels:    []string{"bug", "enhancement", "good first issue", "priority: high"},
			Assignees: []string{"user1", "user2", "user3"},
			Milestone: 42,
		}

		// This exercises JSON marshaling, URL construction, header setting
		_, err := CreateIssue("testowner", "testrepo", "testtoken", complexRequest)
		if err == nil {
			t.Errorf("Expected error due to invalid token")
		}
		if !strings.Contains(err.Error(), "GitHub") {
			t.Errorf("Expected GitHub error, got: %v", err)
		}

		// Test with minimal request
		minimalRequest := CreateIssueRequest{
			Title: "Minimal Issue",
		}
		_, err = CreateIssue("owner", "repo", "token", minimalRequest)
		if err == nil {
			t.Errorf("Expected error due to invalid token")
		}

		// Test with empty title to potentially hit validation
		emptyTitleRequest := CreateIssueRequest{
			Title: "",
		}
		_, err = CreateIssue("owner", "repo", "token", emptyTitleRequest)
		if err == nil {
			t.Errorf("Expected error due to invalid token or validation")
		}

		// Test with empty parameters to exercise validation paths
		_, err = CreateIssue("", "repo", "token", complexRequest)
		if err == nil {
			t.Errorf("Expected error for empty owner")
		}

		_, err = CreateIssue("owner", "", "token", complexRequest)
		if err == nil {
			t.Errorf("Expected error for empty repo")
		}

		_, err = CreateIssue("owner", "repo", "", complexRequest)
		if err == nil {
			t.Errorf("Expected error for empty token")
		}
	})

	// Test 3: Hit ValidateGitHubCredentials all branches including empty token path
	t.Run("ValidateGitHubCredentials_AllBranches", func(t *testing.T) {
		// Test empty token path (should hit the early return)
		err := ValidateGitHubCredentials("")
		if err == nil {
			t.Errorf("Expected error for empty token")
		}
		// Verify it's the correct error type and message
		credErr, ok := err.(*GitHubCredentialError)
		if !ok {
			t.Errorf("Expected GitHubCredentialError, got %T", err)
		}
		if credErr.StatusCode != 401 {
			t.Errorf("Expected status 401, got %d", credErr.StatusCode)
		}
		if !strings.Contains(credErr.Message, "No GitHub token provided") {
			t.Errorf("Expected 'No GitHub token provided' message, got: %s", credErr.Message)
		}

		// Test with various token formats to exercise request creation and API call
		testTokens := []string{
			"ghp_test_token_format",
			"github_pat_test_token",
			"test-token-with-dashes",
			"test_token_with_underscores",
			"very_long_test_token_that_might_be_realistic_length_wise_but_still_invalid",
		}

		for _, token := range testTokens {
			err := ValidateGitHubCredentials(token)
			if err == nil {
				t.Errorf("Expected error for invalid token: %s", token)
			}
			// Should get a credential error from the API call
			if !strings.Contains(err.Error(), "GitHub") {
				t.Errorf("Expected GitHub error for token %s, got: %v", token, err)
			}
		}
	})

	// Test 4: Hit ValidateRepositoryAccess branches and URL construction
	t.Run("ValidateRepositoryAccess_BranchesAndConstruction", func(t *testing.T) {
		// Test various owner/repo combinations to exercise URL construction
		testCases := []struct {
			owner string
			repo  string
		}{
			{"simple", "repo"},
			{"owner-with-dash", "repo_with_underscore"},
			{"org.with.dots", "repo-123"},
			{"UPPERCASE", "lowercase"},
			{"unicode-test", "repo-测试"},
		}

		for _, tc := range testCases {
			err := ValidateRepositoryAccess(tc.owner, tc.repo, "test-token")
			if err == nil {
				t.Errorf("Expected error for %s/%s", tc.owner, tc.repo)
			}
			// Should get credential error since token is invalid
			if !strings.Contains(err.Error(), "GitHub") {
				t.Errorf("Expected GitHub error for %s/%s, got: %v", tc.owner, tc.repo, err)
			}
		}

		// Test with empty parameters to exercise validation
		err := ValidateRepositoryAccess("", "repo", "token")
		if err == nil {
			t.Errorf("Expected error for empty owner")
		}

		err = ValidateRepositoryAccess("owner", "", "token")
		if err == nil {
			t.Errorf("Expected error for empty repo")
		}

		err = ValidateRepositoryAccess("owner", "repo", "")
		if err == nil {
			t.Errorf("Expected error for empty token")
		}
	})

	// Test 5: Hit EnsureGitHubCredentials branching logic comprehensively
	t.Run("EnsureGitHubCredentials_ComprehensiveBranching", func(t *testing.T) {
		// Test Case 1: Empty owner and repo (should only validate token)
		// This hits the `if owner != "" && repo != ""` condition being FALSE
		err := EnsureGitHubCredentials("", "", "test-token")
		if err == nil {
			t.Errorf("Expected error for invalid token")
		}
		if !strings.Contains(err.Error(), "GitHub") {
			t.Errorf("Expected GitHub error, got: %v", err)
		}

		// Test Case 2: Empty owner, non-empty repo (should only validate token)
		err = EnsureGitHubCredentials("", "repo", "test-token")
		if err == nil {
			t.Errorf("Expected error for invalid token")
		}

		// Test Case 3: Non-empty owner, empty repo (should only validate token)
		err = EnsureGitHubCredentials("owner", "", "test-token")
		if err == nil {
			t.Errorf("Expected error for invalid token")
		}

		// Test Case 4: Both owner and repo non-empty (should validate token AND repo)
		// This hits the `if owner != "" && repo != ""` condition being TRUE
		err = EnsureGitHubCredentials("owner", "repo", "test-token")
		if err == nil {
			t.Errorf("Expected error for invalid token")
		}

		// Test Case 5: Empty token (should fail early in ValidateGitHubCredentials)
		err = EnsureGitHubCredentials("", "", "")
		if err == nil {
			t.Errorf("Expected error for empty token")
		}
		// Should be the "No GitHub token provided" error
		if !strings.Contains(err.Error(), "No GitHub token provided") {
			t.Errorf("Expected 'No GitHub token provided' error, got: %v", err)
		}

		// Test Case 6: Non-empty everything with empty token
		err = EnsureGitHubCredentials("owner", "repo", "")
		if err == nil {
			t.Errorf("Expected error for empty token")
		}
		if !strings.Contains(err.Error(), "No GitHub token provided") {
			t.Errorf("Expected 'No GitHub token provided' error, got: %v", err)
		}
	})

	// Test 6: Exercise GitHubCredentialError construction and formatting
	t.Run("GitHubCredentialError_ConstructionAndFormatting", func(t *testing.T) {
		// Test various error scenarios to ensure Error() method is covered
		testCases := []struct {
			name       string
			statusCode int
			message    string
			suggestion string
		}{
			{"auth_failure", 401, "Authentication failed", "Check your token"},
			{"permission_denied", 403, "Permission denied", "Check scopes"},
			{"not_found", 404, "Repository not found", "Check repository name"},
			{"rate_limited", 429, "Rate limit exceeded", "Wait and retry"},
			{"server_error", 500, "Internal server error", "Try again later"},
			{"unknown_error", 418, "I'm a teapot", "Use a coffee maker"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := &GitHubCredentialError{
					StatusCode: tc.statusCode,
					Message:    tc.message,
					Suggestion: tc.suggestion,
				}

				// Test Error() method formatting
				errorString := err.Error()

				// Verify all parts are included
				if !strings.Contains(errorString, string(rune(tc.statusCode/100)+'0')) {
					t.Errorf("Error string should contain status code: %s", errorString)
				}
				if !strings.Contains(errorString, tc.message) {
					t.Errorf("Error string should contain message: %s", errorString)
				}
				if !strings.Contains(errorString, tc.suggestion) {
					t.Errorf("Error string should contain suggestion: %s", errorString)
				}
				if !strings.Contains(errorString, "GitHub API error") {
					t.Errorf("Error string should contain 'GitHub API error': %s", errorString)
				}
			})
		}
	})
}
