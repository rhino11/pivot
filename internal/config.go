package internal

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// InitConfig creates a new config.yml file with interactive prompts
func InitConfig() error {
	// Check if config.yml already exists
	if _, err := os.Stat("config.yml"); err == nil {
		fmt.Print("config.yml already exists. Overwrite? (y/N): ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Configuration setup cancelled.")
			return nil
		}
	}

	config := Config{
		Database: "./pivot.db",
		Sync: SyncConfig{
			IncludeClosed: true,
			BatchSize:     100,
		},
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("🚀 Pivot CLI Configuration Setup")
	fmt.Println("================================")
	fmt.Println()

	// GitHub Owner/Organization
	fmt.Print("Enter GitHub username or organization: ")
	owner, _ := reader.ReadString('\n')
	config.Owner = strings.TrimSpace(owner)
	if config.Owner == "" {
		return fmt.Errorf("GitHub owner/organization is required")
	}

	// Repository Name
	fmt.Print("Enter repository name: ")
	repo, _ := reader.ReadString('\n')
	config.Repo = strings.TrimSpace(repo)
	if config.Repo == "" {
		return fmt.Errorf("repository name is required")
	}

	// GitHub Token
	fmt.Println()
	fmt.Println("GitHub Personal Access Token:")
	fmt.Println("  • Go to: https://github.com/settings/tokens")
	fmt.Println("  • Click 'Generate new token (classic)'")
	fmt.Println("  • Required scopes: 'repo' (for private repos) or 'public_repo' (for public repos)")
	fmt.Print("Enter your GitHub token: ")
	token, _ := reader.ReadString('\n')
	config.Token = strings.TrimSpace(token)
	if config.Token == "" {
		return fmt.Errorf("GitHub token is required")
	}

	// Database path (optional)
	fmt.Printf("Database file path (default: %s): ", config.Database)
	dbPath, _ := reader.ReadString('\n')
	dbPath = strings.TrimSpace(dbPath)
	if dbPath != "" {
		config.Database = dbPath
	}

	// Sync options
	fmt.Println()
	fmt.Println("Sync Options:")

	// Include closed issues
	fmt.Printf("Include closed issues? (default: %t) [y/N]: ", config.Sync.IncludeClosed)
	closedResponse, _ := reader.ReadString('\n')
	closedResponse = strings.TrimSpace(strings.ToLower(closedResponse))
	if closedResponse == "n" || closedResponse == "no" {
		config.Sync.IncludeClosed = false
	} else if closedResponse == "y" || closedResponse == "yes" {
		config.Sync.IncludeClosed = true
	}

	// Batch size
	fmt.Printf("Batch size for API requests (default: %d): ", config.Sync.BatchSize)
	batchResponse, _ := reader.ReadString('\n')
	batchResponse = strings.TrimSpace(batchResponse)
	if batchResponse != "" {
		if batchSize, err := strconv.Atoi(batchResponse); err == nil && batchSize > 0 {
			config.Sync.BatchSize = batchSize
		}
	}

	// Add comments to the config file
	configContent := fmt.Sprintf(`# GitHub repository details
owner: %s
repo: %s

# GitHub Personal Access Token
# Required scopes: repo (for private repos) or public_repo (for public repos)
token: %s

# Optional: Database file path (default: ./pivot.db)
database: %s

# Optional: Sync options
sync:
  include_closed: %t    # Include closed issues (default: true)
  batch_size: %d        # Number of issues to fetch per request (default: 100)
`,
		config.Owner,
		config.Repo,
		config.Token,
		config.Database,
		config.Sync.IncludeClosed,
		config.Sync.BatchSize,
	)

	if err := os.WriteFile("config.yml", []byte(configContent), 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Println()
	fmt.Println("✓ Configuration saved to config.yml")
	fmt.Println("✓ Configuration setup complete!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Run 'pivot init' to initialize the database")
	fmt.Println("  2. Run 'pivot sync' to sync your GitHub issues")

	return nil
}
