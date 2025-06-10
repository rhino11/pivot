package main

import (
	"fmt"
	"os"

	"github.com/rhino11/pivot/internal"
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
