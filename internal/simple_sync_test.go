package internal

import (
	"testing"
)

func TestSimpleSync(t *testing.T) {
	// Test that we can reference GitHubConfig
	var cfg GitHubConfig
	cfg.Owner = "test"
	cfg.Repo = "test"
	cfg.Token = "test"

	if cfg.Owner != "test" {
		t.Errorf("Expected owner 'test', got '%s'", cfg.Owner)
	}
}
