package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/rhino11/pivot/internal"
	"github.com/rhino11/pivot/internal/csv"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func NewRootCommand() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   "pivot",
		Short: "GitHub Issues Management CLI",
		Long:  `Pivot is a CLI tool for managing GitHub issues locally with offline sync capabilities.`,
	}

	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize configuration and local issues database",
		Long:  `Initialize Pivot by creating a configuration file and setting up the local database.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check for import flag
			importFile, _ := cmd.Flags().GetString("import")
			multiProject, _ := cmd.Flags().GetBool("multi-project")

			if importFile != "" {
				// Import configuration from file
				fmt.Printf("📥 Importing configuration from: %s\n", importFile)
				if err := internal.ImportConfigFile(importFile); err != nil {
					return fmt.Errorf("config import failed: %w", err)
				}
			} else {
				// Check if config file exists, if not, setup configuration
				if _, err := os.Stat("config.yml"); os.IsNotExist(err) {
					if _, err := os.Stat("config.yaml"); os.IsNotExist(err) {
						fmt.Println("Setting up Pivot configuration...")

						if multiProject {
							// Use new multi-project setup
							if err := internal.InitMultiProjectConfig(); err != nil {
								return fmt.Errorf("config setup failed: %w", err)
							}
						} else {
							// Use legacy single-project setup
							if err := internal.InitConfig(); err != nil {
								return fmt.Errorf("config setup failed: %w", err)
							}
						}
					}
				}
			}

			// Initialize the database
			fmt.Println("Initializing local issues database...")

			// Try to load as multi-project config first
			if _, err := internal.LoadMultiProjectConfig(); err == nil {
				// Multi-project database initialization
				if err := internal.InitMultiProjectDatabase(); err != nil {
					return fmt.Errorf("multi-project database init failed: %w", err)
				}
			} else {
				// Legacy single-project database initialization
				if err := internal.Init(); err != nil {
					return fmt.Errorf("database init failed: %w", err)
				}
			}

			fmt.Println("✓ Initialized local issues database.")

			fmt.Println()
			fmt.Println("🎉 Pivot is ready to use!")
			fmt.Println("Run 'pivot sync' to fetch your GitHub issues.")
			return nil
		},
	}

	var configCmd = &cobra.Command{
		Use:   "config",
		Short: "Configure Pivot settings",
		Long:  `Set up or modify Pivot configuration interactively.`,
	}

	var configSetupCmd = &cobra.Command{
		Use:   "setup",
		Short: "Interactive configuration setup",
		Long:  `Set up Pivot configuration interactively with multi-project support.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			multiProject, _ := cmd.Flags().GetBool("multi-project")

			if multiProject {
				if err := internal.InitMultiProjectConfig(); err != nil {
					return fmt.Errorf("multi-project config setup failed: %w", err)
				}
			} else {
				if err := internal.InitConfig(); err != nil {
					return fmt.Errorf("config setup failed: %w", err)
				}
			}
			return nil
		},
	}

	var configShowCmd = &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Long:  `Display the current Pivot configuration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Try to load as multi-project config first
			if err := internal.ShowMultiProjectConfig(); err != nil {
				// Fall back to legacy config display
				config, err := internal.LoadConfig()
				if err != nil {
					return fmt.Errorf("failed to load config: %w", err)
				}

				fmt.Println("📋 Current Configuration (Legacy Format)")
				fmt.Println("========================================")
				fmt.Println()
				fmt.Printf("Owner: %s\n", config.Owner)
				fmt.Printf("Repository: %s\n", config.Repo)
				fmt.Printf("Database: %s\n", config.Database)
				if config.Token != "" {
					fmt.Printf("Token: %s***\n", config.Token[:8])
				} else {
					fmt.Println("Token: (not set)")
				}
			}
			return nil
		},
	}

	var configAddProjectCmd = &cobra.Command{
		Use:   "add-project",
		Short: "Add a new project to multi-project configuration",
		Long:  `Add a new GitHub project to your multi-project configuration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := internal.AddProject(); err != nil {
				return fmt.Errorf("failed to add project: %w", err)
			}
			return nil
		},
	}

	var configImportCmd = &cobra.Command{
		Use:   "import <file>",
		Short: "Import configuration from file",
		Long:  `Import Pivot configuration from a YAML file.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]
			if err := internal.ImportConfigFile(filePath); err != nil {
				return fmt.Errorf("config import failed: %w", err)
			}
			return nil
		},
	}

	var syncCmd = &cobra.Command{
		Use:   "sync",
		Short: "Sync issues between upstream and local database",
		RunE: func(cmd *cobra.Command, args []string) error {
			project, _ := cmd.Flags().GetString("project")

			// Try to load multi-project config first
			if _, err := internal.LoadMultiProjectConfig(); err == nil {
				if err := internal.SyncMultiProject(project); err != nil {
					return fmt.Errorf("multi-project sync failed: %w", err)
				}
			} else {
				// Check if it's a config file error (file exists but invalid)
				if _, statErr := os.Stat("config.yml"); statErr == nil {
					return fmt.Errorf("sync failed: %w", err)
				}
				if _, statErr := os.Stat("config.yaml"); statErr == nil {
					return fmt.Errorf("sync failed: %w", err)
				}

				// Fall back to legacy single-project sync
				if err := internal.Sync(); err != nil {
					return fmt.Errorf("sync failed: %w", err)
				}
			}

			fmt.Println("✓ Sync complete.")
			return nil
		},
	}

	// CSV Import/Export commands
	var importCmd = &cobra.Command{
		Use:   "import",
		Short: "Import data from external sources",
		Long:  `Import issues and other data from CSV files or other external sources.`,
	}

	var csvImportCmd = &cobra.Command{
		Use:   "csv <file>",
		Short: "Import issues from CSV file",
		Long: `Import GitHub issues from a CSV file. The CSV should contain columns like:
title, state, priority, labels, assignee, milestone, body, etc.

Examples:
  pivot import csv backlog.csv
  pivot import csv --preview backlog.csv
  pivot import csv --dry-run --repository myorg/myrepo backlog.csv`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]

			// Get flags
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			preview, _ := cmd.Flags().GetBool("preview")
			repository, _ := cmd.Flags().GetString("repository")
			skipDuplicates, _ := cmd.Flags().GetBool("skip-duplicates")

			// Validate CSV file exists
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				return fmt.Errorf("CSV file not found: %s", filePath)
			}

			// Validate CSV format
			fmt.Println("📋 Validating CSV format...")
			if err := csv.ValidateCSV(filePath); err != nil {
				return fmt.Errorf("CSV validation failed: %w", err)
			}
			fmt.Println("✓ CSV format is valid")

			// Parse CSV
			fmt.Println("📊 Parsing CSV data...")
			config := &csv.ImportConfig{
				FilePath:       filePath,
				Repository:     repository,
				DryRun:         dryRun || preview,
				SkipDuplicates: skipDuplicates,
			}

			issues, err := csv.ParseCSV(filePath, config)
			if err != nil {
				return fmt.Errorf("CSV parsing failed: %w", err)
			}

			cmd.Printf("✓ Parsed %d issues from CSV\n", len(issues))

			// Preview mode - just show the data
			if preview {
				cmd.Println("\n📋 Import Preview:")
				cmd.Println("=================")
				for i, issue := range issues {
					if i >= 5 { // Show only first 5 issues in preview
						cmd.Printf("... and %d more issues\n", len(issues)-5)
						break
					}
					cmd.Printf("%d. %s [%s] - %s\n",
						i+1, issue.Title, issue.State, strings.Join(issue.Labels, ", "))
				}
				cmd.Println("\nRun without --preview to perform the actual import.")
				return nil
			}

			// Dry run mode
			if dryRun {
				cmd.Println("\n🧪 Dry Run Mode - No issues will be created")
				cmd.Println("==========================================")
				for _, issue := range issues {
					cmd.Printf("Would create: %s [%s]\n", issue.Title, issue.State)
				}
				cmd.Printf("\nTotal: %d issues would be created\n", len(issues))
				return nil
			}

			// Actual import to GitHub
			if repository == "" {
				return fmt.Errorf("repository flag is required for import (use --repository owner/repo)")
			}

			// Parse repository owner/repo
			repoParts := strings.Split(repository, "/")
			if len(repoParts) != 2 {
				return fmt.Errorf("repository must be in format 'owner/repo', got: %s", repository)
			}
			owner, repoName := repoParts[0], repoParts[1]

			cmd.Println("\n🚀 Starting import to GitHub...")

			// Load configuration to get GitHub token
			cfg, err := internal.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w (run 'pivot init' to set up config)", err)
			}

			result, err := csv.ImportCSVToGitHub(filePath, owner, repoName, cfg.Token, config)
			if err != nil {
				return fmt.Errorf("GitHub import failed: %w", err)
			}

			cmd.Printf("✅ Import complete!\n")
			cmd.Printf("   Total issues: %d\n", result.Total)
			cmd.Printf("   Created: %d\n", result.Created)
			cmd.Printf("   Skipped: %d\n", result.Skipped)
			if len(result.Errors) > 0 {
				cmd.Printf("   Errors: %d\n", len(result.Errors))
				for _, err := range result.Errors {
					cmd.Printf("     - %s\n", err)
				}
			}

			return nil
		},
	}

	var exportCmd = &cobra.Command{
		Use:   "export",
		Short: "Export data to external formats",
		Long:  `Export issues and other data to CSV files or other external formats.`,
	}

	var csvExportCmd = &cobra.Command{
		Use:   "csv [output-file]",
		Short: "Export issues to CSV file",
		Long: `Export GitHub issues to a CSV file.

Examples:
  pivot export csv
  pivot export csv --output issues.csv
  pivot export csv --fields title,state,labels --filter "state:open"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get flags
			outputFile, _ := cmd.Flags().GetString("output")
			fields, _ := cmd.Flags().GetStringSlice("fields")
			filter, _ := cmd.Flags().GetString("filter")
			repository, _ := cmd.Flags().GetString("repository")

			// Default output file
			if outputFile == "" {
				if len(args) > 0 {
					outputFile = args[0]
				} else {
					outputFile = "issues.csv"
				}
			}

			// Make sure output file has .csv extension
			if !strings.HasSuffix(outputFile, ".csv") {
				outputFile += ".csv"
			}

			cmd.Printf("📤 Exporting issues to: %s\n", outputFile)

			// TODO: Get issues from local database or GitHub API
			// For now, create some sample data
			sampleIssues := []*csv.Issue{
				{
					ID:       1,
					Title:    "Sample Issue 1",
					State:    "open",
					Priority: "high",
					Labels:   []string{"bug", "urgent"},
					Body:     "This is a sample issue for testing CSV export",
				},
				{
					ID:       2,
					Title:    "Sample Issue 2",
					State:    "closed",
					Priority: "medium",
					Labels:   []string{"feature"},
					Body:     "Another sample issue",
				},
			}

			config := &csv.ExportConfig{
				FilePath:   outputFile,
				Repository: repository,
				Fields:     fields,
				Filter:     filter,
			}

			if err := csv.WriteCSV(sampleIssues, outputFile, config); err != nil {
				return fmt.Errorf("CSV export failed: %w", err)
			}

			cmd.Printf("✓ Exported %d issues to %s\n", len(sampleIssues), outputFile)
			return nil
		},
	}

	// Add flags to commands
	initCmd.Flags().String("import", "", "Import configuration from file")
	initCmd.Flags().Bool("multi-project", false, "Use multi-project configuration setup")

	configSetupCmd.Flags().Bool("multi-project", false, "Use multi-project configuration setup")

	syncCmd.Flags().String("project", "", "Sync specific project (format: owner/repo)")

	// Add flags to CSV import command
	csvImportCmd.Flags().Bool("preview", false, "Preview the import without creating issues")
	csvImportCmd.Flags().Bool("dry-run", false, "Show what would be imported without making changes")
	csvImportCmd.Flags().String("repository", "", "Target GitHub repository (e.g., owner/repo)")
	csvImportCmd.Flags().Bool("skip-duplicates", false, "Skip issues that appear to be duplicates")

	// Add flags to CSV export command
	csvExportCmd.Flags().StringP("output", "o", "", "Output CSV file path")
	csvExportCmd.Flags().StringSlice("fields", []string{}, "Specific fields to export (comma-separated)")
	csvExportCmd.Flags().String("filter", "", "Filter expression for issues to export")
	csvExportCmd.Flags().String("repository", "", "Source GitHub repository (e.g., owner/repo)")

	// Build command hierarchy
	configCmd.AddCommand(configSetupCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configAddProjectCmd)
	configCmd.AddCommand(configImportCmd)

	importCmd.AddCommand(csvImportCmd)
	exportCmd.AddCommand(csvExportCmd)

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Printf("pivot version %s\n", version)
			cmd.Printf("commit: %s\n", commit)
			cmd.Printf("built: %s\n", date)
		},
	}

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(createCSVHelpCommand())
	rootCmd.AddCommand(versionCmd)

	return rootCmd
}

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

// Run executes the main application logic and returns an exit code
func Run() int {
	rootCmd := NewRootCommand()

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	return 0
}

func main() {
	os.Exit(Run())
}
