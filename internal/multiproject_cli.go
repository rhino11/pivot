package internal

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strings"
)

// InitMultiProjectConfig creates a new multi-project config with interactive prompts
func InitMultiProjectConfig() error {
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

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("🚀 Pivot CLI Multi-Project Configuration Setup")
	fmt.Println("==============================================")
	fmt.Println()

	config := &MultiProjectConfig{
		Global: GlobalConfig{
			Database: "~/.pivot/pivot.db",
		},
		Projects: []ProjectConfig{},
	}

	// Global configuration
	fmt.Println("📁 Global Configuration")
	fmt.Println("-----------------------")

	// Global GitHub token
	fmt.Println("GitHub Personal Access Token (global default):")
	fmt.Println("  • Go to: https://github.com/settings/tokens")
	fmt.Println("  • Click 'Generate new token (classic)'")
	fmt.Println("  • Required scopes: 'repo' (for private repos) or 'public_repo' (for public repos)")
	fmt.Print("Enter your GitHub token (leave empty to set per-project): ")
	token, _ := reader.ReadString('\n')
	config.Global.Token = strings.TrimSpace(token)

	// Global database path
	fmt.Printf("Central database file path (default: %s): ", config.Global.Database)
	dbPath, _ := reader.ReadString('\n')
	dbPath = strings.TrimSpace(dbPath)
	if dbPath != "" {
		config.Global.Database = dbPath
	}

	fmt.Println()

	// Project configuration
	fmt.Println("📚 Project Configuration")
	fmt.Println("------------------------")
	fmt.Println("You can set up multiple GitHub projects to manage with Pivot.")
	fmt.Println()

	// Auto-detect current project from git
	fmt.Print("Auto-detect current project from git repository? (Y/n): ")
	autoDetect, _ := reader.ReadString('\n')
	autoDetect = strings.TrimSpace(strings.ToLower(autoDetect))

	if autoDetect == "" || autoDetect == "y" || autoDetect == "yes" {
		if project, err := DetectProjectFromGit(); err == nil {
			fmt.Printf("✓ Detected project: %s/%s at %s\n", project.Owner, project.Repo, project.Path)

			// Set project-specific token if global token is empty
			if config.Global.Token == "" {
				fmt.Printf("Enter GitHub token for %s/%s: ", project.Owner, project.Repo)
				projectToken, _ := reader.ReadString('\n')
				project.Token = strings.TrimSpace(projectToken)
			}

			config.Projects = append(config.Projects, *project)
		} else {
			fmt.Printf("⚠ Could not auto-detect project: %v\n", err)
			fmt.Println("You can manually add projects below.")
		}
	}

	// Allow adding more projects
	for {
		fmt.Println()
		fmt.Print("Add another project? (y/N): ")
		addMore, _ := reader.ReadString('\n')
		addMore = strings.TrimSpace(strings.ToLower(addMore))

		if addMore != "y" && addMore != "yes" {
			break
		}

		project := ProjectConfig{}

		fmt.Print("GitHub username or organization: ")
		owner, _ := reader.ReadString('\n')
		project.Owner = strings.TrimSpace(owner)
		if project.Owner == "" {
			fmt.Println("Owner is required, skipping project.")
			continue
		}

		fmt.Print("Repository name: ")
		repo, _ := reader.ReadString('\n')
		project.Repo = strings.TrimSpace(repo)
		if project.Repo == "" {
			fmt.Println("Repository name is required, skipping project.")
			continue
		}

		fmt.Print("Local project path (leave empty for current directory): ")
		path, _ := reader.ReadString('\n')
		path = strings.TrimSpace(path)
		if path == "" {
			path, _ = os.Getwd()
		}
		project.Path = path

		// Project-specific token if global token is empty
		if config.Global.Token == "" {
			fmt.Printf("GitHub token for %s/%s: ", project.Owner, project.Repo)
			projectToken, _ := reader.ReadString('\n')
			project.Token = strings.TrimSpace(projectToken)
		}

		config.Projects = append(config.Projects, project)
	}

	// Save configuration
	if err := SaveMultiProjectConfig(config); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Println()
	fmt.Println("✓ Configuration saved to config.yml")
	fmt.Println("✓ Multi-project configuration setup complete!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Run 'pivot init' to initialize the database")
	fmt.Println("  2. Run 'pivot sync' to sync your GitHub issues")

	return nil
}

// InitMultiProjectDatabase initializes the multi-project database
func InitMultiProjectDatabase() error {
	// Load configuration
	config, err := LoadMultiProjectConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Resolve and initialize central database
	dbPath, err := ResolveDatabasePath(config.Global.Database)
	if err != nil {
		return fmt.Errorf("failed to resolve database path: %w", err)
	}

	fmt.Printf("Initializing database at: %s\n", dbPath)

	db, err := InitMultiProjectDBFromPath(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	// Create project entries in database
	for _, project := range config.Projects {
		projectID, err := CreateProject(db, &project)
		if err != nil {
			return fmt.Errorf("failed to create project %s/%s: %w", project.Owner, project.Repo, err)
		}
		fmt.Printf("✓ Registered project: %s/%s (ID: %d)\n", project.Owner, project.Repo, projectID)
	}

	return nil
}

// SyncMultiProject syncs all projects or a specific project
func SyncMultiProject(projectFilter string) error {
	// Load configuration
	config, err := LoadMultiProjectConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Validate configuration has projects
	if len(config.Projects) == 0 {
		return fmt.Errorf("no projects configured in multi-project configuration")
	}

	// Open central database
	dbPath, err := ResolveDatabasePath(config.Global.Database)
	if err != nil {
		return fmt.Errorf("failed to resolve database path: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Determine which projects to sync
	var projectsToSync []ProjectConfig
	if projectFilter != "" {
		// Parse project filter (owner/repo format)
		parts := strings.Split(projectFilter, "/")
		if len(parts) != 2 {
			return fmt.Errorf("project filter must be in format 'owner/repo', got: %s", projectFilter)
		}

		// Find the specific project
		found := false
		for _, project := range config.Projects {
			if project.Owner == parts[0] && project.Repo == parts[1] {
				projectsToSync = append(projectsToSync, project)
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("project %s not found in configuration", projectFilter)
		}
	} else {
		// Sync all projects
		projectsToSync = config.Projects
	}

	// Sync each project
	for _, project := range projectsToSync {
		fmt.Printf("🔄 Syncing %s/%s...\n", project.Owner, project.Repo)

		if err := syncProject(db, &config.Global, &project); err != nil {
			fmt.Printf("❌ Failed to sync %s/%s: %v\n", project.Owner, project.Repo, err)
			continue
		}

		fmt.Printf("✓ Synced %s/%s\n", project.Owner, project.Repo)
	}

	return nil
}

// syncProject syncs a single project
func syncProject(db *sql.DB, global *GlobalConfig, project *ProjectConfig) error {
	// Get effective token for this project
	token := project.GetEffectiveToken(global)
	if token == "" {
		return fmt.Errorf("no GitHub token configured for project %s/%s", project.Owner, project.Repo)
	}

	// Validate GitHub credentials before attempting sync
	if err := EnsureGitHubCredentials(project.Owner, project.Repo, token); err != nil {
		return fmt.Errorf("GitHub credential validation failed for %s/%s: %w", project.Owner, project.Repo, err)
	}

	// Ensure project exists in database
	projectID, err := CreateProject(db, project)
	if err != nil {
		return fmt.Errorf("failed to ensure project in database: %w", err)
	}

	// Fetch issues from GitHub
	issues, err := FetchIssues(project.Owner, project.Repo, token)
	if err != nil {
		return fmt.Errorf("failed to fetch issues from GitHub: %w", err)
	}

	// Save issues to database
	for _, issue := range issues {
		dbIssue := ConvertIssueToDBIssue(&issue)
		if err := SaveIssue(db, projectID, dbIssue); err != nil {
			return fmt.Errorf("failed to save issue %d: %w", issue.ID, err)
		}
	}

	fmt.Printf("  Saved %d issues\n", len(issues))
	return nil
}

// ShowMultiProjectConfig displays the current multi-project configuration
func ShowMultiProjectConfig() error {
	config, err := LoadMultiProjectConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("📋 Current Multi-Project Configuration")
	fmt.Println("=====================================")
	fmt.Println()

	fmt.Println("🌍 Global Settings:")
	fmt.Printf("  Database: %s\n", config.Global.Database)
	if config.Global.Token != "" {
		fmt.Printf("  Token: %s***\n", config.Global.Token[:8])
	} else {
		fmt.Println("  Token: (not set)")
	}
	fmt.Println()

	fmt.Printf("📚 Projects (%d configured):\n", len(config.Projects))
	for i, project := range config.Projects {
		fmt.Printf("  %d. %s/%s\n", i+1, project.Owner, project.Repo)
		if project.Path != "" {
			fmt.Printf("     Path: %s\n", project.Path)
		}
		if project.Token != "" {
			fmt.Printf("     Token: %s*** (project-specific)\n", project.Token[:8])
		}
		fmt.Println()
	}

	return nil
}

// AddProject adds a new project to the configuration
func AddProject() error {
	config, err := LoadMultiProjectConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	reader := bufio.NewReader(os.Stdin)
	project := ProjectConfig{}

	fmt.Println("➕ Add New Project")
	fmt.Println("==================")
	fmt.Println()

	// Auto-detect from git first
	fmt.Print("Auto-detect project from current git repository? (Y/n): ")
	autoDetect, _ := reader.ReadString('\n')
	autoDetect = strings.TrimSpace(strings.ToLower(autoDetect))

	if autoDetect == "" || autoDetect == "y" || autoDetect == "yes" {
		if detected, err := DetectProjectFromGit(); err == nil {
			fmt.Printf("✓ Detected project: %s/%s at %s\n", detected.Owner, detected.Repo, detected.Path)
			fmt.Print("Use detected project? (Y/n): ")
			useDetected, _ := reader.ReadString('\n')
			useDetected = strings.TrimSpace(strings.ToLower(useDetected))

			if useDetected == "" || useDetected == "y" || useDetected == "yes" {
				project = *detected
			}
		}
	}

	// Manual entry if not auto-detected
	if project.Owner == "" {
		fmt.Print("GitHub username or organization: ")
		owner, _ := reader.ReadString('\n')
		project.Owner = strings.TrimSpace(owner)
		if project.Owner == "" {
			return fmt.Errorf("owner is required")
		}

		fmt.Print("Repository name: ")
		repo, _ := reader.ReadString('\n')
		project.Repo = strings.TrimSpace(repo)
		if project.Repo == "" {
			return fmt.Errorf("repository name is required")
		}

		fmt.Print("Local project path (leave empty for current directory): ")
		path, _ := reader.ReadString('\n')
		path = strings.TrimSpace(path)
		if path == "" {
			path, _ = os.Getwd()
		}
		project.Path = path
	}

	// Project-specific token if global token is empty
	if config.Global.Token == "" {
		fmt.Printf("GitHub token for %s/%s: ", project.Owner, project.Repo)
		projectToken, _ := reader.ReadString('\n')
		project.Token = strings.TrimSpace(projectToken)
	}

	// Check if project already exists
	for _, existing := range config.Projects {
		if existing.Owner == project.Owner && existing.Repo == project.Repo {
			return fmt.Errorf("project %s/%s already exists in configuration", project.Owner, project.Repo)
		}
	}

	// Add to configuration
	config.Projects = append(config.Projects, project)

	// Save configuration
	if err := SaveMultiProjectConfig(config); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("✓ Added project %s/%s to configuration\n", project.Owner, project.Repo)
	fmt.Println("Run 'pivot sync' to sync this project's issues.")

	return nil
}

// ImportConfigFile imports a configuration from a file
func ImportConfigFile(filePath string) error {
	fmt.Printf("📥 Importing configuration from: %s\n", filePath)

	imported, err := ImportConfigFromFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to import config: %w", err)
	}

	// Check if current config exists
	var merge bool
	if _, err := os.Stat("config.yml"); err == nil {
		fmt.Print("Current config.yml exists. Merge with imported config? (y/N): ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		merge = response == "y" || response == "yes"
	}

	if merge {
		// Load current config and merge
		current, err := LoadMultiProjectConfig()
		if err != nil {
			return fmt.Errorf("failed to load current config: %w", err)
		}

		// Merge global settings (imported takes precedence)
		if imported.Global.Token != "" {
			current.Global.Token = imported.Global.Token
		}
		if imported.Global.Database != "" {
			current.Global.Database = imported.Global.Database
		}

		// Merge projects (avoid duplicates)
		for _, importedProject := range imported.Projects {
			found := false
			for i, currentProject := range current.Projects {
				if currentProject.Owner == importedProject.Owner && currentProject.Repo == importedProject.Repo {
					// Update existing project
					current.Projects[i] = importedProject
					found = true
					break
				}
			}
			if !found {
				current.Projects = append(current.Projects, importedProject)
			}
		}

		imported = current
	}

	// Save the final configuration
	if err := SaveMultiProjectConfig(imported); err != nil {
		return fmt.Errorf("failed to save imported config: %w", err)
	}

	fmt.Printf("✓ Imported configuration with %d projects\n", len(imported.Projects))
	for _, project := range imported.Projects {
		fmt.Printf("  - %s/%s\n", project.Owner, project.Repo)
	}

	return nil
}
