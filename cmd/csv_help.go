package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// createCSVHelpCommand creates the CSV help command that provides format guidance
func createCSVHelpCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "csv-format",
		Short: "Show CSV format guidelines and examples",
		Long: `Display comprehensive CSV format guidelines, field descriptions, and examples for importing/exporting GitHub issues.

This command provides detailed information about:
- Required and optional CSV fields  
- Formatting rules and best practices
- Common use cases and examples
- Troubleshooting guidance`,
		RunE: func(cmd *cobra.Command, args []string) error {
			showCSVFormatGuide()
			return nil
		},
	}

	return cmd
}

// showCSVFormatGuide displays the comprehensive CSV format guide
func showCSVFormatGuide() {
	fmt.Println("📊 Pivot CSV Format Guide")
	fmt.Println("=" + strings.Repeat("=", 24))
	fmt.Println()

	// Required Fields
	fmt.Println("🔴 Required Fields:")
	fmt.Println("  • title - Issue title (cannot be empty)")
	fmt.Println()

	// Optional Fields
	fmt.Println("🟡 Optional Fields:")
	fields := []struct {
		name        string
		description string
		example     string
	}{
		{"id", "Issue ID", "123"},
		{"state", "Issue state", "open, closed"},
		{"priority", "Issue priority", "high, medium, low"},
		{"labels", "Comma-separated labels", "bug,urgent,security"},
		{"assignee", "Assigned user", "john.doe"},
		{"milestone", "Milestone name", "v1.0.0"},
		{"body", "Issue description", "Detailed description..."},
		{"estimated_hours", "Estimated work hours", "8"},
		{"story_points", "Story points", "5"},
		{"epic", "Epic name", "User Authentication"},
		{"dependencies", "Comma-separated issue IDs", "45,67,89"},
		{"acceptance_criteria", "Acceptance criteria", "User can login successfully"},
		{"created_at", "Creation timestamp", "2024-01-15T10:00:00Z"},
		{"updated_at", "Last update timestamp", "2024-01-15T10:30:00Z"},
	}

	for _, field := range fields {
		fmt.Printf("  • %-18s - %s\n", field.name, field.description)
		fmt.Printf("    %sExample: %s%s\n", "\033[90m", field.example, "\033[0m")
	}
	fmt.Println()

	// Formatting Rules
	fmt.Println("📝 Formatting Rules:")
	fmt.Println("  1. Header row must contain field names (case-insensitive)")
	fmt.Println("  2. Enclose values with commas/quotes in double quotes")
	fmt.Println("  3. Escape internal quotes by doubling: \"Issue with \"\"quotes\"\"\"")
	fmt.Println("  4. Use comma separation for multi-value fields")
	fmt.Println("  5. Use RFC3339 format for dates: 2024-01-15T10:00:00Z")
	fmt.Println("  6. Leave empty for optional fields: \"\"")
	fmt.Println()

	// Examples
	fmt.Println("📋 Example CSV Files:")
	fmt.Println()

	fmt.Println("🟢 Minimal CSV:")
	fmt.Println("```csv")
	fmt.Println("title,state")
	fmt.Println("\"Fix login bug\",\"open\"")
	fmt.Println("\"Add user dashboard\",\"closed\"")
	fmt.Println("```")
	fmt.Println()

	fmt.Println("🟢 Complete CSV:")
	fmt.Println("```csv")
	fmt.Println("id,title,state,priority,labels,assignee,milestone,estimated_hours,story_points")
	fmt.Println("1,\"Implement authentication\",\"open\",\"high\",\"backend,security\",\"john\",\"v1.0\",16,8")
	fmt.Println("2,\"Design login UI\",\"open\",\"medium\",\"frontend,ui\",\"jane\",\"v1.0\",8,5")
	fmt.Println("```")
	fmt.Println()

	// Commands
	fmt.Println("🚀 Quick Start Commands:")
	fmt.Println("  # Preview import (recommended first step)")
	fmt.Println("  pivot import csv --preview issues.csv")
	fmt.Println()
	fmt.Println("  # Dry-run import (validates without creating)")
	fmt.Println("  pivot import csv --dry-run issues.csv")
	fmt.Println()
	fmt.Println("  # Actual import")
	fmt.Println("  pivot import csv issues.csv")
	fmt.Println()
	fmt.Println("  # Export current issues")
	fmt.Println("  pivot export csv --output my-issues.csv")
	fmt.Println()

	// Common Issues
	fmt.Println("🔧 Common Issues & Solutions:")
	fmt.Println("  • 'CSV validation failed: EOF' → File is empty or has no data rows")
	fmt.Println("  • 'Required column title not found' → Add title column to header")
	fmt.Println("  • 'Column count mismatch' → Ensure all rows have same number of fields")
	fmt.Println("  • 'Failed to parse date' → Use RFC3339 format: 2024-01-15T10:00:00Z")
	fmt.Println()

	// Best Practices
	fmt.Println("💡 Best Practices:")
	fmt.Println("  1. Always use --preview flag first")
	fmt.Println("  2. Test with --dry-run before actual import")
	fmt.Println("  3. Backup existing issues before importing")
	fmt.Println("  4. Start with small test files")
	fmt.Println("  5. Ensure UTF-8 encoding without BOM")
	fmt.Println()

	fmt.Println("📖 For detailed documentation, see: docs/CSV_FORMAT_GUIDE.md")
}
