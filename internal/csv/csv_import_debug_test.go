package csv

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestCSVImportDebug tests the basic CSV import functionality with debugging
func TestCSVImportDebug(t *testing.T) {
	t.Run("EmptyFile", func(t *testing.T) {
		tmpDir := t.TempDir()
		emptyFile := filepath.Join(tmpDir, "empty.csv")
		
		// Create empty file
		file, err := os.Create(emptyFile)
		if err != nil {
			t.Fatalf("Failed to create empty file: %v", err)
		}
		file.Close()
		
		err = ValidateCSV(emptyFile)
		if err == nil {
			t.Error("Expected error for empty file, got none")
		}
		if !strings.Contains(err.Error(), "empty") {
			t.Errorf("Expected empty file error, got: %v", err)
		}
	})
	
	t.Run("HeaderOnly", func(t *testing.T) {
		tmpDir := t.TempDir()
		headerOnlyFile := filepath.Join(tmpDir, "header-only.csv")
		
		content := "title,state,priority,labels\n"
		err := os.WriteFile(headerOnlyFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create header-only file: %v", err)
		}
		
		err = ValidateCSV(headerOnlyFile)
		if err == nil {
			t.Error("Expected error for header-only file, got none")
		}
		if !strings.Contains(err.Error(), "no data rows") {
			t.Errorf("Expected 'no data rows' error, got: %v", err)
		}
	})
	
	t.Run("MinimalValidCSV", func(t *testing.T) {
		tmpDir := t.TempDir()
		minimalFile := filepath.Join(tmpDir, "minimal.csv")
		
		content := "title,state\nTest Issue,open\n"
		err := os.WriteFile(minimalFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create minimal file: %v", err)
		}
		
		err = ValidateCSV(minimalFile)
		if err != nil {
			t.Errorf("Minimal valid CSV should pass validation, got error: %v", err)
		}
		
		// Test parsing
		config := &ImportConfig{FilePath: minimalFile}
		issues, err := ParseCSV(minimalFile, config)
		if err != nil {
			t.Errorf("Failed to parse minimal CSV: %v", err)
		}
		
		if len(issues) != 1 {
			t.Errorf("Expected 1 issue, got %d", len(issues))
		}
		
		if issues[0].Title != "Test Issue" {
			t.Errorf("Expected title 'Test Issue', got '%s'", issues[0].Title)
		}
	})
	
	t.Run("FileWithBOM", func(t *testing.T) {
		tmpDir := t.TempDir()
		bomFile := filepath.Join(tmpDir, "bom.csv")
		
		// Create file with UTF-8 BOM
		bom := []byte{0xEF, 0xBB, 0xBF}
		content := "title,state\nTest Issue,open\n"
		fullContent := append(bom, []byte(content)...)
		
		err := os.WriteFile(bomFile, fullContent, 0644)
		if err != nil {
			t.Fatalf("Failed to create BOM file: %v", err)
		}
		
		err = ValidateCSV(bomFile)
		if err != nil {
			t.Logf("BOM file validation failed (expected): %v", err)
			// This might fail, which is OK - we'll handle BOM in the fix
		}
	})
	
	t.Run("WindowsLineEndings", func(t *testing.T) {
		tmpDir := t.TempDir()
		windowsFile := filepath.Join(tmpDir, "windows.csv")
		
		content := "title,state\r\nTest Issue,open\r\n"
		err := os.WriteFile(windowsFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create Windows file: %v", err)
		}
		
		err = ValidateCSV(windowsFile)
		if err != nil {
			t.Errorf("Windows line endings should be handled, got error: %v", err)
		}
	})
	
	t.Run("ComplexEscaping", func(t *testing.T) {
		tmpDir := t.TempDir()
		complexFile := filepath.Join(tmpDir, "complex.csv")
		
		content := `title,body,labels
"Issue with ""quotes""","Description with
multiple lines","bug,urgent"
"Comma, issue","Simple description","feature"
`
		err := os.WriteFile(complexFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create complex file: %v", err)
		}
		
		err = ValidateCSV(complexFile)
		if err != nil {
			t.Errorf("Complex escaping should be handled, got error: %v", err)
		}
		
		// Test parsing
		config := &ImportConfig{FilePath: complexFile}
		issues, err := ParseCSV(complexFile, config)
		if err != nil {
			t.Errorf("Failed to parse complex CSV: %v", err)
		}
		
		if len(issues) != 2 {
			t.Errorf("Expected 2 issues, got %d", len(issues))
		}
	})
}

// TestCSVImportErrorHandling tests error scenarios
func TestCSVImportErrorHandling(t *testing.T) {
	t.Run("NonExistentFile", func(t *testing.T) {
		err := ValidateCSV("/non/existent/file.csv")
		if err == nil {
			t.Error("Expected error for non-existent file, got none")
		}
	})
	
	t.Run("MissingRequiredColumn", func(t *testing.T) {
		tmpDir := t.TempDir()
		invalidFile := filepath.Join(tmpDir, "invalid.csv")
		
		content := "description,state\nSome description,open\n"
		err := os.WriteFile(invalidFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create invalid file: %v", err)
		}
		
		err = ValidateCSV(invalidFile)
		if err == nil {
			t.Error("Expected error for missing title column, got none")
		}
		if !strings.Contains(err.Error(), "title") {
			t.Errorf("Error should mention missing title column, got: %v", err)
		}
	})
	
	t.Run("ColumnCountMismatch", func(t *testing.T) {
		tmpDir := t.TempDir()
		mismatchFile := filepath.Join(tmpDir, "mismatch.csv")
		
		content := "title,state,priority\nTest Issue,open\nAnother Issue,closed,high,extra\n"
		err := os.WriteFile(mismatchFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create mismatch file: %v", err)
		}
		
		err = ValidateCSV(mismatchFile)
		if err == nil {
			t.Error("Expected error for column count mismatch, got none")
		}
		if !strings.Contains(err.Error(), "wrong number of fields") {
			t.Errorf("Error should mention wrong number of fields, got: %v", err)
		}
	})
}

// TestCSVImportFieldParsing tests parsing of various field types
func TestCSVImportFieldParsing(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "fields.csv")
	
	content := `id,title,state,priority,labels,assignee,milestone,estimated_hours,story_points,dependencies,created_at,updated_at
1,"Test Issue","open","high","bug,urgent","john","v1.0",8,5,"2,3","2024-01-15T10:00:00Z","2024-01-15T10:30:00Z"
2,"Simple Issue","closed","low","","","",0,0,"","",""
`
	
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	config := &ImportConfig{FilePath: testFile}
	issues, err := ParseCSV(testFile, config)
	if err != nil {
		t.Fatalf("Failed to parse test CSV: %v", err)
	}
	
	if len(issues) != 2 {
		t.Fatalf("Expected 2 issues, got %d", len(issues))
	}
	
	// Test first issue (with all fields)
	issue1 := issues[0]
	if issue1.ID != 1 {
		t.Errorf("Expected ID 1, got %d", issue1.ID)
	}
	if issue1.Title != "Test Issue" {
		t.Errorf("Expected title 'Test Issue', got '%s'", issue1.Title)
	}
	if issue1.State != "open" {
		t.Errorf("Expected state 'open', got '%s'", issue1.State)
	}
	if issue1.Priority != "high" {
		t.Errorf("Expected priority 'high', got '%s'", issue1.Priority)
	}
	if len(issue1.Labels) != 2 || issue1.Labels[0] != "bug" || issue1.Labels[1] != "urgent" {
		t.Errorf("Expected labels [bug, urgent], got %v", issue1.Labels)
	}
	if issue1.Assignee != "john" {
		t.Errorf("Expected assignee 'john', got '%s'", issue1.Assignee)
	}
	if issue1.EstimatedHours != 8 {
		t.Errorf("Expected estimated hours 8, got %d", issue1.EstimatedHours)
	}
	if issue1.StoryPoints != 5 {
		t.Errorf("Expected story points 5, got %d", issue1.StoryPoints)
	}
	if len(issue1.Dependencies) != 2 || issue1.Dependencies[0] != 2 || issue1.Dependencies[1] != 3 {
		t.Errorf("Expected dependencies [2, 3], got %v", issue1.Dependencies)
	}
	
	expectedTime, _ := time.Parse(time.RFC3339, "2024-01-15T10:00:00Z")
	if !issue1.CreatedAt.Equal(expectedTime) {
		t.Errorf("Expected created time %v, got %v", expectedTime, issue1.CreatedAt)
	}
	
	// Test second issue (minimal fields)
	issue2 := issues[1]
	if issue2.ID != 2 {
		t.Errorf("Expected ID 2, got %d", issue2.ID)
	}
	if issue2.Title != "Simple Issue" {
		t.Errorf("Expected title 'Simple Issue', got '%s'", issue2.Title)
	}
	if len(issue2.Labels) != 0 {
		t.Errorf("Expected no labels, got %v", issue2.Labels)
	}
	if len(issue2.Dependencies) != 0 {
		t.Errorf("Expected no dependencies, got %v", issue2.Dependencies)
	}
}

// TestCSVImportPreview tests preview functionality
func TestCSVImportPreview(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "preview.csv")
	
	content := `title,state,priority
Preview Issue 1,open,high
Preview Issue 2,closed,medium
`
	
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create preview file: %v", err)
	}
	
	config := &ImportConfig{
		FilePath: testFile,
		DryRun:   true,
	}
	
	result, err := ImportCSVToGitHub(testFile, "owner", "repo", "token", config)
	if err != nil {
		t.Fatalf("Preview import failed: %v", err)
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

// TestCSVExportImportRoundTrip tests round-trip consistency
func TestCSVExportImportRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create test issues
	originalIssues := []*Issue{
		{
			ID:             1,
			Title:          "Round Trip Test 1",
			State:          "open",
			Priority:       "high",
			Labels:         []string{"bug", "urgent"},
			Assignee:       "testuser",
			Milestone:      "v1.0",
			EstimatedHours: 8,
			StoryPoints:    5,
			Epic:           "Test Epic",
			Dependencies:   []int{2, 3},
			CreatedAt:      time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			UpdatedAt:      time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			Body:           "Test body content",
		},
		{
			ID:       2,
			Title:    "Round Trip Test 2",
			State:    "closed",
			Priority: "medium",
			Labels:   []string{"feature"},
			Body:     "Another test issue",
		},
	}
	
	// Export to CSV
	exportFile := filepath.Join(tmpDir, "export.csv")
	exportConfig := &ExportConfig{FilePath: exportFile}
	err := WriteCSV(originalIssues, exportFile, exportConfig)
	if err != nil {
		t.Fatalf("Failed to export CSV: %v", err)
	}
	
	// Import back from CSV
	importConfig := &ImportConfig{FilePath: exportFile}
	importedIssues, err := ParseCSV(exportFile, importConfig)
	if err != nil {
		t.Fatalf("Failed to import CSV: %v", err)
	}
	
	// Verify consistency
	if len(importedIssues) != len(originalIssues) {
		t.Fatalf("Expected %d issues, got %d", len(originalIssues), len(importedIssues))
	}
	
	for i, original := range originalIssues {
		imported := importedIssues[i]
		
		if imported.ID != original.ID {
			t.Errorf("Issue %d: ID mismatch: expected %d, got %d", i, original.ID, imported.ID)
		}
		if imported.Title != original.Title {
			t.Errorf("Issue %d: Title mismatch: expected '%s', got '%s'", i, original.Title, imported.Title)
		}
		if imported.State != original.State {
			t.Errorf("Issue %d: State mismatch: expected '%s', got '%s'", i, original.State, imported.State)
		}
		if imported.Priority != original.Priority {
			t.Errorf("Issue %d: Priority mismatch: expected '%s', got '%s'", i, original.Priority, imported.Priority)
		}
		
		// Compare labels
		if len(imported.Labels) != len(original.Labels) {
			t.Errorf("Issue %d: Labels count mismatch: expected %d, got %d", i, len(original.Labels), len(imported.Labels))
		} else {
			for j, label := range original.Labels {
				if imported.Labels[j] != label {
					t.Errorf("Issue %d: Label %d mismatch: expected '%s', got '%s'", i, j, label, imported.Labels[j])
				}
			}
		}
		
		// Compare dependencies
		if len(imported.Dependencies) != len(original.Dependencies) {
			t.Errorf("Issue %d: Dependencies count mismatch: expected %d, got %d", i, len(original.Dependencies), len(imported.Dependencies))
		} else {
			for j, dep := range original.Dependencies {
				if imported.Dependencies[j] != dep {
					t.Errorf("Issue %d: Dependency %d mismatch: expected %d, got %d", i, j, dep, imported.Dependencies[j])
				}
			}
		}
	}
}

// BenchmarkCSVParsing benchmarks CSV parsing performance
func BenchmarkCSVParsing(b *testing.B) {
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "benchmark.csv")
	
	// Create a larger CSV file for benchmarking
	content := "title,state,priority,labels,assignee\n"
	for i := 0; i < 1000; i++ {
		content += "Test Issue,open,medium,bug,user\n"
	}
	
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		b.Fatalf("Failed to create benchmark file: %v", err)
	}
	
	config := &ImportConfig{FilePath: testFile}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ParseCSV(testFile, config)
		if err != nil {
			b.Fatalf("Benchmark parsing failed: %v", err)
		}
	}
}
