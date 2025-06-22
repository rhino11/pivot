package internal

import (
	"database/sql"
	"fmt"
	"time"
)

// SyncState represents the sync state of an issue
type SyncState string

const (
	// Issue created locally, not yet pushed to GitHub
	SyncStateLocalOnly SyncState = "LOCAL_ONLY"

	// Local issue queued for creation on GitHub
	SyncStatePendingPush SyncState = "PENDING_PUSH"

	// Failed to create issue on GitHub
	SyncStatePushFailed SyncState = "PUSH_FAILED"

	// Issue exists in both places, no local changes
	SyncStateSynced SyncState = "SYNCED"

	// Issue exists in both, but modified locally
	SyncStateLocalModified SyncState = "LOCAL_MODIFIED"

	// Local changes queued for sync to GitHub
	SyncStatePendingSync SyncState = "PENDING_SYNC"

	// Failed to sync local changes to GitHub
	SyncStateSyncFailed SyncState = "SYNC_FAILED"

	// Both local and remote changes detected
	SyncStateConflicted SyncState = "CONFLICTED"

	// Unrecoverable error state
	SyncStateError SyncState = "ERROR"
)

// IssueSyncState represents the sync state record for an issue
type IssueSyncState struct {
	ID                 int64      `json:"id"`
	IssueLocalID       int64      `json:"issue_local_id"`
	GitHubID           *int64     `json:"github_id"`
	SyncState          SyncState  `json:"sync_state"`
	LastLocalModified  *time.Time `json:"last_local_modified"`
	LastRemoteModified *time.Time `json:"last_remote_modified"`
	LastSyncAttempt    *time.Time `json:"last_sync_attempt"`
	SyncError          *string    `json:"sync_error"`
	RetryCount         int        `json:"retry_count"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// CreateSyncStateTable creates the issue_sync_state table
func CreateSyncStateTable(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS issue_sync_state (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		issue_local_id INTEGER NOT NULL,
		github_id INTEGER,
		sync_state TEXT NOT NULL CHECK(sync_state IN (
			'LOCAL_ONLY', 'PENDING_PUSH', 'PUSH_FAILED', 
			'SYNCED', 'LOCAL_MODIFIED', 'PENDING_SYNC', 
			'SYNC_FAILED', 'CONFLICTED', 'ERROR'
		)),
		last_local_modified TEXT,
		last_remote_modified TEXT,
		last_sync_attempt TEXT,
		sync_error TEXT,
		retry_count INTEGER DEFAULT 0,
		created_at TEXT DEFAULT CURRENT_TIMESTAMP,
		updated_at TEXT DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(issue_local_id)
	);`

	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create issue_sync_state table: %w", err)
	}

	// Create index on github_id for faster lookups
	indexSchema := `
	CREATE INDEX IF NOT EXISTS idx_sync_state_github_id 
	ON issue_sync_state(github_id) WHERE github_id IS NOT NULL;`

	_, err = db.Exec(indexSchema)
	if err != nil {
		return fmt.Errorf("failed to create sync state index: %w", err)
	}

	return nil
}

// AddSyncColumnsToIssues adds sync-related columns to the issues table
func AddSyncColumnsToIssues(db *sql.DB) error {
	// Check if columns already exist
	hasLocalModified, err := hasColumn(db, "issues", "local_modified_at")
	if err != nil {
		return fmt.Errorf("failed to check for local_modified_at column: %w", err)
	}

	hasSyncHash, err := hasColumn(db, "issues", "sync_hash")
	if err != nil {
		return fmt.Errorf("failed to check for sync_hash column: %w", err)
	}

	// Add columns if they don't exist
	if !hasLocalModified {
		_, err = db.Exec("ALTER TABLE issues ADD COLUMN local_modified_at TEXT")
		if err != nil {
			return fmt.Errorf("failed to add local_modified_at column: %w", err)
		}
	}

	if !hasSyncHash {
		_, err = db.Exec("ALTER TABLE issues ADD COLUMN sync_hash TEXT")
		if err != nil {
			return fmt.Errorf("failed to add sync_hash column: %w", err)
		}
	}

	return nil
}

// InitSyncStateSchema initializes the sync state schema
func InitSyncStateSchema(db *sql.DB) error {
	// Create sync state table
	if err := CreateSyncStateTable(db); err != nil {
		return err
	}

	// Add sync columns to issues table
	if err := AddSyncColumnsToIssues(db); err != nil {
		return err
	}

	return nil
}

// CreateSyncState creates a new sync state record for an issue
func CreateSyncState(db *sql.DB, issueLocalID int64, state SyncState, githubID *int64) error {
	now := time.Now().Format(time.RFC3339)

	query := `
		INSERT INTO issue_sync_state 
		(issue_local_id, github_id, sync_state, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err := db.Exec(query, issueLocalID, githubID, string(state), now, now)
	if err != nil {
		return fmt.Errorf("failed to create sync state: %w", err)
	}

	return nil
}

// UpdateSyncState updates the sync state of an issue
func UpdateSyncState(db *sql.DB, issueLocalID int64, state SyncState, githubID *int64, syncError *string) error {
	now := time.Now().Format(time.RFC3339)

	query := `
		UPDATE issue_sync_state 
		SET sync_state = ?, github_id = ?, sync_error = ?, updated_at = ?,
		    last_sync_attempt = CASE WHEN ? IN ('PENDING_PUSH', 'PENDING_SYNC') THEN ? ELSE last_sync_attempt END,
		    retry_count = CASE WHEN ? IN ('PUSH_FAILED', 'SYNC_FAILED') THEN retry_count + 1 ELSE retry_count END
		WHERE issue_local_id = ?
	`

	_, err := db.Exec(query, string(state), githubID, syncError, now, string(state), now, string(state), issueLocalID)
	if err != nil {
		return fmt.Errorf("failed to update sync state: %w", err)
	}

	return nil
}

// GetSyncState retrieves the sync state for an issue
func GetSyncState(db *sql.DB, issueLocalID int64) (*IssueSyncState, error) {
	query := `
		SELECT id, issue_local_id, github_id, sync_state, last_local_modified,
		       last_remote_modified, last_sync_attempt, sync_error, retry_count,
		       created_at, updated_at
		FROM issue_sync_state
		WHERE issue_local_id = ?
	`

	var state IssueSyncState
	var githubID sql.NullInt64
	var lastLocalModified, lastRemoteModified, lastSyncAttempt sql.NullString
	var syncError sql.NullString
	var createdAt, updatedAt string

	err := db.QueryRow(query, issueLocalID).Scan(
		&state.ID, &state.IssueLocalID, &githubID, &state.SyncState,
		&lastLocalModified, &lastRemoteModified, &lastSyncAttempt,
		&syncError, &state.RetryCount, &createdAt, &updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No sync state record found
		}
		return nil, fmt.Errorf("failed to get sync state: %w", err)
	}

	// Convert nullable fields
	if githubID.Valid {
		state.GitHubID = &githubID.Int64
	}

	if lastLocalModified.Valid {
		if t, err := time.Parse(time.RFC3339, lastLocalModified.String); err == nil {
			state.LastLocalModified = &t
		}
	}

	if lastRemoteModified.Valid {
		if t, err := time.Parse(time.RFC3339, lastRemoteModified.String); err == nil {
			state.LastRemoteModified = &t
		}
	}

	if lastSyncAttempt.Valid {
		if t, err := time.Parse(time.RFC3339, lastSyncAttempt.String); err == nil {
			state.LastSyncAttempt = &t
		}
	}

	if syncError.Valid {
		state.SyncError = &syncError.String
	}

	// Parse timestamps
	if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
		state.CreatedAt = t
	}

	if t, err := time.Parse(time.RFC3339, updatedAt); err == nil {
		state.UpdatedAt = t
	}

	return &state, nil
}

// GetSyncStatesByState retrieves all issues in a specific sync state
func GetSyncStatesByState(db *sql.DB, state SyncState) ([]IssueSyncState, error) {
	query := `
		SELECT id, issue_local_id, github_id, sync_state, last_local_modified,
		       last_remote_modified, last_sync_attempt, sync_error, retry_count,
		       created_at, updated_at
		FROM issue_sync_state
		WHERE sync_state = ?
		ORDER BY updated_at DESC
	`

	rows, err := db.Query(query, string(state))
	if err != nil {
		return nil, fmt.Errorf("failed to query sync states: %w", err)
	}
	defer rows.Close()

	var states []IssueSyncState
	for rows.Next() {
		var s IssueSyncState
		var githubID sql.NullInt64
		var lastLocalModified, lastRemoteModified, lastSyncAttempt sql.NullString
		var syncError sql.NullString
		var createdAt, updatedAt string

		err := rows.Scan(
			&s.ID, &s.IssueLocalID, &githubID, &s.SyncState,
			&lastLocalModified, &lastRemoteModified, &lastSyncAttempt,
			&syncError, &s.RetryCount, &createdAt, &updatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan sync state: %w", err)
		}

		// Convert nullable fields (same logic as GetSyncState)
		if githubID.Valid {
			s.GitHubID = &githubID.Int64
		}

		if lastLocalModified.Valid {
			if t, err := time.Parse(time.RFC3339, lastLocalModified.String); err == nil {
				s.LastLocalModified = &t
			}
		}

		if lastRemoteModified.Valid {
			if t, err := time.Parse(time.RFC3339, lastRemoteModified.String); err == nil {
				s.LastRemoteModified = &t
			}
		}

		if lastSyncAttempt.Valid {
			if t, err := time.Parse(time.RFC3339, lastSyncAttempt.String); err == nil {
				s.LastSyncAttempt = &t
			}
		}

		if syncError.Valid {
			s.SyncError = &syncError.String
		}

		// Parse timestamps
		if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
			s.CreatedAt = t
		}

		if t, err := time.Parse(time.RFC3339, updatedAt); err == nil {
			s.UpdatedAt = t
		}

		states = append(states, s)
	}

	return states, nil
}

// GetSyncStateSummary returns a summary of all sync states
func GetSyncStateSummary(db *sql.DB) (map[SyncState]int, error) {
	query := `
		SELECT sync_state, COUNT(*) 
		FROM issue_sync_state 
		GROUP BY sync_state
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query sync state summary: %w", err)
	}
	defer rows.Close()

	summary := make(map[SyncState]int)
	for rows.Next() {
		var state string
		var count int

		if err := rows.Scan(&state, &count); err != nil {
			return nil, fmt.Errorf("failed to scan sync state summary: %w", err)
		}

		summary[SyncState(state)] = count
	}

	return summary, nil
}
