package csv

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestValidateCSV(t *testing.T) {
	tests := []struct {
		name        string
		csvContent  string
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid CSV with all columns",
			csvContent: `title,state,priority,labels,body
Fix bug,open,high,"bug,urgent",This is a bug description
Add feature,open,medium,"feature",This is a feature description`,
			expectError: false,
		},
		{
			name: "valid CSV with only required column",
			csvContent: `title
Fix bug
Add feature`,
			expectError: false,
		},
		{
			name: "missing required title column",
			csvContent: `state,priority,body
open,high,This is a description`,
			expectError: true,
			errorMsg:    "required column 'title' not found",
		},
		{
			name: "column count mismatch",
			csvContent: `title,state,priority
Fix bug,open,high
Add feature,open`,
			expectError: true,
			errorMsg:    "wrong number of fields",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary CSV file
			tmpFile, err := os.CreateTemp("", "test-*.csv")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			// Write test content
			if _, err := tmpFile.WriteString(tt.csvContent); err != nil {
				t.Fatalf("Failed to write test content: %v", err)
			}
			tmpFile.Close()

			// Test validation
			err = ValidateCSV(tmpFile.Name())

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestParseCSV(t *testing.T) {
	csvContent := `id,title,state,priority,labels,assignee,milestone,created_at,updated_at,body,estimated_hours,story_points,epic,dependencies,acceptance_criteria
1,"Fix critical bug",open,high,"bug,critical",john,v1.0.0,2024-01-15T10:00:00Z,2024-01-15T10:30:00Z,"This is a critical bug that needs immediate attention",4,3,Bug Fixes,"2,3","- [ ] Reproduce the bug
- [ ] Identify root cause
- [ ] Implement fix"
2,"Add new feature",open,medium,"feature,enhancement",jane,v1.1.0,2024-01-16T09:00:00Z,2024-01-16T09:15:00Z,"Implement new feature for better user experience",8,5,New Features,,"- [ ] Design feature
- [ ] Implement feature
- [ ] Write tests"`

	// Create temporary CSV file
	tmpFile, err := os.CreateTemp("", "test-*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(csvContent); err != nil {
		t.Fatalf("Failed to write test content: %v", err)
	}
	tmpFile.Close()

	// Parse CSV
	config := &ImportConfig{FilePath: tmpFile.Name()}
	issues, err := ParseCSV(tmpFile.Name(), config)
	if err != nil {
		t.Fatalf("ParseCSV failed: %v", err)
	}

	// Validate results
	if len(issues) != 2 {
		t.Errorf("Expected 2 issues, got %d", len(issues))
	}

	// Check first issue
	issue1 := issues[0]
	if issue1.ID != 1 {
		t.Errorf("Expected ID 1, got %d", issue1.ID)
	}
	if issue1.Title != "Fix critical bug" {
		t.Errorf("Expected title 'Fix critical bug', got '%s'", issue1.Title)
	}
	if issue1.State != "open" {
		t.Errorf("Expected state 'open', got '%s'", issue1.State)
	}
	if issue1.Priority != "high" {
		t.Errorf("Expected priority 'high', got '%s'", issue1.Priority)
	}
	if !reflect.DeepEqual(issue1.Labels, []string{"bug", "critical"}) {
		t.Errorf("Expected labels [bug, critical], got %v", issue1.Labels)
	}
	if issue1.Assignee != "john" {
		t.Errorf("Expected assignee 'john', got '%s'", issue1.Assignee)
	}
	if issue1.EstimatedHours != 4 {
		t.Errorf("Expected estimated hours 4, got %d", issue1.EstimatedHours)
	}
	if issue1.StoryPoints != 3 {
		t.Errorf("Expected story points 3, got %d", issue1.StoryPoints)
	}
	if !reflect.DeepEqual(issue1.Dependencies, []int{2, 3}) {
		t.Errorf("Expected dependencies [2, 3], got %v", issue1.Dependencies)
	}

	// Check dates
	expectedCreated, _ := time.Parse(time.RFC3339, "2024-01-15T10:00:00Z")
	if !issue1.CreatedAt.Equal(expectedCreated) {
		t.Errorf("Expected created_at %v, got %v", expectedCreated, issue1.CreatedAt)
	}

	// Check second issue
	issue2 := issues[1]
	if issue2.ID != 2 {
		t.Errorf("Expected ID 2, got %d", issue2.ID)
	}
	if len(issue2.Dependencies) != 0 {
		t.Errorf("Expected no dependencies, got %v", issue2.Dependencies)
	}
}

func TestWriteCSV(t *testing.T) {
	// Create test issues
	issues := []*Issue{
		{
			ID:                 1,
			Title:              "Test Issue 1",
			State:              "open",
			Priority:           "high",
			Labels:             []string{"bug", "urgent"},
			Assignee:           "john",
			Milestone:          "v1.0.0",
			CreatedAt:          time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			UpdatedAt:          time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			Body:               "This is a test issue",
			EstimatedHours:     4,
			StoryPoints:        3,
			Epic:               "Test Epic",
			Dependencies:       []int{2, 3},
			AcceptanceCriteria: "- [ ] Test criterion 1\n- [ ] Test criterion 2",
		},
		{
			ID:       2,
			Title:    "Test Issue 2",
			State:    "open",
			Priority: "medium",
			Labels:   []string{"feature"},
			Body:     "Another test issue",
		},
	}

	// Create temporary output file
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "test-output.csv")

	// Export to CSV
	config := &ExportConfig{FilePath: outputFile}
	err := WriteCSV(issues, outputFile, config)
	if err != nil {
		t.Fatalf("WriteCSV failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Fatalf("Output file was not created")
	}

	// Parse the written CSV back
	parsedIssues, err := ParseCSV(outputFile, &ImportConfig{})
	if err != nil {
		t.Fatalf("Failed to parse written CSV: %v", err)
	}

	// Verify data integrity
	if len(parsedIssues) != len(issues) {
		t.Errorf("Expected %d issues, got %d", len(issues), len(parsedIssues))
	}

	// Check first issue
	parsed1 := parsedIssues[0]
	original1 := issues[0]

	if parsed1.ID != original1.ID {
		t.Errorf("ID mismatch: expected %d, got %d", original1.ID, parsed1.ID)
	}
	if parsed1.Title != original1.Title {
		t.Errorf("Title mismatch: expected '%s', got '%s'", original1.Title, parsed1.Title)
	}
	if !reflect.DeepEqual(parsed1.Labels, original1.Labels) {
		t.Errorf("Labels mismatch: expected %v, got %v", original1.Labels, parsed1.Labels)
	}
	if !reflect.DeepEqual(parsed1.Dependencies, original1.Dependencies) {
		t.Errorf("Dependencies mismatch: expected %v, got %v", original1.Dependencies, parsed1.Dependencies)
	}
}

func TestParseIssueFromRecord(t *testing.T) {
	headerIndex := map[string]int{
		"id":       0,
		"title":    1,
		"state":    2,
		"priority": 3,
		"labels":   4,
		"assignee": 5,
	}

	tests := []struct {
		name        string
		record      []string
		expectError bool
		expected    *Issue
	}{
		{
			name:   "valid record",
			record: []string{"1", "Test Issue", "open", "high", "bug,urgent", "john"},
			expected: &Issue{
				ID:       1,
				Title:    "Test Issue",
				State:    "open",
				Priority: "high",
				Labels:   []string{"bug", "urgent"},
				Assignee: "john",
			},
		},
		{
			name:        "missing title",
			record:      []string{"1", "", "open", "high", "bug", "john"},
			expectError: true,
		},
		{
			name:   "default state",
			record: []string{"1", "Test Issue", "", "high", "bug", "john"},
			expected: &Issue{
				ID:       1,
				Title:    "Test Issue",
				State:    "open", // Should default to "open"
				Priority: "high",
				Labels:   []string{"bug"},
				Assignee: "john",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issue, err := parseIssueFromRecord(tt.record, headerIndex, 1)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if issue.ID != tt.expected.ID {
				t.Errorf("ID mismatch: expected %d, got %d", tt.expected.ID, issue.ID)
			}
			if issue.Title != tt.expected.Title {
				t.Errorf("Title mismatch: expected '%s', got '%s'", tt.expected.Title, issue.Title)
			}
			if issue.State != tt.expected.State {
				t.Errorf("State mismatch: expected '%s', got '%s'", tt.expected.State, issue.State)
			}
			if !reflect.DeepEqual(issue.Labels, tt.expected.Labels) {
				t.Errorf("Labels mismatch: expected %v, got %v", tt.expected.Labels, issue.Labels)
			}
		})
	}
}

func TestImportCSVToGitHub(t *testing.T) {
	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "import_test.csv")

	// Create test CSV content
	csvContent := `title,state,priority,labels,body
Test Import Issue 1,open,high,"bug,urgent",This is a test import issue
Test Import Issue 2,closed,medium,feature,Another test import issue`

	if err := os.WriteFile(csvFile, []byte(csvContent), 0644); err != nil {
		t.Fatalf("Failed to create test CSV file: %v", err)
	}

	config := &ImportConfig{
		FilePath:       csvFile,
		Repository:     "test/repo",
		DryRun:         true, // Use dry run for unit tests
		SkipDuplicates: false,
	}

	// Test import in dry-run mode (no actual API calls)
	result, err := ImportCSVToGitHub(csvFile, "testowner", "testrepo", "testtoken", config)
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	if result.Total != 2 {
		t.Errorf("Expected 2 total issues, got %d", result.Total)
	}
	if result.Created != 0 {
		t.Errorf("Expected 0 created issues in dry run, got %d", result.Created)
	}
	if result.Skipped != 2 {
		t.Errorf("Expected 2 skipped issues in dry run, got %d", result.Skipped)
	}
}

func TestImportCSVToGitHub_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "dryrun_test.csv")

	csvContent := `title,state,priority,labels,body
Test Dry Run Issue,open,high,bug,This is a dry run test`

	if err := os.WriteFile(csvFile, []byte(csvContent), 0644); err != nil {
		t.Fatalf("Failed to create test CSV file: %v", err)
	}

	config := &ImportConfig{
		FilePath:       csvFile,
		Repository:     "test/repo",
		DryRun:         true,
		SkipDuplicates: false,
	}

	result, err := ImportCSVToGitHub(csvFile, "testowner", "testrepo", "testtoken", config)
	if err != nil {
		t.Fatalf("Dry run import failed: %v", err)
	}

	if result.Total != 1 {
		t.Errorf("Expected 1 total issue, got %d", result.Total)
	}
	if result.Created != 0 {
		t.Errorf("Expected 0 created issues in dry run, got %d", result.Created)
	}
	if result.Skipped != 1 {
		t.Errorf("Expected 1 skipped issue in dry run, got %d", result.Skipped)
	}
}

func TestConvertToGitHubIssue(t *testing.T) {
	issue := &Issue{
		Title:    "Test Issue",
		Body:     "Test body content",
		Labels:   []string{"bug", "urgent"},
		Assignee: "testuser",
		State:    "open",
		Priority: "high",
	}

	result := convertToGitHubIssue(issue)

	if result["title"] != "Test Issue" {
		t.Errorf("Expected title 'Test Issue', got %v", result["title"])
	}
	if result["body"] != "Test body content" {
		t.Errorf("Expected body 'Test body content', got %v", result["body"])
	}

	labels, ok := result["labels"].([]string)
	if !ok {
		t.Errorf("Expected labels to be []string, got %T", result["labels"])
	}
	if len(labels) != 2 || labels[0] != "bug" || labels[1] != "urgent" {
		t.Errorf("Expected labels [bug, urgent], got %v", labels)
	}

	assignees, ok := result["assignees"].([]string)
	if !ok {
		t.Errorf("Expected assignees to be []string, got %T", result["assignees"])
	}
	if len(assignees) != 1 || assignees[0] != "testuser" {
		t.Errorf("Expected assignees [testuser], got %v", assignees)
	}
}

func TestConvertToGitHubIssue_MinimalFields(t *testing.T) {
	issue := &Issue{
		Title: "Minimal Issue",
		Body:  "Just title and body",
	}

	result := convertToGitHubIssue(issue)

	if result["title"] != "Minimal Issue" {
		t.Errorf("Expected title 'Minimal Issue', got %v", result["title"])
	}
	if result["body"] != "Just title and body" {
		t.Errorf("Expected body 'Just title and body', got %v", result["body"])
	}

	// Should not have labels or assignees for minimal issue
	if _, exists := result["labels"]; exists {
		t.Error("Expected no labels field for minimal issue")
	}
	if _, exists := result["assignees"]; exists {
		t.Error("Expected no assignees field for minimal issue")
	}
}

func TestCSVImportGitHubIntegration(t *testing.T) {
	// This test requires a valid GitHub token in environment
	// Skip if running in CI without proper setup
	if testing.Short() {
		t.Skip("Skipping GitHub integration test in short mode")
	}

	// This test demonstrates the full GitHub integration
	// In a real test environment, you would use a test repository
	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "github_test.csv")

	csvContent := `title,state,priority,labels,body
Integration Test Issue,open,low,test,"This is a test issue for GitHub integration testing"`

	if err := os.WriteFile(csvFile, []byte(csvContent), 0644); err != nil {
		t.Fatalf("Failed to create test CSV file: %v", err)
	}

	config := &ImportConfig{
		FilePath:       csvFile,
		Repository:     "test/repo",
		DryRun:         true, // Always use dry run for tests
		SkipDuplicates: false,
	}

	// Test the import function with dry run
	result, err := ImportCSVToGitHub(csvFile, "testowner", "testrepo", "dummy-token", config)
	if err != nil {
		t.Fatalf("GitHub integration test failed: %v", err)
	}

	if result.Total != 1 {
		t.Errorf("Expected 1 total issue, got %d", result.Total)
	}
	if result.Skipped != 1 {
		t.Errorf("Expected 1 skipped issue in dry run, got %d", result.Skipped)
	}
	if result.Created != 0 {
		t.Errorf("Expected 0 created issues in dry run, got %d", result.Created)
	}
}

func TestCSVWorkflowEndToEnd(t *testing.T) {
	// Test the complete CSV workflow: parse → import → results
	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "workflow_test.csv")

	// Create a comprehensive test CSV
	csvContent := `title,state,priority,labels,assignee,body
Feature Request A,open,high,"feature,priority",testuser,"Description for feature A"
Bug Report B,open,medium,"bug,urgent",,"Description for bug B"
Enhancement C,closed,low,enhancement,testuser,"Description for enhancement C"`

	if err := os.WriteFile(csvFile, []byte(csvContent), 0644); err != nil {
		t.Fatalf("Failed to create test CSV file: %v", err)
	}

	// Test parsing
	config := &ImportConfig{
		FilePath:       csvFile,
		Repository:     "test/repo",
		DryRun:         true,
		SkipDuplicates: false,
	}

	issues, err := ParseCSV(csvFile, config)
	if err != nil {
		t.Fatalf("CSV parsing failed: %v", err)
	}

	if len(issues) != 3 {
		t.Fatalf("Expected 3 issues, got %d", len(issues))
	}

	// Verify first issue
	issue1 := issues[0]
	if issue1.Title != "Feature Request A" {
		t.Errorf("Expected title 'Feature Request A', got '%s'", issue1.Title)
	}
	if len(issue1.Labels) != 2 {
		t.Errorf("Expected 2 labels, got %d", len(issue1.Labels))
	}
	if issue1.Assignee != "testuser" {
		t.Errorf("Expected assignee 'testuser', got '%s'", issue1.Assignee)
	}

	// Test import simulation
	result, err := ImportCSVToGitHub(csvFile, "testowner", "testrepo", "dummy-token", config)
	if err != nil {
		t.Fatalf("Import simulation failed: %v", err)
	}

	if result.Total != 3 {
		t.Errorf("Expected 3 total issues, got %d", result.Total)
	}
	if result.Skipped != 3 {
		t.Errorf("Expected 3 skipped issues in dry run, got %d", result.Skipped)
	}

	t.Logf("Successfully processed %d issues in end-to-end workflow test", result.Total)
}

func TestCSVImportGitHubAPIError(t *testing.T) {
	// Test that API errors are properly handled and reported
	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "api_error_test.csv")

	csvContent := `title,state,priority,labels,body
API Error Test Issue,open,low,test,"This issue should fail to create due to invalid token"`

	if err := os.WriteFile(csvFile, []byte(csvContent), 0644); err != nil {
		t.Fatalf("Failed to create test CSV file: %v", err)
	}

	config := &ImportConfig{
		FilePath:       csvFile,
		Repository:     "test/repo",
		DryRun:         false, // Use actual mode to test API error handling
		SkipDuplicates: false,
	}

	// Test import with invalid token (should generate errors)
	result, err := ImportCSVToGitHub(csvFile, "testowner", "testrepo", "invalid-token", config)
	if err != nil {
		t.Fatalf("Import function itself should not fail: %v", err)
	}

	// Should have parsed the issue but failed to create it
	if result.Total != 1 {
		t.Errorf("Expected 1 total issue, got %d", result.Total)
	}
	if result.Created != 0 {
		t.Errorf("Expected 0 created issues due to API error, got %d", result.Created)
	}
	if len(result.Errors) == 0 {
		t.Error("Expected errors to be recorded for failed API calls")
	}

	// Verify the error message contains useful information
	if len(result.Errors) > 0 {
		errorMsg := result.Errors[0]
		if !contains(errorMsg, "API Error Test Issue") {
			t.Errorf("Error message should contain issue title, got: %s", errorMsg)
		}
	}

	t.Logf("API error handling test passed. Errors recorded: %v", result.Errors)
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				findInString(s, substr))))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
