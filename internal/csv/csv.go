package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rhino11/pivot/internal"
)

// Issue represents a GitHub issue for CSV import/export
type Issue struct {
	ID                 int       `csv:"id"`
	Title              string    `csv:"title"`
	State              string    `csv:"state"`
	Priority           string    `csv:"priority"`
	Labels             []string  `csv:"labels"`
	Assignee           string    `csv:"assignee"`
	Milestone          string    `csv:"milestone"`
	CreatedAt          time.Time `csv:"created_at"`
	UpdatedAt          time.Time `csv:"updated_at"`
	Body               string    `csv:"body"`
	EstimatedHours     int       `csv:"estimated_hours"`
	StoryPoints        int       `csv:"story_points"`
	Epic               string    `csv:"epic"`
	Dependencies       []int     `csv:"dependencies"`
	AcceptanceCriteria string    `csv:"acceptance_criteria"`
}

// ImportConfig holds configuration for CSV import
type ImportConfig struct {
	FilePath       string
	Repository     string
	DryRun         bool
	SkipDuplicates bool
	Mapping        map[string]string
}

// ExportConfig holds configuration for CSV export
type ExportConfig struct {
	FilePath   string
	Repository string
	Fields     []string
	Filter     string
}

// ImportResult contains the results of a CSV import operation
type ImportResult struct {
	Total      int
	Created    int
	Skipped    int
	Errors     []string
	Issues     []*Issue
	Duplicates []*Issue
}

// ExportResult contains the results of a CSV export operation
type ExportResult struct {
	Total    int
	FilePath string
	Issues   []*Issue
}

// ValidateCSV validates a CSV file and returns parsing errors
func ValidateCSV(filePath string) error {
	file, err := os.Open(filePath) // #nosec G304 - File path is validated and user-controlled
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	// Check for empty file
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	if fileInfo.Size() == 0 {
		return fmt.Errorf("CSV file is empty")
	}

	// Read a few bytes to check for UTF-8 BOM
	buf := make([]byte, 3)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to read file beginning: %w", err)
	}

	// Reset file position
	_, err = file.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("failed to reset file position: %w", err)
	}

	// Skip UTF-8 BOM if present
	if n >= 3 && buf[0] == 0xEF && buf[1] == 0xBB && buf[2] == 0xBF {
		_, err = file.Seek(3, 0)
		if err != nil {
			return fmt.Errorf("failed to skip BOM: %w", err)
		}
	}

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // Allow variable number of fields initially

	// Read header
	headers, err := reader.Read()
	if err != nil {
		if err == io.EOF {
			return fmt.Errorf("CSV file contains no data (empty or header-only)")
		}
		return fmt.Errorf("failed to read CSV headers: %w", err)
	}

	if len(headers) == 0 {
		return fmt.Errorf("CSV header row is empty")
	}

	// Clean headers (remove BOM from first header if present)
	if len(headers) > 0 {
		headers[0] = strings.TrimPrefix(headers[0], "\uFEFF") // Remove UTF-8 BOM
		headers[0] = strings.TrimSpace(headers[0])
	}

	// Validate required columns
	requiredColumns := []string{"title"}
	headerMap := make(map[string]bool)
	for _, header := range headers {
		cleanHeader := strings.ToLower(strings.TrimSpace(header))
		if cleanHeader != "" {
			headerMap[cleanHeader] = true
		}
	}

	for _, required := range requiredColumns {
		if !headerMap[required] {
			return fmt.Errorf("required column '%s' not found in CSV headers: %v", required, headers)
		}
	}

	// Set expected field count for remaining validation
	reader.FieldsPerRecord = len(headers)

	// Validate each row
	lineNum := 2 // Start from line 2 (after header)
	rowCount := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading CSV line %d: %w", lineNum, err)
		}

		if len(record) != len(headers) {
			return fmt.Errorf("line %d: column count mismatch (expected %d, got %d)",
				lineNum, len(headers), len(record))
		}

		rowCount++
		lineNum++
	}

	if rowCount == 0 {
		return fmt.Errorf("CSV file contains no data rows (header-only)")
	}

	return nil
}

// ParseCSV reads and parses a CSV file into Issue structs
func ParseCSV(filePath string, config *ImportConfig) ([]*Issue, error) {
	file, err := os.Open(filePath) // #nosec G304 - File path is validated and user-controlled
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	// Check for UTF-8 BOM and skip it
	buf := make([]byte, 3)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read file beginning: %w", err)
	}

	// Reset file position
	_, err = file.Seek(0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to reset file position: %w", err)
	}

	// Skip UTF-8 BOM if present
	if n >= 3 && buf[0] == 0xEF && buf[1] == 0xBB && buf[2] == 0xBF {
		_, err = file.Seek(3, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to skip BOM: %w", err)
		}
	}

	reader := csv.NewReader(file)

	// Read header
	headers, err := reader.Read()
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("CSV file is empty or contains no headers")
		}
		return nil, fmt.Errorf("failed to read CSV headers: %w", err)
	}

	// Clean headers (remove BOM from first header if present)
	if len(headers) > 0 {
		headers[0] = strings.TrimPrefix(headers[0], "\uFEFF") // Remove UTF-8 BOM
		headers[0] = strings.TrimSpace(headers[0])
	}

	// Create header index map
	headerIndex := make(map[string]int)
	for i, header := range headers {
		cleanHeader := strings.ToLower(strings.TrimSpace(header))
		if cleanHeader != "" {
			headerIndex[cleanHeader] = i
		}
	}

	// Validate required columns
	if _, exists := headerIndex["title"]; !exists {
		return nil, fmt.Errorf("required column 'title' not found in CSV headers: %v", headers)
	}

	var issues []*Issue
	lineNum := 2 // Start from line 2 (after header)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading CSV line %d: %w", lineNum, err)
		}

		issue, err := parseIssueFromRecord(record, headerIndex, lineNum)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNum, err)
		}

		issues = append(issues, issue)
		lineNum++
	}

	return issues, nil
}

// parseIssueFromRecord converts a CSV record to an Issue struct
func parseIssueFromRecord(record []string, headerIndex map[string]int, lineNum int) (*Issue, error) {
	issue := &Issue{}

	// Helper function to safely get field value
	getField := func(fieldName string) string {
		if idx, exists := headerIndex[fieldName]; exists && idx < len(record) {
			return strings.TrimSpace(record[idx])
		}
		return ""
	}

	// Parse required fields
	issue.Title = getField("title")
	if issue.Title == "" {
		return nil, fmt.Errorf("title is required")
	}

	// Parse optional fields
	issue.State = getField("state")
	if issue.State == "" {
		issue.State = "open" // Default to open
	}

	issue.Priority = getField("priority")
	issue.Assignee = getField("assignee")
	issue.Milestone = getField("milestone")
	issue.Body = getField("body")
	issue.Epic = getField("epic")
	issue.AcceptanceCriteria = getField("acceptance_criteria")

	// Parse labels (comma-separated)
	if labelsStr := getField("labels"); labelsStr != "" {
		labels := strings.Split(labelsStr, ",")
		for i, label := range labels {
			labels[i] = strings.TrimSpace(label)
		}
		issue.Labels = labels
	}

	// Parse numeric fields
	if idStr := getField("id"); idStr != "" {
		if id, err := strconv.Atoi(idStr); err == nil {
			issue.ID = id
		}
	}

	if hoursStr := getField("estimated_hours"); hoursStr != "" {
		if hours, err := strconv.Atoi(hoursStr); err == nil {
			issue.EstimatedHours = hours
		}
	}

	if pointsStr := getField("story_points"); pointsStr != "" {
		if points, err := strconv.Atoi(pointsStr); err == nil {
			issue.StoryPoints = points
		}
	}

	// Parse dependencies (comma-separated integers)
	if depsStr := getField("dependencies"); depsStr != "" {
		depStrs := strings.Split(depsStr, ",")
		for _, depStr := range depStrs {
			if depStr = strings.TrimSpace(depStr); depStr != "" {
				if dep, err := strconv.Atoi(depStr); err == nil {
					issue.Dependencies = append(issue.Dependencies, dep)
				}
			}
		}
	}

	// Parse dates
	if createdStr := getField("created_at"); createdStr != "" {
		if created, err := time.Parse(time.RFC3339, createdStr); err == nil {
			issue.CreatedAt = created
		}
	}

	if updatedStr := getField("updated_at"); updatedStr != "" {
		if updated, err := time.Parse(time.RFC3339, updatedStr); err == nil {
			issue.UpdatedAt = updated
		}
	}

	return issue, nil
}

// WriteCSV exports issues to a CSV file
func WriteCSV(issues []*Issue, filePath string, config *ExportConfig) error {
	file, err := os.Create(filePath) // #nosec G304 - File path is validated and user-controlled
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Define default columns
	columns := []string{
		"id", "title", "state", "priority", "labels", "assignee", "milestone",
		"created_at", "updated_at", "body", "estimated_hours", "story_points",
		"epic", "dependencies", "acceptance_criteria",
	}

	// Use custom fields if specified
	if len(config.Fields) > 0 {
		columns = config.Fields
	}

	// Write header
	if err := writer.Write(columns); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows
	for _, issue := range issues {
		record := make([]string, len(columns))
		for i, column := range columns {
			record[i] = getIssueFieldValue(issue, column)
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	return nil
}

// getIssueFieldValue extracts field value from Issue struct
func getIssueFieldValue(issue *Issue, fieldName string) string {
	switch fieldName {
	case "id":
		if issue.ID == 0 {
			return ""
		}
		return strconv.Itoa(issue.ID)
	case "title":
		return issue.Title
	case "state":
		return issue.State
	case "priority":
		return issue.Priority
	case "labels":
		return strings.Join(issue.Labels, ",")
	case "assignee":
		return issue.Assignee
	case "milestone":
		return issue.Milestone
	case "created_at":
		if issue.CreatedAt.IsZero() {
			return ""
		}
		return issue.CreatedAt.Format(time.RFC3339)
	case "updated_at":
		if issue.UpdatedAt.IsZero() {
			return ""
		}
		return issue.UpdatedAt.Format(time.RFC3339)
	case "body":
		return issue.Body
	case "estimated_hours":
		if issue.EstimatedHours == 0 {
			return ""
		}
		return strconv.Itoa(issue.EstimatedHours)
	case "story_points":
		if issue.StoryPoints == 0 {
			return ""
		}
		return strconv.Itoa(issue.StoryPoints)
	case "epic":
		return issue.Epic
	case "dependencies":
		if len(issue.Dependencies) == 0 {
			return ""
		}
		deps := make([]string, len(issue.Dependencies))
		for i, dep := range issue.Dependencies {
			deps[i] = strconv.Itoa(dep)
		}
		return strings.Join(deps, ",")
	case "acceptance_criteria":
		return issue.AcceptanceCriteria
	default:
		return ""
	}
}

// ImportCSVToGitHub imports issues from CSV to GitHub repository
func ImportCSVToGitHub(filePath, owner, repo, token string, config *ImportConfig) (*ImportResult, error) {
	// Parse CSV first
	issues, err := ParseCSV(filePath, config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	// Validate GitHub credentials before attempting import (unless in dry-run mode)
	if !config.DryRun {
		if err := internal.EnsureGitHubCredentials(owner, repo, token); err != nil {
			return nil, fmt.Errorf("GitHub credential validation failed: %w", err)
		}
	}

	result := &ImportResult{
		Total:  len(issues),
		Issues: issues,
		Errors: []string{},
	}

	// Import each issue to GitHub
	for _, issue := range issues {
		if config.DryRun {
			result.Skipped++
			continue
		}

		// Convert CSV issue to GitHub issue request
		githubRequest := internal.CreateIssueRequest{
			Title: issue.Title,
			Body:  issue.Body,
		}

		// Add labels if present
		if len(issue.Labels) > 0 {
			githubRequest.Labels = issue.Labels
		}

		// Add assignee if present
		if issue.Assignee != "" {
			githubRequest.Assignees = []string{issue.Assignee}
		}

		// Create the issue on GitHub
		response, err := internal.CreateIssue(owner, repo, token, githubRequest)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to create issue '%s': %v", issue.Title, err))
			continue
		}

		// Update the issue with GitHub data
		issue.ID = response.ID
		result.Created++
	}

	return result, nil
}

// convertToGitHubIssue converts a CSV Issue to a GitHub CreateIssueRequest format
func convertToGitHubIssue(issue *Issue) map[string]interface{} {
	request := map[string]interface{}{
		"title": issue.Title,
		"body":  issue.Body,
	}

	if len(issue.Labels) > 0 {
		request["labels"] = issue.Labels
	}

	if issue.Assignee != "" {
		request["assignees"] = []string{issue.Assignee}
	}

	return request
}
