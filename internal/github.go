package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Issue struct {
	ID        int    `json:"id"`
	Number    int    `json:"number"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	State     string `json:"state"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	ClosedAt  string `json:"closed_at"`
	Labels    []struct {
		Name string `json:"name"`
	} `json:"labels"`
	Assignees []struct {
		Login string `json:"login"`
	} `json:"assignees"`
}

func FetchIssues(owner, repo, token string) ([]Issue, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues?state=all&per_page=100", owner, repo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Validate credentials after successful request creation
	if err := EnsureGitHubCredentials(owner, repo, token); err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			return nil, &GitHubCredentialError{
				StatusCode: 401,
				Message:    "Authentication failed",
				Suggestion: "Your GitHub token is invalid or expired. Run 'pivot init' to update it",
			}
		case http.StatusForbidden:
			return nil, &GitHubCredentialError{
				StatusCode: 403,
				Message:    "Access forbidden to repository issues",
				Suggestion: "Your GitHub token needs 'repo' scope permissions to access repository issues",
			}
		case http.StatusNotFound:
			return nil, &GitHubCredentialError{
				StatusCode: 404,
				Message:    fmt.Sprintf("Repository %s/%s not found", owner, repo),
				Suggestion: "Check the repository name or ensure your token has access to this repository",
			}
		default:
			return nil, fmt.Errorf("GitHub API error (%d): %s", resp.StatusCode, string(body))
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var issues []Issue
	if err := json.Unmarshal(body, &issues); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return issues, nil
}

// CreateIssueRequest represents the request payload for creating a GitHub issue
type CreateIssueRequest struct {
	Title     string   `json:"title"`
	Body      string   `json:"body,omitempty"`
	Labels    []string `json:"labels,omitempty"`
	Assignees []string `json:"assignees,omitempty"`
	Milestone int      `json:"milestone,omitempty"`
}

// CreateIssueResponse represents the response from GitHub when creating an issue
type CreateIssueResponse struct {
	ID      int    `json:"id"`
	Number  int    `json:"number"`
	Title   string `json:"title"`
	State   string `json:"state"`
	HTMLURL string `json:"html_url"`
}

// CreateIssue creates a new GitHub issue
func CreateIssue(owner, repo, token string, request CreateIssueRequest) (*CreateIssueResponse, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues", owner, repo)

	payload, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Validate credentials after successful request creation
	if err := EnsureGitHubCredentials(owner, repo, token); err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		// Provide specific error messages for common issues
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			return nil, &GitHubCredentialError{
				StatusCode: 401,
				Message:    "Authentication failed during issue creation",
				Suggestion: "Your GitHub token is invalid or expired. Run 'pivot init' to update it",
			}
		case http.StatusForbidden:
			return nil, &GitHubCredentialError{
				StatusCode: 403,
				Message:    "Access denied - cannot create issues in this repository",
				Suggestion: "Your GitHub token needs 'repo' scope permissions to create issues",
			}
		case http.StatusNotFound:
			return nil, &GitHubCredentialError{
				StatusCode: 404,
				Message:    fmt.Sprintf("Repository %s/%s not found", owner, repo),
				Suggestion: "Check the repository name or ensure your token has access to this repository",
			}
		case http.StatusUnprocessableEntity:
			return nil, fmt.Errorf("validation failed: %s", string(body))
		default:
			return nil, fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(body))
		}
	}

	var issueResponse CreateIssueResponse
	if err := json.Unmarshal(body, &issueResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &issueResponse, nil
}

// GitHubCredentialError represents authentication/authorization errors
type GitHubCredentialError struct {
	StatusCode int
	Message    string
	Suggestion string
}

func (e GitHubCredentialError) Error() string {
	return fmt.Sprintf("GitHub API error (%d): %s\n%s", e.StatusCode, e.Message, e.Suggestion)
}

// ValidateGitHubCredentials validates a GitHub token by making a test API call
func ValidateGitHubCredentials(token string) error {
	if token == "" {
		return &GitHubCredentialError{
			StatusCode: 401,
			Message:    "No GitHub token provided",
			Suggestion: "Run 'pivot init' to configure your GitHub token, or set it in config.yml",
		}
	}

	// Test the token by calling the user endpoint
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
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

// ValidateRepositoryAccess validates that the token has access to a specific repository
func ValidateRepositoryAccess(owner, repo, token string) error {
	if err := ValidateGitHubCredentials(token); err != nil {
		return err
	}

	// Test repository access
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)
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

// EnsureGitHubCredentials validates credentials and provides user-friendly error messages
func EnsureGitHubCredentials(owner, repo, token string) error {
	// First validate the basic token
	if err := ValidateGitHubCredentials(token); err != nil {
		return err
	}

	// Then validate repository access if specified
	if owner != "" && repo != "" {
		if err := ValidateRepositoryAccess(owner, repo, token); err != nil {
			return err
		}
	}

	return nil
}
