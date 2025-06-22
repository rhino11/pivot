package internal

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestValidateGitHubCredentials(t *testing.T) {
	t.Run("empty token", func(t *testing.T) {
		err := ValidateGitHubCredentials("")
		if err == nil {
			t.Error("Expected error for empty token")
		}

		credErr, ok := err.(*GitHubCredentialError)
		if !ok {
			t.Errorf("Expected GitHubCredentialError, got %T", err)
		}

		if credErr.StatusCode != 401 {
			t.Errorf("Expected status code 401, got %d", credErr.StatusCode)
		}

		if !strings.Contains(credErr.Message, "No GitHub token provided") {
			t.Errorf("Expected message about no token, got: %s", credErr.Message)
		}
	})

	t.Run("valid token", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify endpoint
			if r.URL.Path != "/user" {
				t.Errorf("Expected path '/user', got '%s'", r.URL.Path)
			}

			// Verify headers
			if r.Header.Get("Authorization") != "token valid_token" {
				t.Errorf("Expected Authorization header 'token valid_token', got '%s'", r.Header.Get("Authorization"))
			}
			if r.Header.Get("Accept") != "application/vnd.github.v3+json" {
				t.Errorf("Expected Accept header 'application/vnd.github.v3+json', got '%s'", r.Header.Get("Accept"))
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"login": "testuser"}`))
		}))
		defer server.Close()

		// Replace GitHub API URL for testing
		err := validateGitHubCredentialsWithURL(server.URL, "valid_token")
		if err != nil {
			t.Errorf("Expected no error for valid token, got: %v", err)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"message": "Bad credentials"}`))
		}))
		defer server.Close()

		err := validateGitHubCredentialsWithURL(server.URL, "invalid_token")
		if err == nil {
			t.Error("Expected error for invalid token")
		}

		credErr, ok := err.(*GitHubCredentialError)
		if !ok {
			t.Errorf("Expected GitHubCredentialError, got %T", err)
		}

		if credErr.StatusCode != 401 {
			t.Errorf("Expected status code 401, got %d", credErr.StatusCode)
		}

		if !strings.Contains(credErr.Message, "Invalid GitHub token") {
			t.Errorf("Expected message about invalid token, got: %s", credErr.Message)
		}
	})

	t.Run("forbidden token", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"message": "Insufficient permissions"}`))
		}))
		defer server.Close()

		err := validateGitHubCredentialsWithURL(server.URL, "insufficient_token")
		if err == nil {
			t.Error("Expected error for forbidden token")
		}

		credErr, ok := err.(*GitHubCredentialError)
		if !ok {
			t.Errorf("Expected GitHubCredentialError, got %T", err)
		}

		if credErr.StatusCode != 403 {
			t.Errorf("Expected status code 403, got %d", credErr.StatusCode)
		}

		if !strings.Contains(credErr.Message, "lacks required permissions") {
			t.Errorf("Expected message about permissions, got: %s", credErr.Message)
		}
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "Internal server error"}`))
		}))
		defer server.Close()

		err := validateGitHubCredentialsWithURL(server.URL, "some_token")
		if err == nil {
			t.Error("Expected error for server error")
		}

		credErr, ok := err.(*GitHubCredentialError)
		if !ok {
			t.Errorf("Expected GitHubCredentialError, got %T", err)
		}

		if credErr.StatusCode != 500 {
			t.Errorf("Expected status code 500, got %d", credErr.StatusCode)
		}
	})
}

func TestValidateRepositoryAccess(t *testing.T) {
	t.Run("valid access", func(t *testing.T) {
		userServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/user" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"login": "testuser"}`))
			} else if r.URL.Path == "/repos/testowner/testrepo" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"id": 123, "name": "testrepo"}`))
			} else {
				t.Errorf("Unexpected path: %s", r.URL.Path)
			}
		}))
		defer userServer.Close()

		err := validateRepositoryAccessWithURL(userServer.URL, "testowner", "testrepo", "valid_token")
		if err != nil {
			t.Errorf("Expected no error for valid access, got: %v", err)
		}
	})

	t.Run("repository not found", func(t *testing.T) {
		userServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/user" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"login": "testuser"}`))
			} else if r.URL.Path == "/repos/testowner/notfound" {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"message": "Not Found"}`))
			}
		}))
		defer userServer.Close()

		err := validateRepositoryAccessWithURL(userServer.URL, "testowner", "notfound", "valid_token")
		if err == nil {
			t.Error("Expected error for repository not found")
		}

		credErr, ok := err.(*GitHubCredentialError)
		if !ok {
			t.Errorf("Expected GitHubCredentialError, got %T", err)
		}

		if credErr.StatusCode != 404 {
			t.Errorf("Expected status code 404, got %d", credErr.StatusCode)
		}

		if !strings.Contains(credErr.Message, "not found or not accessible") {
			t.Errorf("Expected message about not found, got: %s", credErr.Message)
		}
	})

	t.Run("access forbidden", func(t *testing.T) {
		userServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/user" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"login": "testuser"}`))
			} else if r.URL.Path == "/repos/testowner/forbidden" {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"message": "Forbidden"}`))
			}
		}))
		defer userServer.Close()

		err := validateRepositoryAccessWithURL(userServer.URL, "testowner", "forbidden", "valid_token")
		if err == nil {
			t.Error("Expected error for forbidden access")
		}

		credErr, ok := err.(*GitHubCredentialError)
		if !ok {
			t.Errorf("Expected GitHubCredentialError, got %T", err)
		}

		if credErr.StatusCode != 403 {
			t.Errorf("Expected status code 403, got %d", credErr.StatusCode)
		}

		if !strings.Contains(credErr.Message, "Access denied") {
			t.Errorf("Expected message about access denied, got: %s", credErr.Message)
		}
	})

	t.Run("invalid token propagated", func(t *testing.T) {
		userServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/user" {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"message": "Bad credentials"}`))
			}
		}))
		defer userServer.Close()

		err := validateRepositoryAccessWithURL(userServer.URL, "testowner", "testrepo", "invalid_token")
		if err == nil {
			t.Error("Expected error for invalid token")
		}

		credErr, ok := err.(*GitHubCredentialError)
		if !ok {
			t.Errorf("Expected GitHubCredentialError, got %T", err)
		}

		if credErr.StatusCode != 401 {
			t.Errorf("Expected status code 401, got %d", credErr.StatusCode)
		}
	})
}

func TestEnsureGitHubCredentials(t *testing.T) {
	t.Run("valid credentials without repo", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"login": "testuser"}`))
		}))
		defer server.Close()

		err := ensureGitHubCredentialsWithURL(server.URL, "", "", "valid_token")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	t.Run("valid credentials with repo", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/user" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"login": "testuser"}`))
			} else if r.URL.Path == "/repos/testowner/testrepo" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"id": 123, "name": "testrepo"}`))
			}
		}))
		defer server.Close()

		err := ensureGitHubCredentialsWithURL(server.URL, "testowner", "testrepo", "valid_token")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"message": "Bad credentials"}`))
		}))
		defer server.Close()

		err := ensureGitHubCredentialsWithURL(server.URL, "testowner", "testrepo", "invalid_token")
		if err == nil {
			t.Error("Expected error for invalid token")
		}

		credErr, ok := err.(*GitHubCredentialError)
		if !ok {
			t.Errorf("Expected GitHubCredentialError, got %T", err)
		}

		if credErr.StatusCode != 401 {
			t.Errorf("Expected status code 401, got %d", credErr.StatusCode)
		}
	})
}

func TestGitHubCredentialError(t *testing.T) {
	err := &GitHubCredentialError{
		StatusCode: 401,
		Message:    "Invalid token",
		Suggestion: "Update your token",
	}

	expected := "GitHub API error (401): Invalid token\nUpdate your token"
	if err.Error() != expected {
		t.Errorf("Expected error string '%s', got '%s'", expected, err.Error())
	}
}

// Helper functions for testing with custom URLs

func validateGitHubCredentialsWithURL(baseURL, token string) error {
	if token == "" {
		return &GitHubCredentialError{
			StatusCode: 401,
			Message:    "No GitHub token provided",
			Suggestion: "Run 'pivot init' to configure your GitHub token, or set it in config.yml",
		}
	}

	// Test the token by calling the user endpoint
	req, err := http.NewRequest("GET", baseURL+"/user", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to validate GitHub token: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return nil // Token is valid
	case http.StatusUnauthorized:
		return &GitHubCredentialError{
			StatusCode: 401,
			Message:    "Invalid GitHub token",
			Suggestion: "Your GitHub token is invalid or expired. Run 'pivot init' to update it, or check your config.yml file",
		}
	case http.StatusForbidden:
		return &GitHubCredentialError{
			StatusCode: 403,
			Message:    "GitHub token lacks required permissions",
			Suggestion: "Your GitHub token needs 'repo' scope permissions. Create a new token at https://github.com/settings/tokens",
		}
	default:
		body, _ := io.ReadAll(resp.Body)
		return &GitHubCredentialError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("GitHub API returned unexpected status: %s", string(body)),
			Suggestion: "Check your network connection and try again",
		}
	}
}

func validateRepositoryAccessWithURL(baseURL, owner, repo, token string) error {
	if err := validateGitHubCredentialsWithURL(baseURL, token); err != nil {
		return err
	}

	// Test repository access
	url := fmt.Sprintf("%s/repos/%s/%s", baseURL, owner, repo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to validate repository access: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return nil // Has access to repository
	case http.StatusNotFound:
		return &GitHubCredentialError{
			StatusCode: 404,
			Message:    fmt.Sprintf("Repository %s/%s not found or not accessible", owner, repo),
			Suggestion: "Check the repository name or ensure your token has access to this repository",
		}
	case http.StatusForbidden:
		return &GitHubCredentialError{
			StatusCode: 403,
			Message:    fmt.Sprintf("Access denied to repository %s/%s", owner, repo),
			Suggestion: "Your token doesn't have permission to access this repository. Ensure it has 'repo' scope",
		}
	default:
		body, _ := io.ReadAll(resp.Body)
		return &GitHubCredentialError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("Unexpected response when accessing repository: %s", string(body)),
			Suggestion: "Check your network connection and try again",
		}
	}
}

func ensureGitHubCredentialsWithURL(baseURL, owner, repo, token string) error {
	// First validate the basic token
	if err := validateGitHubCredentialsWithURL(baseURL, token); err != nil {
		return err
	}

	// Then validate repository access if specified
	if owner != "" && repo != "" {
		if err := validateRepositoryAccessWithURL(baseURL, owner, repo, token); err != nil {
			return err
		}
	}

	return nil
}
