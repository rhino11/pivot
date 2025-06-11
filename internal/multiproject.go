package internal

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

// MultiProjectConfig represents the new multi-project configuration format
type MultiProjectConfig struct {
	Global   GlobalConfig    `yaml:"global"`
	Projects []ProjectConfig `yaml:"projects"`
}

// GlobalConfig contains global settings for all projects
type GlobalConfig struct {
	Database string `yaml:"database,omitempty"`
	Token    string `yaml:"token,omitempty"`
}

// ProjectConfig represents configuration for a single project
type ProjectConfig struct {
	ID       int    `yaml:"-"` // Database ID (not in YAML)
	Owner    string `yaml:"owner"`
	Repo     string `yaml:"repo"`
	Path     string `yaml:"path,omitempty"`     // Local filesystem path
	Token    string `yaml:"token,omitempty"`    // Project-specific token (overrides global)
	Database string `yaml:"database,omitempty"` // Project-specific database (rare)
}

// LoadMultiProjectConfig loads configuration supporting both new multi-project and legacy formats
func LoadMultiProjectConfig() (*MultiProjectConfig, error) {
	data, err := os.ReadFile("config.yml")
	if err != nil {
		// Try config.yaml for backward compatibility
		data, err = os.ReadFile("config.yaml")
		if err != nil {
			return nil, err
		}
	}

	// Try to parse as multi-project config first
	var multiConfig MultiProjectConfig
	if err := yaml.Unmarshal(data, &multiConfig); err == nil &&
		(len(multiConfig.Projects) > 0 ||
			(multiConfig.Global.Database != "" && multiConfig.Global.Token != "") ||
			(strings.Contains(string(data), "global:") || strings.Contains(string(data), "projects:"))) {
		// Successfully parsed as multi-project config
		setDefaults(&multiConfig)
		return &multiConfig, nil
	}

	// Try to parse as legacy single-project config
	var legacyConfig Config
	if err := yaml.Unmarshal(data, &legacyConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config as either multi-project or legacy format: %w", err)
	}

	// Convert legacy config to multi-project format
	currentDir, _ := os.Getwd()

	// Only convert if we have valid legacy config fields
	if legacyConfig.Owner == "" && legacyConfig.Repo == "" {
		return nil, fmt.Errorf("invalid configuration: missing required fields")
	}

	converted := &MultiProjectConfig{
		Global: GlobalConfig{
			Database: legacyConfig.Database,
			Token:    legacyConfig.Token,
		},
		Projects: []ProjectConfig{
			{
				Owner: legacyConfig.Owner,
				Repo:  legacyConfig.Repo,
				Path:  currentDir,
				// Note: Token and Database will be inherited from Global
			},
		},
	}

	setDefaults(converted)
	return converted, nil
}

// setDefaults sets default values for the configuration
func setDefaults(config *MultiProjectConfig) {
	// Set global defaults
	if config.Global.Database == "" {
		config.Global.Database = "~/.pivot/pivot.db"
	}

	// Set project defaults and resolve paths
	for i := range config.Projects {
		project := &config.Projects[i]
		if project.Path == "" {
			project.Path, _ = os.Getwd()
		}
	}
}

// DetectProjectFromGit attempts to detect project configuration from Git repository
func DetectProjectFromGit() (*ProjectConfig, error) {
	// Find .git directory
	gitDir, err := findGitDirectory()
	if err != nil {
		return nil, fmt.Errorf("not in a git repository: %w", err)
	}

	// Read git config
	gitConfigPath := filepath.Join(gitDir, "config")
	gitConfigData, err := os.ReadFile(gitConfigPath) // #nosec G304 - Reading git config from detected git dir
	if err != nil {
		return nil, fmt.Errorf("failed to read git config: %w", err)
	}

	// Parse remote origin URL
	owner, repo, err := parseGitRemoteOrigin(string(gitConfigData))
	if err != nil {
		return nil, fmt.Errorf("failed to parse git remote origin: %w", err)
	}

	// Get project root directory (parent of .git)
	projectPath := filepath.Dir(gitDir)

	return &ProjectConfig{
		Owner: owner,
		Repo:  repo,
		Path:  projectPath,
	}, nil
}

// findGitDirectory finds the .git directory starting from current directory
func findGitDirectory() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		gitDir := filepath.Join(dir, ".git")
		if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
			return gitDir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root directory
			break
		}
		dir = parent
	}

	return "", fmt.Errorf(".git directory not found")
}

// parseGitRemoteOrigin extracts owner and repo from git remote origin URL
func parseGitRemoteOrigin(gitConfig string) (string, string, error) {
	lines := strings.Split(gitConfig, "\n")
	inRemoteOrigin := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "[remote \"origin\"]") {
			inRemoteOrigin = true
			continue
		}

		if strings.HasPrefix(line, "[") && inRemoteOrigin {
			// Entered a new section, no longer in remote origin
			inRemoteOrigin = false
			continue
		}

		if inRemoteOrigin && strings.HasPrefix(line, "url = ") {
			url := strings.TrimPrefix(line, "url = ")
			return parseGitHubURL(url)
		}
	}

	return "", "", fmt.Errorf("remote origin URL not found in git config")
}

// parseGitHubURL extracts owner and repo from GitHub URL
func parseGitHubURL(url string) (string, string, error) {
	// Handle both HTTPS and SSH URLs
	// HTTPS: https://github.com/owner/repo.git
	// SSH: git@github.com:owner/repo.git

	var path string
	if strings.HasPrefix(url, "https://github.com/") {
		path = strings.TrimPrefix(url, "https://github.com/")
	} else if strings.HasPrefix(url, "git@github.com:") {
		path = strings.TrimPrefix(url, "git@github.com:")
	} else {
		return "", "", fmt.Errorf("unsupported git URL format: %s", url)
	}

	// Remove .git suffix if present
	path = strings.TrimSuffix(path, ".git")

	// Split into owner and repo
	parts := strings.Split(path, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid GitHub URL format: %s", url)
	}

	return parts[0], parts[1], nil
}

// ResolveDatabasePath resolves a database path, expanding ~ to home directory
func ResolveDatabasePath(dbPath string) (string, error) {
	if !strings.HasPrefix(dbPath, "~") {
		return dbPath, nil
	}

	usr, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %w", err)
	}

	return filepath.Join(usr.HomeDir, strings.TrimPrefix(dbPath, "~")), nil
}

// ImportConfigFromFile imports configuration from a file
func ImportConfigFromFile(filePath string) (*MultiProjectConfig, error) {
	data, err := os.ReadFile(filePath) // #nosec G304 - User controls config file path
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config MultiProjectConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	setDefaults(&config)
	return &config, nil
}

// SaveMultiProjectConfig saves a multi-project configuration to config.yml
func SaveMultiProjectConfig(config *MultiProjectConfig) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile("config.yml", data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetEffectiveToken returns the effective token for a project (project-specific or global)
func (p *ProjectConfig) GetEffectiveToken(global *GlobalConfig) string {
	if p.Token != "" {
		return p.Token
	}
	return global.Token
}

// GetEffectiveDatabase returns the effective database path for a project
func (p *ProjectConfig) GetEffectiveDatabase(global *GlobalConfig) string {
	if p.Database != "" {
		return p.Database
	}
	return global.Database
}
