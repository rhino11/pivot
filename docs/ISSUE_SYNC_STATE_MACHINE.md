# Issue Sync State Machine Design

## Problem Statement

Currently, the pivot sync system lacks proper state tracking for issues as they move between local database and GitHub. This creates several problems:

1. **Data Loss**: Local changes get overwritten by GitHub data during sync
2. **No Bidirectional Sync**: Cannot push local-only issues to GitHub
3. **No Conflict Detection**: Cannot identify when both local and remote changes exist
4. **Limited Offline Support**: Cannot work with locally created issues

## Proposed Solution: Finite State Machine

We propose implementing a finite state machine to track the sync state of each issue.

## Issue Sync States

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   LOCAL_ONLY    │───▶│   PENDING_PUSH  │───▶│     SYNCED      │
│                 │    │                 │    │                 │
│ Created locally │    │ Queued for      │    │ Exists in both  │
│ No GitHub ID    │    │ GitHub creation │    │ Local & GitHub  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       ▼
         │                       │              ┌─────────────────┐
         │                       │              │  LOCAL_MODIFIED │
         │                       │              │                 │
         │                       │              │ Modified locally│
         │                       │              │ after last sync │
         │                       │              └─────────────────┘
         │                       │                       │
         │                       ▼                       ▼
         │              ┌─────────────────┐    ┌─────────────────┐
         │              │   PUSH_FAILED   │    │  PENDING_SYNC   │
         │              │                 │    │                 │
         │              │ GitHub creation │    │ Queued for sync │
         │              │ failed          │    │ to GitHub       │
         │              └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       ▼
         │                       │              ┌─────────────────┐
         │                       │              │   SYNC_FAILED   │
         │                       │              │                 │
         │                       │              │ Sync to GitHub  │
         │                       │              │ failed          │
         │                       │              └─────────────────┘
         │                       │                       │
         │                       └───────┬───────────────┘
         │                               │
         ▼                               ▼
┌─────────────────┐              ┌─────────────────┐
│   CONFLICTED    │              │     ERROR       │
│                 │              │                 │
│ Local & remote  │              │ Unrecoverable   │
│ changes exist   │              │ error state     │
└─────────────────┘              └─────────────────┘
```

## State Definitions

| State | Description | GitHub ID | Local Changes | Remote Changes |
|-------|-------------|-----------|---------------|----------------|
| `LOCAL_ONLY` | Issue created locally, not yet pushed to GitHub | NULL | Yes | No |
| `PENDING_PUSH` | Local issue queued for creation on GitHub | NULL | Yes | No |
| `PUSH_FAILED` | Failed to create issue on GitHub | NULL | Yes | No |
| `SYNCED` | Issue exists in both places, no local changes | Present | No | May exist |
| `LOCAL_MODIFIED` | Issue exists in both, but modified locally | Present | Yes | May exist |
| `PENDING_SYNC` | Local changes queued for sync to GitHub | Present | Yes | May exist |
| `SYNC_FAILED` | Failed to sync local changes to GitHub | Present | Yes | May exist |
| `CONFLICTED` | Both local and remote changes detected | Present | Yes | Yes |
| `ERROR` | Unrecoverable error state | Any | Any | Any |

## State Transitions

### From LOCAL_ONLY
- → `PENDING_PUSH`: User requests sync/push to GitHub
- → `CONFLICTED`: Manual state change if GitHub issue appears with same title

### From PENDING_PUSH
- → `SYNCED`: Successfully created on GitHub, no further local changes
- → `LOCAL_MODIFIED`: Successfully created on GitHub, but local changes made since
- → `PUSH_FAILED`: GitHub creation failed
- → `ERROR`: Unrecoverable error during push

### From PUSH_FAILED
- → `PENDING_PUSH`: Retry push operation
- → `LOCAL_ONLY`: User cancels push, keeps local-only
- → `ERROR`: Give up after too many retries

### From SYNCED
- → `LOCAL_MODIFIED`: User modifies issue locally
- → `CONFLICTED`: Remote changes detected during fetch
- → `ERROR`: Unrecoverable error

### From LOCAL_MODIFIED
- → `PENDING_SYNC`: User requests sync to GitHub
- → `CONFLICTED`: Remote changes detected during fetch
- → `SYNCED`: User discards local changes

### From PENDING_SYNC
- → `SYNCED`: Successfully synced to GitHub
- → `SYNC_FAILED`: Sync to GitHub failed
- → `CONFLICTED`: Remote changes detected during sync
- → `LOCAL_MODIFIED`: Sync cancelled, local changes remain

### From SYNC_FAILED
- → `PENDING_SYNC`: Retry sync operation
- → `LOCAL_MODIFIED`: User keeps local changes
- → `ERROR`: Give up after too many retries

### From CONFLICTED
- → `SYNCED`: User resolves conflict, accepts remote version
- → `LOCAL_MODIFIED`: User resolves conflict, keeps local version
- → `PENDING_SYNC`: User resolves conflict, merges and syncs
- → `ERROR`: Conflict resolution failed

## Database Schema Changes

### New Table: issue_sync_state

```sql
CREATE TABLE issue_sync_state (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    issue_local_id INTEGER NOT NULL,  -- FK to issues table
    github_id INTEGER,                -- GitHub issue ID (NULL for local-only)
    sync_state TEXT NOT NULL,         -- State from enum above
    last_local_modified TEXT,         -- When issue was last modified locally
    last_remote_modified TEXT,        -- When issue was last modified on GitHub
    last_sync_attempt TEXT,           -- When we last tried to sync
    sync_error TEXT,                  -- Last error message if any
    retry_count INTEGER DEFAULT 0,    -- How many times we've retried
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(issue_local_id) REFERENCES issues(rowid) ON DELETE CASCADE,
    UNIQUE(issue_local_id)
);
```

### Enhanced Issues Table

```sql
-- Add columns to existing issues table
ALTER TABLE issues ADD COLUMN local_modified_at TEXT;
ALTER TABLE issues ADD COLUMN sync_hash TEXT; -- Hash of content at last sync
```

## Implementation Plan

### Phase 1: Database Schema
1. Create migration script to add sync state table
2. Add helper functions for state management
3. Update issue creation/modification to update sync state

### Phase 2: State Machine Core
1. Implement state transition logic
2. Add state validation and constraints
3. Create state machine event handlers

### Phase 3: Sync Engine
1. Modify sync logic to respect states
2. Implement bidirectional sync (local → GitHub)
3. Add conflict detection and resolution

### Phase 4: CLI Commands
1. Add `pivot status` command to show sync state
2. Add `pivot push` command to push local issues
3. Add `pivot resolve` command for conflict resolution
4. Enhance `pivot sync` with state awareness

### Phase 5: Advanced Features
1. Automatic retry logic for failed operations
2. Conflict resolution UI/prompts
3. Batch operations for multiple issues
4. Background sync daemon mode

## Benefits

1. **Data Safety**: No more accidental overwrites of local changes
2. **Bidirectional Sync**: Can create issues locally and push to GitHub
3. **Conflict Detection**: Know when manual intervention is needed
4. **Better Offline Support**: Work locally without losing data
5. **Transparency**: Always know the sync state of each issue
6. **Reliability**: Retry failed operations, handle errors gracefully

## Usage Examples

```bash
# Create issue locally
pivot create "Fix bug in authentication"

# Check sync status
pivot status
# Output:
# LOCAL_ONLY: 3 issues
# SYNCED: 15 issues
# LOCAL_MODIFIED: 2 issues
# CONFLICTED: 1 issue

# Push local-only issues to GitHub
pivot push

# Sync all changes
pivot sync

# Resolve conflicts interactively
pivot resolve
```

This design provides a robust foundation for proper bidirectional sync while maintaining data integrity and providing clear visibility into the sync state of each issue.
