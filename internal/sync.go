package internal

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Owner    string     `yaml:"owner"`
	Repo     string     `yaml:"repo"`
	Token    string     `yaml:"token"`
	Database string     `yaml:"database,omitempty"`
	Sync     SyncConfig `yaml:"sync,omitempty"`
}

type SyncConfig struct {
	IncludeClosed bool `yaml:"include_closed,omitempty"`
	BatchSize     int  `yaml:"batch_size,omitempty"`
}

// GitHubConfig is kept for backward compatibility
type GitHubConfig struct {
	Owner string `yaml:"owner"`
	Repo  string `yaml:"repo"`
	Token string `yaml:"token"`
}

func loadConfig() (*Config, error) {
	data, err := os.ReadFile("config.yml")
	if err != nil {
		// Try config.yaml for backward compatibility
		data, err = os.ReadFile("config.yaml")
		if err != nil {
			return nil, err
		}
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Set defaults
	if cfg.Database == "" {
		cfg.Database = "./pivot.db"
	}
	if cfg.Sync.BatchSize == 0 {
		cfg.Sync.BatchSize = 100
	}

	return &cfg, nil
}

// LoadConfig loads the configuration file and returns the config
// This is a public wrapper around the private loadConfig function
func LoadConfig() (*Config, error) {
	return loadConfig()
}

func Sync() error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	db, err := InitDB()
	if err != nil {
		return err
	}
	defer db.Close()
	issues, err := FetchIssues(cfg.Owner, cfg.Repo, cfg.Token)
	if err != nil {
		return err
	}
	for _, iss := range issues {
		// Convert labels and assignees to comma-separated
		var labels, assignees string
		for i, l := range iss.Labels {
			if i > 0 {
				labels += ","
			}
			labels += l.Name
		}
		for i, a := range iss.Assignees {
			if i > 0 {
				assignees += ","
			}
			assignees += a.Login
		}
		_, err := db.Exec(`
			INSERT OR REPLACE INTO issues (github_id, number, title, body, state, labels, assignees, created_at, updated_at, closed_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			iss.ID, iss.Number, iss.Title, iss.Body, iss.State, labels, assignees, iss.CreatedAt, iss.UpdatedAt, iss.ClosedAt)
		if err != nil {
			fmt.Println("Failed to insert issue:", iss.Number, err)
		}
	}
	return nil
}
