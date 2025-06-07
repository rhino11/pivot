package internal

import (
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
