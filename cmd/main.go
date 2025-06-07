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

func main() {
	var rootCmd = &cobra.Command{
		Use:   "pivot",
		Short: "GitHub Issues Management CLI",
		Long:  `Pivot is a CLI tool for managing GitHub issues locally with offline sync capabilities.`,
	}

	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize the local issues database",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := internal.Init(); err != nil {
				return fmt.Errorf("init failed: %w", err)
			}
			fmt.Println("✓ Initialized local issues database.")
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
			fmt.Printf("pivot %s\n", version)
			fmt.Printf("commit: %s\n", commit)
			fmt.Printf("built: %s\n", date)
		},
	}

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
