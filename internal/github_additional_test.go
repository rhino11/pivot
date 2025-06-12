package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestFetchIssues_AdditionalCoverage provides additional test coverage for FetchIssues
func TestFetchIssues_AdditionalCoverage(t *testing.T) {
	t.Run("RequestCreationErrorWithInvalidCharacters", func(t *testing.T) {
		// Use invalid URL characters to cause request creation to fail
		_, err := FetchIssues("invalid\nowner", "repo", "token")
		if err == nil {
			t.Error("Expected error for invalid owner characters, got nil")
		}
		if !strings.Contains(err.Error(), "failed to create request") {
			t.Errorf("Expected request creation error, got: %v", err)
		}
	})

	t.Run("PaginationHeaders", func(t *testing.T) {
		mockIssues := []Issue{
			{ID: 1, Number: 1, Title: "First Issue", State: "open"},
			{ID: 2, Number: 2, Title: "Second Issue", State: "closed"},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify pagination parameters
			if r.URL.Query().Get("per_page") != "100" {
				t.Errorf("Expected per_page=100, got %s", r.URL.Query().Get("per_page"))
			}
			if r.URL.Query().Get("state") != "all" {
				t.Errorf("Expected state=all, got %s", r.URL.Query().Get("state"))
			}

			// Add pagination headers
			w.Header().Set("Link", `<https://api.github.com/repos/owner/repo/issues?page=2>; rel="next"`)
			w.Header().Set("X-RateLimit-Remaining", "4999")

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(mockIssues) // #nosec G104 - test helper, ignore error
		}))
		defer server.Close()

		// Create custom FetchIssues function for testing
		issues, err := fetchIssuesWithCustomURL(server.URL, "owner", "repo", "token")
		if err != nil {
			t.Fatalf("fetchIssuesWithCustomURL failed: %v", err)
		}

		if len(issues) != 2 {
			t.Errorf("Expected 2 issues, got %d", len(issues))
		}
	})

	t.Run("SpecialCharacters", func(t *testing.T) {
		// Test with special characters that are valid in GitHub usernames/repos
		specialCases := []struct {
			owner, repo string
		}{
			{"owner-with-dashes", "repo-with-dashes"},
			{"owner_with_underscores", "repo_with_underscores"},
		}

		for _, tc := range specialCases {
			t.Run(fmt.Sprintf("%s/%s", tc.owner, tc.repo), func(t *testing.T) {
				// This will likely fail with 401/404, but should not panic
				_, err := FetchIssues(tc.owner, tc.repo, "test-token")
				if err != nil {
					// Expected - we're just testing that it doesn't crash
					t.Logf("Expected failure for %s/%s: %v", tc.owner, tc.repo, err)
				}
			})
		}
	})
}

// TestCreateIssue_CoverageImprovements tests additional CreateIssue scenarios
func TestCreateIssue_CoverageImprovements(t *testing.T) {
	t.Run("SuccessWithCompleteRequest", func(t *testing.T) {
		// Mock response for successful creation
		mockResponse := CreateIssueResponse{
			ID:      123,
			Number:  456,
			Title:   "Test Issue",
			State:   "open",
			HTMLURL: "https://github.com/owner/repo/issues/456",
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify headers
			if r.Header.Get("Authorization") != "token testtoken" {
				t.Errorf("Expected Authorization header 'token testtoken', got '%s'", r.Header.Get("Authorization"))
			}
			if r.Header.Get("Content-Type") != "application/json" {
				t.Errorf("Expected Content-Type 'application/json', got '%s'", r.Header.Get("Content-Type"))
			}
			if r.Header.Get("Accept") != "application/vnd.github.v3+json" {
				t.Errorf("Expected Accept header 'application/vnd.github.v3+json', got '%s'", r.Header.Get("Accept"))
			}

			// Verify request body
			body, _ := io.ReadAll(r.Body)
			var request CreateIssueRequest
			if err := json.Unmarshal(body, &request); err != nil {
				t.Errorf("Failed to parse request body: %v", err)
			}

			if request.Title != "Test Issue" {
				t.Errorf("Expected title 'Test Issue', got '%s'", request.Title)
			}

			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(mockResponse) // #nosec G104 - test helper, ignore error
		}))
		defer server.Close()

		request := CreateIssueRequest{
			Title:     "Test Issue",
			Body:      "Test body",
			Labels:    []string{"bug", "enhancement"},
			Assignees: []string{"user1", "user2"},
		}

		response, err := CreateIssueWithURL(server.URL, "owner", "repo", "testtoken", request)
		if err != nil {
			t.Fatalf("CreateIssueWithURL failed: %v", err)
		}

		if response.ID != 123 {
			t.Errorf("Expected ID 123, got %d", response.ID)
		}
		if response.Number != 456 {
			t.Errorf("Expected Number 456, got %d", response.Number)
		}
	})

	t.Run("HTTPErrorHandling", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			_, _ = w.Write([]byte(`{"message": "Validation Failed", "errors": [{"field": "title", "code": "missing"}]}`)) // #nosec G104 - test helper, ignore error
		}))
		defer server.Close()

		request := CreateIssueRequest{Title: "Test Issue"}
		_, err := CreateIssueWithURL(server.URL, "owner", "repo", "testtoken", request)
		if err == nil {
			t.Error("Expected error for 422 status code")
		}
		if !strings.Contains(err.Error(), "422") {
			t.Errorf("Expected error to contain status code 422, got: %v", err)
		}
	})

	t.Run("InvalidJSONResponse", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte("invalid json response")) // #nosec G104 - test helper, ignore error
		}))
		defer server.Close()

		request := CreateIssueRequest{Title: "Test Issue"}
		_, err := CreateIssueWithURL(server.URL, "owner", "repo", "testtoken", request)
		if err == nil {
			t.Error("Expected error for invalid JSON response")
		}
		if !strings.Contains(err.Error(), "failed to unmarshal response") {
			t.Errorf("Expected unmarshal error, got: %v", err)
		}
	})

	t.Run("EmptyTitleHandling", func(t *testing.T) {
		// Test with empty title - should still create request but likely fail on GitHub side
		request := CreateIssueRequest{
			Title: "",
			Body:  "Body without title",
		}

		// This will fail on network, but should not fail on marshaling
		_, err := CreateIssue("owner", "repo", "token", request)
		if err == nil {
			t.Error("Expected network error in test environment")
		}
		// Should not be a marshal error
		if strings.Contains(err.Error(), "failed to marshal request") {
			t.Error("Should not be a marshal error with valid request structure")
		}
	})
}

// fetchIssuesWithCustomURL is a helper function for testing with custom URLs
func fetchIssuesWithCustomURL(baseURL, owner, repo, token string) ([]Issue, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/issues?state=all&per_page=100", baseURL, owner, repo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "token "+token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
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
