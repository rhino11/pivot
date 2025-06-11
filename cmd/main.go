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
			// Check if config file exists, if not, setup configuration
			if _, err := os.Stat("config.yml"); os.IsNotExist(err) {
				if _, err := os.Stat("config.yaml"); os.IsNotExist(err) {
					fmt.Println("Setting up Pivot configuration...")
					if err := internal.InitConfig(); err != nil {
						return fmt.Errorf("config setup failed: %w", err)
					}
				}
			}

			// Then initialize the database
			fmt.Println("Initializing local issues database...")
			if err := internal.Init(); err != nil {
				return fmt.Errorf("database init failed: %w", err)
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
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := internal.InitConfig(); err != nil {
				return fmt.Errorf("config setup failed: %w", err)
			}
			return nil
		},
	}

	var syncCmd = &cobra.Command{
		Use:   "sync",
		Short: "Sync issues between upstream and local database",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := internal.Sync(); err != nil {
				return fmt.Errorf("sync failed: %w", err)
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
	rootCmd.AddCommand(versionCmd)

	return rootCmd
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
