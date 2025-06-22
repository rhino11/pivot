package internal

import (
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Test fixture for sync state machine tests
type syncStateTestFixture struct {
	db            *sql.DB
	tempDBPath    string
	testIssueID   int64
	testProjectID int64
}

func setupSyncStateTest(t *testing.T) *syncStateTestFixture {
	// Create temporary database
	tmpDB, err := os.CreateTemp("", "sync-state-test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp database: %v", err)
	}
	tmpDB.Close()

	// Open database connection
	db, err := sql.Open("sqlite3", tmpDB.Name())
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Create test schema
	if err := createTestSchema(db); err != nil {
		t.Fatalf("Failed to create test schema: %v", err)
	}

	// Initialize sync state schema
	if err := InitSyncStateSchema(db); err != nil {
		t.Fatalf("Failed to initialize sync state schema: %v", err)
	}

	// Create test project
	projectID := int64(1)
	_, err = db.Exec(`
		INSERT INTO projects (id, owner, repo, path, created_at, updated_at)
		VALUES (?, 'test', 'test-repo', '.', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, projectID)
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	// Create test issue
	result, err := db.Exec(`
		INSERT INTO issues (project_id, number, title, body, state, created_at, updated_at)
		VALUES (?, 1, 'Test Issue', 'Test body', 'open', ?, ?)
	`, projectID, time.Now().Format(time.RFC3339), time.Now().Format(time.RFC3339))
	if err != nil {
		t.Fatalf("Failed to create test issue: %v", err)
	}

	issueID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get test issue ID: %v", err)
	}

	return &syncStateTestFixture{
		db:            db,
		tempDBPath:    tmpDB.Name(),
		testIssueID:   issueID,
		testProjectID: projectID,
	}
}

func teardownSyncStateTest(fixture *syncStateTestFixture) {
	if fixture.db != nil {
		fixture.db.Close()
	}
	if fixture.tempDBPath != "" {
		os.Remove(fixture.tempDBPath)
	}
}

func createTestSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS projects (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		owner TEXT NOT NULL,
		repo TEXT NOT NULL,
		path TEXT,
		created_at TEXT DEFAULT CURRENT_TIMESTAMP,
		updated_at TEXT DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS issues (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		github_id INTEGER,
		project_id INTEGER NOT NULL,
		number INTEGER,
		title TEXT,
		body TEXT,
		state TEXT,
		labels TEXT,
		assignees TEXT,
		created_at TEXT,
		updated_at TEXT,
		closed_at TEXT,
		local_modified_at TEXT,
		sync_hash TEXT,
		FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
	);`

	_, err := db.Exec(schema)
	return err
}

// Test all state transitions from LOCAL_ONLY
func TestLocalOnlyStateTransitions(t *testing.T) {
	fixture := setupSyncStateTest(t)
	defer teardownSyncStateTest(fixture)

	// Create initial LOCAL_ONLY sync state
	err := CreateSyncState(fixture.db, fixture.testIssueID, SyncStateLocalOnly, nil)
	if err != nil {
		t.Fatalf("Failed to create LOCAL_ONLY sync state: %v", err)
	}

	t.Run("LOCAL_ONLY → PENDING_PUSH", func(t *testing.T) {
		// Transition to PENDING_PUSH (user requests push)
		err := UpdateSyncState(fixture.db, fixture.testIssueID, SyncStatePendingPush, nil, nil)
		if err != nil {
			t.Errorf("Failed to transition LOCAL_ONLY → PENDING_PUSH: %v", err)
		}

		// Verify state
		state, err := GetSyncState(fixture.db, fixture.testIssueID)
		if err != nil {
			t.Errorf("Failed to get sync state: %v", err)
		}
		if state.SyncState != SyncStatePendingPush {
			t.Errorf("Expected PENDING_PUSH, got %s", state.SyncState)
		}
	})

	t.Run("LOCAL_ONLY → CONFLICTED", func(t *testing.T) {
		// Reset to LOCAL_ONLY
		err := UpdateSyncState(fixture.db, fixture.testIssueID, SyncStateLocalOnly, nil, nil)
		if err != nil {
			t.Fatalf("Failed to reset to LOCAL_ONLY: %v", err)
		}

		// Transition to CONFLICTED (GitHub issue appears with same title)
		githubID := int64(12345)
		err = UpdateSyncState(fixture.db, fixture.testIssueID, SyncStateConflicted, &githubID, nil)
		if err != nil {
			t.Errorf("Failed to transition LOCAL_ONLY → CONFLICTED: %v", err)
		}

		// Verify state and GitHub ID
		state, err := GetSyncState(fixture.db, fixture.testIssueID)
		if err != nil {
			t.Errorf("Failed to get sync state: %v", err)
		}
		if state.SyncState != SyncStateConflicted {
			t.Errorf("Expected CONFLICTED, got %s", state.SyncState)
		}
		if state.GitHubID == nil || *state.GitHubID != githubID {
			t.Errorf("Expected GitHub ID %d, got %v", githubID, state.GitHubID)
		}
	})
}

// Test all state transitions from PENDING_PUSH
func TestPendingPushStateTransitions(t *testing.T) {
	fixture := setupSyncStateTest(t)
	defer teardownSyncStateTest(fixture)

	// Create initial PENDING_PUSH sync state
	err := CreateSyncState(fixture.db, fixture.testIssueID, SyncStatePendingPush, nil)
	if err != nil {
		t.Fatalf("Failed to create PENDING_PUSH sync state: %v", err)
	}

	t.Run("PENDING_PUSH → SYNCED", func(t *testing.T) {
		// Successfully created on GitHub, no further local changes
		githubID := int64(12345)
		err := UpdateSyncState(fixture.db, fixture.testIssueID, SyncStateSynced, &githubID, nil)
		if err != nil {
			t.Errorf("Failed to transition PENDING_PUSH → SYNCED: %v", err)
		}

		state, err := GetSyncState(fixture.db, fixture.testIssueID)
		if err != nil {
			t.Errorf("Failed to get sync state: %v", err)
		}
		if state.SyncState != SyncStateSynced {
			t.Errorf("Expected SYNCED, got %s", state.SyncState)
		}
		if state.GitHubID == nil || *state.GitHubID != githubID {
			t.Errorf("Expected GitHub ID %d, got %v", githubID, state.GitHubID)
		}
	})

	t.Run("PENDING_PUSH → LOCAL_MODIFIED", func(t *testing.T) {
		// Reset to PENDING_PUSH
		err := UpdateSyncState(fixture.db, fixture.testIssueID, SyncStatePendingPush, nil, nil)
		if err != nil {
			t.Fatalf("Failed to reset to PENDING_PUSH: %v", err)
		}

		// Successfully created on GitHub, but local changes made since
		githubID := int64(12346)
		err = UpdateSyncState(fixture.db, fixture.testIssueID, SyncStateLocalModified, &githubID, nil)
		if err != nil {
			t.Errorf("Failed to transition PENDING_PUSH → LOCAL_MODIFIED: %v", err)
		}

		state, err := GetSyncState(fixture.db, fixture.testIssueID)
		if err != nil {
			t.Errorf("Failed to get sync state: %v", err)
		}
		if state.SyncState != SyncStateLocalModified {
			t.Errorf("Expected LOCAL_MODIFIED, got %s", state.SyncState)
		}
	})

	t.Run("PENDING_PUSH → PUSH_FAILED", func(t *testing.T) {
		// Reset to PENDING_PUSH
		err := UpdateSyncState(fixture.db, fixture.testIssueID, SyncStatePendingPush, nil, nil)
		if err != nil {
			t.Fatalf("Failed to reset to PENDING_PUSH: %v", err)
		}

		// GitHub creation failed
		errorMsg := "GitHub API error: rate limit exceeded"
		err = UpdateSyncState(fixture.db, fixture.testIssueID, SyncStatePushFailed, nil, &errorMsg)
		if err != nil {
			t.Errorf("Failed to transition PENDING_PUSH → PUSH_FAILED: %v", err)
		}

		state, err := GetSyncState(fixture.db, fixture.testIssueID)
		if err != nil {
			t.Errorf("Failed to get sync state: %v", err)
		}
		if state.SyncState != SyncStatePushFailed {
			t.Errorf("Expected PUSH_FAILED, got %s", state.SyncState)
		}
		if state.SyncError == nil || *state.SyncError != errorMsg {
			t.Errorf("Expected error message '%s', got %v", errorMsg, state.SyncError)
		}
		if state.RetryCount != 1 {
			t.Errorf("Expected retry count 1, got %d", state.RetryCount)
		}
	})

	t.Run("PENDING_PUSH → ERROR", func(t *testing.T) {
		// Reset to PENDING_PUSH
		err := UpdateSyncState(fixture.db, fixture.testIssueID, SyncStatePendingPush, nil, nil)
		if err != nil {
			t.Fatalf("Failed to reset to PENDING_PUSH: %v", err)
		}

		// Unrecoverable error during push
		errorMsg := "Unrecoverable error: invalid token"
		err = UpdateSyncState(fixture.db, fixture.testIssueID, SyncStateError, nil, &errorMsg)
		if err != nil {
			t.Errorf("Failed to transition PENDING_PUSH → ERROR: %v", err)
		}

		state, err := GetSyncState(fixture.db, fixture.testIssueID)
		if err != nil {
			t.Errorf("Failed to get sync state: %v", err)
		}
		if state.SyncState != SyncStateError {
			t.Errorf("Expected ERROR, got %s", state.SyncState)
		}
		if state.SyncError == nil || *state.SyncError != errorMsg {
			t.Errorf("Expected error message '%s', got %v", errorMsg, state.SyncError)
		}
	})
}

// Test all state transitions from PUSH_FAILED
func TestPushFailedStateTransitions(t *testing.T) {
	fixture := setupSyncStateTest(t)
	defer teardownSyncStateTest(fixture)

	// Create initial PUSH_FAILED sync state
	errorMsg := "GitHub API error"
	err := CreateSyncState(fixture.db, fixture.testIssueID, SyncStatePushFailed, nil)
	if err != nil {
		t.Fatalf("Failed to create PUSH_FAILED sync state: %v", err)
	}
	err = UpdateSyncState(fixture.db, fixture.testIssueID, SyncStatePushFailed, nil, &errorMsg)
	if err != nil {
		t.Fatalf("Failed to set error message: %v", err)
	}

	t.Run("PUSH_FAILED → PENDING_PUSH", func(t *testing.T) {
		// Retry push operation
		err := UpdateSyncState(fixture.db, fixture.testIssueID, SyncStatePendingPush, nil, nil)
		if err != nil {
			t.Errorf("Failed to transition PUSH_FAILED → PENDING_PUSH: %v", err)
		}

		state, err := GetSyncState(fixture.db, fixture.testIssueID)
		if err != nil {
			t.Errorf("Failed to get sync state: %v", err)
		}
		if state.SyncState != SyncStatePendingPush {
			t.Errorf("Expected PENDING_PUSH, got %s", state.SyncState)
		}
	})

	t.Run("PUSH_FAILED → LOCAL_ONLY", func(t *testing.T) {
		// Reset to PUSH_FAILED
		err := UpdateSyncState(fixture.db, fixture.testIssueID, SyncStatePushFailed, nil, &errorMsg)
		if err != nil {
			t.Fatalf("Failed to reset to PUSH_FAILED: %v", err)
		}

		// User cancels push, keeps local-only
		err = UpdateSyncState(fixture.db, fixture.testIssueID, SyncStateLocalOnly, nil, nil)
		if err != nil {
			t.Errorf("Failed to transition PUSH_FAILED → LOCAL_ONLY: %v", err)
		}

		state, err := GetSyncState(fixture.db, fixture.testIssueID)
		if err != nil {
			t.Errorf("Failed to get sync state: %v", err)
		}
		if state.SyncState != SyncStateLocalOnly {
			t.Errorf("Expected LOCAL_ONLY, got %s", state.SyncState)
		}
	})

	t.Run("PUSH_FAILED → ERROR", func(t *testing.T) {
		// Reset to PUSH_FAILED
		err := UpdateSyncState(fixture.db, fixture.testIssueID, SyncStatePushFailed, nil, &errorMsg)
		if err != nil {
			t.Fatalf("Failed to reset to PUSH_FAILED: %v", err)
		}

		// Give up after too many retries
		finalError := "Too many retries, giving up"
		err = UpdateSyncState(fixture.db, fixture.testIssueID, SyncStateError, nil, &finalError)
		if err != nil {
			t.Errorf("Failed to transition PUSH_FAILED → ERROR: %v", err)
		}

		state, err := GetSyncState(fixture.db, fixture.testIssueID)
		if err != nil {
			t.Errorf("Failed to get sync state: %v", err)
		}
		if state.SyncState != SyncStateError {
			t.Errorf("Expected ERROR, got %s", state.SyncState)
		}
		if state.SyncError == nil || *state.SyncError != finalError {
			t.Errorf("Expected error message '%s', got %v", finalError, state.SyncError)
		}
	})
}

// Test all state transitions from SYNCED
func TestSyncedStateTransitions(t *testing.T) {
	fixture := setupSyncStateTest(t)
	defer teardownSyncStateTest(fixture)

	// Create initial SYNCED sync state
	githubID := int64(12345)
	err := CreateSyncState(fixture.db, fixture.testIssueID, SyncStateSynced, &githubID)
	if err != nil {
		t.Fatalf("Failed to create SYNCED sync state: %v", err)
	}

	t.Run("SYNCED → LOCAL_MODIFIED", func(t *testing.T) {
		// User modifies issue locally
		err := UpdateSyncState(fixture.db, fixture.testIssueID, SyncStateLocalModified, &githubID, nil)
		if err != nil {
			t.Errorf("Failed to transition SYNCED → LOCAL_MODIFIED: %v", err)
		}

		state, err := GetSyncState(fixture.db, fixture.testIssueID)
		if err != nil {
			t.Errorf("Failed to get sync state: %v", err)
		}
		if state.SyncState != SyncStateLocalModified {
			t.Errorf("Expected LOCAL_MODIFIED, got %s", state.SyncState)
		}
	})

	t.Run("SYNCED → CONFLICTED", func(t *testing.T) {
		// Reset to SYNCED
		err := UpdateSyncState(fixture.db, fixture.testIssueID, SyncStateSynced, &githubID, nil)
		if err != nil {
			t.Fatalf("Failed to reset to SYNCED: %v", err)
		}

		// Remote changes detected during fetch
		err = UpdateSyncState(fixture.db, fixture.testIssueID, SyncStateConflicted, &githubID, nil)
		if err != nil {
			t.Errorf("Failed to transition SYNCED → CONFLICTED: %v", err)
		}

		state, err := GetSyncState(fixture.db, fixture.testIssueID)
		if err != nil {
			t.Errorf("Failed to get sync state: %v", err)
		}
		if state.SyncState != SyncStateConflicted {
			t.Errorf("Expected CONFLICTED, got %s", state.SyncState)
		}
	})

	t.Run("SYNCED → ERROR", func(t *testing.T) {
		// Reset to SYNCED
		err := UpdateSyncState(fixture.db, fixture.testIssueID, SyncStateSynced, &githubID, nil)
		if err != nil {
			t.Fatalf("Failed to reset to SYNCED: %v", err)
		}

		// Unrecoverable error
		errorMsg := "Database corruption detected"
		err = UpdateSyncState(fixture.db, fixture.testIssueID, SyncStateError, &githubID, &errorMsg)
		if err != nil {
			t.Errorf("Failed to transition SYNCED → ERROR: %v", err)
		}

		state, err := GetSyncState(fixture.db, fixture.testIssueID)
		if err != nil {
			t.Errorf("Failed to get sync state: %v", err)
		}
		if state.SyncState != SyncStateError {
			t.Errorf("Expected ERROR, got %s", state.SyncState)
		}
	})
}

// Test all state transitions from LOCAL_MODIFIED
func TestLocalModifiedStateTransitions(t *testing.T) {
	fixture := setupSyncStateTest(t)
	defer teardownSyncStateTest(fixture)

	// Create initial LOCAL_MODIFIED sync state
	githubID := int64(12345)
	err := CreateSyncState(fixture.db, fixture.testIssueID, SyncStateLocalModified, &githubID)
	if err != nil {
		t.Fatalf("Failed to create LOCAL_MODIFIED sync state: %v", err)
	}

	t.Run("LOCAL_MODIFIED → PENDING_SYNC", func(t *testing.T) {
		// User requests sync to GitHub
		err := UpdateSyncState(fixture.db, fixture.testIssueID, SyncStatePendingSync, &githubID, nil)
		if err != nil {
			t.Errorf("Failed to transition LOCAL_MODIFIED → PENDING_SYNC: %v", err)
		}

		state, err := GetSyncState(fixture.db, fixture.testIssueID)
		if err != nil {
			t.Errorf("Failed to get sync state: %v", err)
		}
		if state.SyncState != SyncStatePendingSync {
			t.Errorf("Expected PENDING_SYNC, got %s", state.SyncState)
		}
	})

	t.Run("LOCAL_MODIFIED → CONFLICTED", func(t *testing.T) {
		// Reset to LOCAL_MODIFIED
		err := UpdateSyncState(fixture.db, fixture.testIssueID, SyncStateLocalModified, &githubID, nil)
		if err != nil {
			t.Fatalf("Failed to reset to LOCAL_MODIFIED: %v", err)
		}

		// Remote changes detected during fetch
		err = UpdateSyncState(fixture.db, fixture.testIssueID, SyncStateConflicted, &githubID, nil)
		if err != nil {
			t.Errorf("Failed to transition LOCAL_MODIFIED → CONFLICTED: %v", err)
		}

		state, err := GetSyncState(fixture.db, fixture.testIssueID)
		if err != nil {
			t.Errorf("Failed to get sync state: %v", err)
		}
		if state.SyncState != SyncStateConflicted {
			t.Errorf("Expected CONFLICTED, got %s", state.SyncState)
		}
	})

	t.Run("LOCAL_MODIFIED → SYNCED", func(t *testing.T) {
		// Reset to LOCAL_MODIFIED
		err := UpdateSyncState(fixture.db, fixture.testIssueID, SyncStateLocalModified, &githubID, nil)
		if err != nil {
			t.Fatalf("Failed to reset to LOCAL_MODIFIED: %v", err)
		}

		// User discards local changes
		err = UpdateSyncState(fixture.db, fixture.testIssueID, SyncStateSynced, &githubID, nil)
		if err != nil {
			t.Errorf("Failed to transition LOCAL_MODIFIED → SYNCED: %v", err)
		}

		state, err := GetSyncState(fixture.db, fixture.testIssueID)
		if err != nil {
			t.Errorf("Failed to get sync state: %v", err)
		}
		if state.SyncState != SyncStateSynced {
			t.Errorf("Expected SYNCED, got %s", state.SyncState)
		}
	})
}

// Test all state transitions from PENDING_SYNC
func TestPendingSyncStateTransitions(t *testing.T) {
	fixture := setupSyncStateTest(t)
	defer teardownSyncStateTest(fixture)

	// Create initial PENDING_SYNC sync state
	githubID := int64(12345)
	err := CreateSyncState(fixture.db, fixture.testIssueID, SyncStatePendingSync, &githubID)
	if err != nil {
		t.Fatalf("Failed to create PENDING_SYNC sync state: %v", err)
	}

	t.Run("PENDING_SYNC → SYNCED", func(t *testing.T) {
		// Successfully synced to GitHub
		err := UpdateSyncState(fixture.db, fixture.testIssueID, SyncStateSynced, &githubID, nil)
		if err != nil {
			t.Errorf("Failed to transition PENDING_SYNC → SYNCED: %v", err)
		}

		state, err := GetSyncState(fixture.db, fixture.testIssueID)
		if err != nil {
			t.Errorf("Failed to get sync state: %v", err)
		}
		if state.SyncState != SyncStateSynced {
			t.Errorf("Expected SYNCED, got %s", state.SyncState)
		}
	})

	t.Run("PENDING_SYNC → SYNC_FAILED", func(t *testing.T) {
		// Reset to PENDING_SYNC
		err := UpdateSyncState(fixture.db, fixture.testIssueID, SyncStatePendingSync, &githubID, nil)
		if err != nil {
			t.Fatalf("Failed to reset to PENDING_SYNC: %v", err)
		}

		// Sync to GitHub failed
		errorMsg := "GitHub API error: 403 Forbidden"
		err = UpdateSyncState(fixture.db, fixture.testIssueID, SyncStateSyncFailed, &githubID, &errorMsg)
		if err != nil {
			t.Errorf("Failed to transition PENDING_SYNC → SYNC_FAILED: %v", err)
		}

		state, err := GetSyncState(fixture.db, fixture.testIssueID)
		if err != nil {
			t.Errorf("Failed to get sync state: %v", err)
		}
		if state.SyncState != SyncStateSyncFailed {
			t.Errorf("Expected SYNC_FAILED, got %s", state.SyncState)
		}
		if state.SyncError == nil || *state.SyncError != errorMsg {
			t.Errorf("Expected error message '%s', got %v", errorMsg, state.SyncError)
		}
	})

	t.Run("PENDING_SYNC → CONFLICTED", func(t *testing.T) {
		// Reset to PENDING_SYNC
		err := UpdateSyncState(fixture.db, fixture.testIssueID, SyncStatePendingSync, &githubID, nil)
		if err != nil {
			t.Fatalf("Failed to reset to PENDING_SYNC: %v", err)
		}

		// Remote changes detected during sync
		err = UpdateSyncState(fixture.db, fixture.testIssueID, SyncStateConflicted, &githubID, nil)
		if err != nil {
			t.Errorf("Failed to transition PENDING_SYNC → CONFLICTED: %v", err)
		}

		state, err := GetSyncState(fixture.db, fixture.testIssueID)
		if err != nil {
			t.Errorf("Failed to get sync state: %v", err)
		}
		if state.SyncState != SyncStateConflicted {
			t.Errorf("Expected CONFLICTED, got %s", state.SyncState)
		}
	})

	t.Run("PENDING_SYNC → LOCAL_MODIFIED", func(t *testing.T) {
		// Reset to PENDING_SYNC
		err := UpdateSyncState(fixture.db, fixture.testIssueID, SyncStatePendingSync, &githubID, nil)
		if err != nil {
			t.Fatalf("Failed to reset to PENDING_SYNC: %v", err)
		}

		// Sync cancelled, local changes remain
		err = UpdateSyncState(fixture.db, fixture.testIssueID, SyncStateLocalModified, &githubID, nil)
		if err != nil {
			t.Errorf("Failed to transition PENDING_SYNC → LOCAL_MODIFIED: %v", err)
		}

		state, err := GetSyncState(fixture.db, fixture.testIssueID)
		if err != nil {
			t.Errorf("Failed to get sync state: %v", err)
		}
		if state.SyncState != SyncStateLocalModified {
			t.Errorf("Expected LOCAL_MODIFIED, got %s", state.SyncState)
		}
	})
}

// Test all state transitions from SYNC_FAILED
func TestSyncFailedStateTransitions(t *testing.T) {
	fixture := setupSyncStateTest(t)
	defer teardownSyncStateTest(fixture)

	// Create initial SYNC_FAILED sync state
	githubID := int64(12345)
	errorMsg := "Sync failed"
	err := CreateSyncState(fixture.db, fixture.testIssueID, SyncStateSyncFailed, &githubID)
	if err != nil {
		t.Fatalf("Failed to create SYNC_FAILED sync state: %v", err)
	}
	err = UpdateSyncState(fixture.db, fixture.testIssueID, SyncStateSyncFailed, &githubID, &errorMsg)
	if err != nil {
		t.Fatalf("Failed to set error message: %v", err)
	}

	t.Run("SYNC_FAILED → PENDING_SYNC", func(t *testing.T) {
		// Retry sync operation
		err := UpdateSyncState(fixture.db, fixture.testIssueID, SyncStatePendingSync, &githubID, nil)
		if err != nil {
			t.Errorf("Failed to transition SYNC_FAILED → PENDING_SYNC: %v", err)
		}

		state, err := GetSyncState(fixture.db, fixture.testIssueID)
		if err != nil {
			t.Errorf("Failed to get sync state: %v", err)
		}
		if state.SyncState != SyncStatePendingSync {
			t.Errorf("Expected PENDING_SYNC, got %s", state.SyncState)
		}
	})

	t.Run("SYNC_FAILED → LOCAL_MODIFIED", func(t *testing.T) {
		// Reset to SYNC_FAILED
		err := UpdateSyncState(fixture.db, fixture.testIssueID, SyncStateSyncFailed, &githubID, &errorMsg)
		if err != nil {
			t.Fatalf("Failed to reset to SYNC_FAILED: %v", err)
		}

		// User keeps local changes
		err = UpdateSyncState(fixture.db, fixture.testIssueID, SyncStateLocalModified, &githubID, nil)
		if err != nil {
			t.Errorf("Failed to transition SYNC_FAILED → LOCAL_MODIFIED: %v", err)
		}

		state, err := GetSyncState(fixture.db, fixture.testIssueID)
		if err != nil {
			t.Errorf("Failed to get sync state: %v", err)
		}
		if state.SyncState != SyncStateLocalModified {
			t.Errorf("Expected LOCAL_MODIFIED, got %s", state.SyncState)
		}
	})

	t.Run("SYNC_FAILED → ERROR", func(t *testing.T) {
		// Reset to SYNC_FAILED
		err := UpdateSyncState(fixture.db, fixture.testIssueID, SyncStateSyncFailed, &githubID, &errorMsg)
		if err != nil {
			t.Fatalf("Failed to reset to SYNC_FAILED: %v", err)
		}

		// Give up after too many retries
		finalError := "Too many sync retries, giving up"
		err = UpdateSyncState(fixture.db, fixture.testIssueID, SyncStateError, &githubID, &finalError)
		if err != nil {
			t.Errorf("Failed to transition SYNC_FAILED → ERROR: %v", err)
		}

		state, err := GetSyncState(fixture.db, fixture.testIssueID)
		if err != nil {
			t.Errorf("Failed to get sync state: %v", err)
		}
		if state.SyncState != SyncStateError {
			t.Errorf("Expected ERROR, got %s", state.SyncState)
		}
		if state.SyncError == nil || *state.SyncError != finalError {
			t.Errorf("Expected error message '%s', got %v", finalError, state.SyncError)
		}
	})
}

// Test all state transitions from CONFLICTED
func TestConflictedStateTransitions(t *testing.T) {
	fixture := setupSyncStateTest(t)
	defer teardownSyncStateTest(fixture)

	// Create initial CONFLICTED sync state
	githubID := int64(12345)
	err := CreateSyncState(fixture.db, fixture.testIssueID, SyncStateConflicted, &githubID)
	if err != nil {
		t.Fatalf("Failed to create CONFLICTED sync state: %v", err)
	}

	t.Run("CONFLICTED → SYNCED", func(t *testing.T) {
		// User resolves conflict, accepts remote version
		err := UpdateSyncState(fixture.db, fixture.testIssueID, SyncStateSynced, &githubID, nil)
		if err != nil {
			t.Errorf("Failed to transition CONFLICTED → SYNCED: %v", err)
		}

		state, err := GetSyncState(fixture.db, fixture.testIssueID)
		if err != nil {
			t.Errorf("Failed to get sync state: %v", err)
		}
		if state.SyncState != SyncStateSynced {
			t.Errorf("Expected SYNCED, got %s", state.SyncState)
		}
	})

	t.Run("CONFLICTED → LOCAL_MODIFIED", func(t *testing.T) {
		// Reset to CONFLICTED
		err := UpdateSyncState(fixture.db, fixture.testIssueID, SyncStateConflicted, &githubID, nil)
		if err != nil {
			t.Fatalf("Failed to reset to CONFLICTED: %v", err)
		}

		// User resolves conflict, keeps local version
		err = UpdateSyncState(fixture.db, fixture.testIssueID, SyncStateLocalModified, &githubID, nil)
		if err != nil {
			t.Errorf("Failed to transition CONFLICTED → LOCAL_MODIFIED: %v", err)
		}

		state, err := GetSyncState(fixture.db, fixture.testIssueID)
		if err != nil {
			t.Errorf("Failed to get sync state: %v", err)
		}
		if state.SyncState != SyncStateLocalModified {
			t.Errorf("Expected LOCAL_MODIFIED, got %s", state.SyncState)
		}
	})

	t.Run("CONFLICTED → PENDING_SYNC", func(t *testing.T) {
		// Reset to CONFLICTED
		err := UpdateSyncState(fixture.db, fixture.testIssueID, SyncStateConflicted, &githubID, nil)
		if err != nil {
			t.Fatalf("Failed to reset to CONFLICTED: %v", err)
		}

		// User resolves conflict, merges and syncs
		err = UpdateSyncState(fixture.db, fixture.testIssueID, SyncStatePendingSync, &githubID, nil)
		if err != nil {
			t.Errorf("Failed to transition CONFLICTED → PENDING_SYNC: %v", err)
		}

		state, err := GetSyncState(fixture.db, fixture.testIssueID)
		if err != nil {
			t.Errorf("Failed to get sync state: %v", err)
		}
		if state.SyncState != SyncStatePendingSync {
			t.Errorf("Expected PENDING_SYNC, got %s", state.SyncState)
		}
	})

	t.Run("CONFLICTED → ERROR", func(t *testing.T) {
		// Reset to CONFLICTED
		err := UpdateSyncState(fixture.db, fixture.testIssueID, SyncStateConflicted, &githubID, nil)
		if err != nil {
			t.Fatalf("Failed to reset to CONFLICTED: %v", err)
		}

		// Conflict resolution failed
		errorMsg := "Conflict resolution failed: data corruption"
		err = UpdateSyncState(fixture.db, fixture.testIssueID, SyncStateError, &githubID, &errorMsg)
		if err != nil {
			t.Errorf("Failed to transition CONFLICTED → ERROR: %v", err)
		}

		state, err := GetSyncState(fixture.db, fixture.testIssueID)
		if err != nil {
			t.Errorf("Failed to get sync state: %v", err)
		}
		if state.SyncState != SyncStateError {
			t.Errorf("Expected ERROR, got %s", state.SyncState)
		}
		if state.SyncError == nil || *state.SyncError != errorMsg {
			t.Errorf("Expected error message '%s', got %v", errorMsg, state.SyncError)
		}
	})
}

// Test sync state CRUD operations
func TestSyncStateCRUD(t *testing.T) {
	fixture := setupSyncStateTest(t)
	defer teardownSyncStateTest(fixture)

	t.Run("CreateSyncState", func(t *testing.T) {
		githubID := int64(99999)
		err := CreateSyncState(fixture.db, fixture.testIssueID, SyncStateLocalOnly, &githubID)
		if err != nil {
			t.Errorf("Failed to create sync state: %v", err)
		}

		// Verify creation
		state, err := GetSyncState(fixture.db, fixture.testIssueID)
		if err != nil {
			t.Errorf("Failed to get sync state: %v", err)
		}
		if state.SyncState != SyncStateLocalOnly {
			t.Errorf("Expected LOCAL_ONLY, got %s", state.SyncState)
		}
		if state.GitHubID == nil || *state.GitHubID != githubID {
			t.Errorf("Expected GitHub ID %d, got %v", githubID, state.GitHubID)
		}
	})

	t.Run("GetSyncStatesByState", func(t *testing.T) {
		// Create another issue with PENDING_PUSH state
		result, err := fixture.db.Exec(`
			INSERT INTO issues (project_id, number, title, body, state, created_at, updated_at)
			VALUES (?, 2, 'Test Issue 2', 'Test body 2', 'open', ?, ?)
		`, fixture.testProjectID, time.Now().Format(time.RFC3339), time.Now().Format(time.RFC3339))
		if err != nil {
			t.Fatalf("Failed to create second test issue: %v", err)
		}

		issueID2, err := result.LastInsertId()
		if err != nil {
			t.Fatalf("Failed to get second test issue ID: %v", err)
		}

		err = CreateSyncState(fixture.db, issueID2, SyncStatePendingPush, nil)
		if err != nil {
			t.Fatalf("Failed to create PENDING_PUSH sync state: %v", err)
		}

		// Query by state
		pendingStates, err := GetSyncStatesByState(fixture.db, SyncStatePendingPush)
		if err != nil {
			t.Errorf("Failed to get PENDING_PUSH states: %v", err)
		}
		if len(pendingStates) != 1 {
			t.Errorf("Expected 1 PENDING_PUSH state, got %d", len(pendingStates))
		}
		if pendingStates[0].IssueLocalID != issueID2 {
			t.Errorf("Expected issue ID %d, got %d", issueID2, pendingStates[0].IssueLocalID)
		}
	})

	t.Run("GetSyncStateSummary", func(t *testing.T) {
		summary, err := GetSyncStateSummary(fixture.db)
		if err != nil {
			t.Errorf("Failed to get sync state summary: %v", err)
		}

		// Should have LOCAL_ONLY and PENDING_PUSH states from previous tests
		if count, exists := summary[SyncStateLocalOnly]; !exists || count != 1 {
			t.Errorf("Expected 1 LOCAL_ONLY issue, got %d", count)
		}
		if count, exists := summary[SyncStatePendingPush]; !exists || count != 1 {
			t.Errorf("Expected 1 PENDING_PUSH issue, got %d", count)
		}
	})
}

// Test retry count increment
func TestRetryCountIncrement(t *testing.T) {
	fixture := setupSyncStateTest(t)
	defer teardownSyncStateTest(fixture)

	// Create sync state
	err := CreateSyncState(fixture.db, fixture.testIssueID, SyncStatePendingPush, nil)
	if err != nil {
		t.Fatalf("Failed to create sync state: %v", err)
	}

	// Transition to PUSH_FAILED (should increment retry count)
	errorMsg := "First failure"
	err = UpdateSyncState(fixture.db, fixture.testIssueID, SyncStatePushFailed, nil, &errorMsg)
	if err != nil {
		t.Errorf("Failed to update to PUSH_FAILED: %v", err)
	}

	state, err := GetSyncState(fixture.db, fixture.testIssueID)
	if err != nil {
		t.Errorf("Failed to get sync state: %v", err)
	}
	if state.RetryCount != 1 {
		t.Errorf("Expected retry count 1, got %d", state.RetryCount)
	}

	// Fail again (should increment retry count again)
	errorMsg2 := "Second failure"
	err = UpdateSyncState(fixture.db, fixture.testIssueID, SyncStatePushFailed, nil, &errorMsg2)
	if err != nil {
		t.Errorf("Failed to update to PUSH_FAILED again: %v", err)
	}

	state, err = GetSyncState(fixture.db, fixture.testIssueID)
	if err != nil {
		t.Errorf("Failed to get sync state: %v", err)
	}
	if state.RetryCount != 2 {
		t.Errorf("Expected retry count 2, got %d", state.RetryCount)
	}

	// Transition to success (retry count should remain)
	githubID := int64(12345)
	err = UpdateSyncState(fixture.db, fixture.testIssueID, SyncStateSynced, &githubID, nil)
	if err != nil {
		t.Errorf("Failed to update to SYNCED: %v", err)
	}

	state, err = GetSyncState(fixture.db, fixture.testIssueID)
	if err != nil {
		t.Errorf("Failed to get sync state: %v", err)
	}
	if state.RetryCount != 2 {
		t.Errorf("Expected retry count to remain 2, got %d", state.RetryCount)
	}
}

// Test sync attempt timestamp tracking
func TestSyncAttemptTracking(t *testing.T) {
	fixture := setupSyncStateTest(t)
	defer teardownSyncStateTest(fixture)

	// Create sync state
	err := CreateSyncState(fixture.db, fixture.testIssueID, SyncStateLocalOnly, nil)
	if err != nil {
		t.Fatalf("Failed to create sync state: %v", err)
	}

	// Initial state should have no sync attempt
	state, err := GetSyncState(fixture.db, fixture.testIssueID)
	if err != nil {
		t.Errorf("Failed to get sync state: %v", err)
	}
	if state.LastSyncAttempt != nil {
		t.Errorf("Expected no sync attempt initially, got %v", state.LastSyncAttempt)
	}

	// Transition to PENDING_PUSH (should set sync attempt timestamp)
	err = UpdateSyncState(fixture.db, fixture.testIssueID, SyncStatePendingPush, nil, nil)
	if err != nil {
		t.Errorf("Failed to update to PENDING_PUSH: %v", err)
	}

	state, err = GetSyncState(fixture.db, fixture.testIssueID)
	if err != nil {
		t.Errorf("Failed to get sync state: %v", err)
	}
	if state.LastSyncAttempt == nil {
		t.Errorf("Expected sync attempt timestamp to be set")
	}

	// Record the attempt time
	firstAttempt := *state.LastSyncAttempt

	// Wait a bit and transition to PENDING_SYNC (should update sync attempt timestamp)
	time.Sleep(1 * time.Second)
	githubID := int64(12345)
	err = UpdateSyncState(fixture.db, fixture.testIssueID, SyncStatePendingSync, &githubID, nil)
	if err != nil {
		t.Errorf("Failed to update to PENDING_SYNC: %v", err)
	}

	state, err = GetSyncState(fixture.db, fixture.testIssueID)
	if err != nil {
		t.Errorf("Failed to get sync state: %v", err)
	}
	if state.LastSyncAttempt == nil {
		t.Errorf("Expected sync attempt timestamp to be set")
	}
	// Use Unix timestamp comparison for more reliable testing
	if state.LastSyncAttempt.Unix() <= firstAttempt.Unix() {
		t.Errorf("Expected sync attempt timestamp to be updated: first=%v, second=%v", firstAttempt, *state.LastSyncAttempt)
	}
}
