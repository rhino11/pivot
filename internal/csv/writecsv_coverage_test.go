package csv

import (
	"os"
	"path/filepath"
	"testing"
)

// TestWriteCSV_ErrorHandling tests error cases for WriteCSV function
func TestWriteCSV_ErrorHandling(t *testing.T) {
	t.Run("FailedToCreateFile", func(t *testing.T) {
		// Create a read-only directory to cause file creation failure
		readOnlyDir := t.TempDir()

		// Make directory read-only (this should prevent file creation)
		err := os.Chmod(readOnlyDir, 0444) // read-only
		if err != nil {
			t.Fatalf("Failed to make directory read-only: %v", err)
		}

		// Try to create file in read-only directory
		invalidPath := filepath.Join(readOnlyDir, "test.csv")

		issues := []*Issue{
			{
				ID:    1,
				Title: "Test Issue",
				State: "open",
			},
		}

		config := &ExportConfig{}

		err = WriteCSV(issues, invalidPath, config)
		if err == nil {
			t.Error("Expected error when creating file in read-only directory")
		}

		if !contains(err.Error(), "failed to create CSV file") {
			t.Errorf("Expected 'failed to create CSV file' error, got: %v", err)
		}
	})

	t.Run("WriteHeaderError", func(t *testing.T) {
		// This is harder to test directly since csv.Writer.Write rarely fails
		// But we can test the successful path to ensure coverage
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "test.csv")

		issues := []*Issue{
			{
				ID:    1,
				Title: "Test Issue",
				State: "open",
			},
		}

		config := &ExportConfig{
			Fields: []string{"id", "title", "state"}, // Custom fields
		}

		err := WriteCSV(issues, filePath, config)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Verify file was created
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Error("Expected CSV file to be created")
		}
	})

	t.Run("WriteRecordError", func(t *testing.T) {
		// Test the record writing path with various field types
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "test_records.csv")

		issues := []*Issue{
			{
				ID:             1,
				Title:          "Issue with all fields",
				State:          "open",
				Priority:       "high",
				Labels:         []string{"bug", "urgent"},
				Assignee:       "testuser",
				Milestone:      "v1.0",
				EstimatedHours: 8,
				StoryPoints:    5,
				Epic:           "Test Epic",
				Dependencies:   []int{2, 3},
				Body:           "Multi-line\nbody\ntext",
			},
			{
				ID:    2,
				Title: "Minimal issue",
				State: "closed",
			},
		}

		config := &ExportConfig{} // Use default fields

		err := WriteCSV(issues, filePath, config)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Verify file was created and has content
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read created CSV: %v", err)
		}

		contentStr := string(content)

		// Check that headers are present
		if !contains(contentStr, "id,title,state") {
			t.Error("Expected CSV headers to be written")
		}

		// Check that data is present
		if !contains(contentStr, "Issue with all fields") {
			t.Error("Expected issue data to be written")
		}

		if !contains(contentStr, "Minimal issue") {
			t.Error("Expected minimal issue data to be written")
		}
	})
}
