package internal

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// DBIssue represents an issue as stored in the database
type DBIssue struct {
	ID        int    `json:"id"`
	Number    int    `json:"number"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	State     string `json:"state"`
	Labels    string `json:"labels"`    // Comma-separated string
	Assignees string `json:"assignees"` // Comma-separated string
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	ClosedAt  string `json:"closed_at"`
}

// InitMultiProjectDB initializes the multi-project database schema
func InitMultiProjectDB(db *sql.DB) error {
	// Create projects table
	projectsSchema := `
	CREATE TABLE IF NOT EXISTS projects (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		owner TEXT NOT NULL,
		repo TEXT NOT NULL,
		path TEXT,
		token TEXT,
		database_path TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(owner, repo)
	);`

	if _, err := db.Exec(projectsSchema); err != nil {
		return fmt.Errorf("failed to create projects table: %w", err)
	}

	// Check if issues table already exists and has project_id column
	hasProjectID, err := hasColumn(db, "issues", "project_id")
	if err != nil {
		return fmt.Errorf("failed to check issues table structure: %w", err)
	}

	if !hasProjectID {
		// Create new issues table with project_id
		issuesSchema := `
		CREATE TABLE IF NOT EXISTS issues_new (
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
			PRIMARY KEY(github_id, project_id),
			FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
		);`

		if _, err := db.Exec(issuesSchema); err != nil {
			return fmt.Errorf("failed to create new issues table: %w", err)
		}

		// Copy data from old issues table if it exists
		if hasTable(db, "issues") {
			// This will be handled by migration function
			// For now, just rename the old table
			if _, err := db.Exec("ALTER TABLE issues RENAME TO issues_old"); err != nil {
				return fmt.Errorf("failed to rename old issues table: %w", err)
			}
		}

		// Rename new table to issues
		if _, err := db.Exec("ALTER TABLE issues_new RENAME TO issues"); err != nil {
			return fmt.Errorf("failed to rename new issues table: %w", err)
		}
	}

	return nil
}

// hasColumn checks if a table has a specific column
func hasColumn(db *sql.DB, tableName, columnName string) (bool, error) {
	query := "PRAGMA table_info(" + tableName + ")"
	rows, err := db.Query(query)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, pk int
		var defaultValue interface{}

		err = rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk)
		if err != nil {
			return false, err
		}

		if name == columnName {
			return true, nil
		}
	}

	return false, nil
}

// hasTable checks if a table exists
func hasTable(db *sql.DB, tableName string) bool {
	var count int
	query := "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?"
	err := db.QueryRow(query, tableName).Scan(&count)
	return err == nil && count > 0
}

// CreateProject creates a new project in the database
func CreateProject(db *sql.DB, project *ProjectConfig) (int64, error) {
	query := `
		INSERT INTO projects (owner, repo, path, token, database_path, updated_at)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(owner, repo) DO UPDATE SET
			path = excluded.path,
			token = excluded.token,
			database_path = excluded.database_path,
			updated_at = CURRENT_TIMESTAMP
	`

	result, err := db.Exec(query, project.Owner, project.Repo, project.Path, project.Token, project.Database)
	if err != nil {
		return 0, fmt.Errorf("failed to create/update project: %w", err)
	}

	projectID, err := result.LastInsertId()
	if err != nil {
		// If LastInsertId fails (conflict case), try to get the existing ID
		return getProjectID(db, project.Owner, project.Repo)
	}

	return projectID, nil
}

// getProjectID gets the database ID for a project by owner/repo
func getProjectID(db *sql.DB, owner, repo string) (int64, error) {
	var id int64
	query := "SELECT id FROM projects WHERE owner = ? AND repo = ?"
	err := db.QueryRow(query, owner, repo).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to get project ID: %w", err)
	}
	return id, nil
}

// FindProjectByPath finds a project by its filesystem path
func FindProjectByPath(db *sql.DB, path string) (*ProjectConfig, error) {
	query := "SELECT id, owner, repo, path, token, database_path FROM projects WHERE path = ?"
	var project ProjectConfig
	var token, dbPath sql.NullString

	err := db.QueryRow(query, path).Scan(&project.ID, &project.Owner, &project.Repo, &project.Path, &token, &dbPath)
	if err != nil {
		return nil, fmt.Errorf("project not found at path %s: %w", path, err)
	}

	if token.Valid {
		project.Token = token.String
	}
	if dbPath.Valid {
		project.Database = dbPath.String
	}

	return &project, nil
}

// FindProjectByOwnerRepo finds a project by owner and repository name
func FindProjectByOwnerRepo(db *sql.DB, owner, repo string) (*ProjectConfig, error) {
	query := "SELECT id, owner, repo, path, token, database_path FROM projects WHERE owner = ? AND repo = ?"
	var project ProjectConfig
	var path, token, dbPath sql.NullString

	err := db.QueryRow(query, owner, repo).Scan(&project.ID, &project.Owner, &project.Repo, &path, &token, &dbPath)
	if err != nil {
		return nil, fmt.Errorf("project not found for %s/%s: %w", owner, repo, err)
	}

	if path.Valid {
		project.Path = path.String
	}
	if token.Valid {
		project.Token = token.String
	}
	if dbPath.Valid {
		project.Database = dbPath.String
	}

	return &project, nil
}

// ListProjects returns all projects in the database
func ListProjects(db *sql.DB) ([]ProjectConfig, error) {
	query := "SELECT id, owner, repo, path, token, database_path FROM projects ORDER BY owner, repo"
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	defer rows.Close()

	var projects []ProjectConfig
	for rows.Next() {
		var project ProjectConfig
		var path, token, dbPath sql.NullString

		err := rows.Scan(&project.ID, &project.Owner, &project.Repo, &path, &token, &dbPath)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}

		if path.Valid {
			project.Path = path.String
		}
		if token.Valid {
			project.Token = token.String
		}
		if dbPath.Valid {
			project.Database = dbPath.String
		}

		projects = append(projects, project)
	}

	return projects, nil
}

// SaveIssue saves an issue to the database for a specific project
func SaveIssue(db *sql.DB, projectID int64, issue *DBIssue) error {
	query := `
		INSERT OR REPLACE INTO issues (github_id, project_id, number, title, body, state, labels, assignees, created_at, updated_at, closed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := db.Exec(query,
		issue.ID, projectID, issue.Number, issue.Title, issue.Body,
		issue.State, issue.Labels, issue.Assignees,
		issue.CreatedAt, issue.UpdatedAt, issue.ClosedAt)

	if err != nil {
		return fmt.Errorf("failed to save issue: %w", err)
	}

	return nil
}

// GetIssuesForProject retrieves all issues for a specific project
func GetIssuesForProject(db *sql.DB, projectID int64) ([]DBIssue, error) {
	query := `
		SELECT github_id, number, title, body, state, labels, assignees, created_at, updated_at, closed_at
		FROM issues 
		WHERE project_id = ?
		ORDER BY number
	`

	rows, err := db.Query(query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to query issues: %w", err)
	}
	defer rows.Close()

	var issues []DBIssue
	for rows.Next() {
		var issue DBIssue
		var labels, assignees sql.NullString
		var closedAt sql.NullString

		err := rows.Scan(&issue.ID, &issue.Number, &issue.Title, &issue.Body,
			&issue.State, &labels, &assignees,
			&issue.CreatedAt, &issue.UpdatedAt, &closedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan issue: %w", err)
		}

		if labels.Valid {
			issue.Labels = labels.String
		}
		if assignees.Valid {
			issue.Assignees = assignees.String
		}
		if closedAt.Valid {
			issue.ClosedAt = closedAt.String
		}

		issues = append(issues, issue)
	}

	return issues, nil
}

// MigrateToMultiProject migrates a legacy single-project database to multi-project format
func MigrateToMultiProject(db *sql.DB, legacyOwner, legacyRepo, legacyPath string) error {
	// First, initialize the multi-project schema
	if err := InitMultiProjectDB(db); err != nil {
		return fmt.Errorf("failed to initialize multi-project schema: %w", err)
	}

	// Create the legacy project entry
	projectConfig := &ProjectConfig{
		Owner: legacyOwner,
		Repo:  legacyRepo,
		Path:  legacyPath,
	}

	projectID, err := CreateProject(db, projectConfig)
	if err != nil {
		return fmt.Errorf("failed to create legacy project: %w", err)
	}

	// Check if we have legacy data to migrate
	if !hasTable(db, "issues_old") {
		return nil // No legacy data to migrate
	}

	// Migrate existing issues from the old table
	query := `
		INSERT INTO issues (github_id, project_id, number, title, body, state, labels, assignees, created_at, updated_at, closed_at)
		SELECT github_id, ?, number, title, body, state, labels, assignees, created_at, updated_at, closed_at
		FROM issues_old
	`

	_, err = db.Exec(query, projectID)
	if err != nil {
		return fmt.Errorf("failed to migrate legacy issues: %w", err)
	}

	// Drop the old table
	if _, err := db.Exec("DROP TABLE issues_old"); err != nil {
		return fmt.Errorf("failed to drop old issues table: %w", err)
	}

	return nil
}

// InitMultiProjectDBFromPath initializes a multi-project database at the specified path
func InitMultiProjectDBFromPath(dbPath string) (*sql.DB, error) {
	// Resolve database path (expand ~ etc.)
	resolvedPath, err := ResolveDatabasePath(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve database path: %w", err)
	}

	// Ensure directory exists
	if err := ensureDirectoryExists(resolvedPath); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := sql.Open("sqlite3", resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := InitMultiProjectDB(db); err != nil {
		db.Close() // #nosec G104 - Intentionally ignoring close error in error path
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return db, nil
}

// ensureDirectoryExists creates the directory for the database file if it doesn't exist
func ensureDirectoryExists(dbPath string) error {
	dir := filepath.Dir(dbPath)
	if dir == "." {
		// No directory component
		return nil
	}

	return os.MkdirAll(dir, 0755) // #nosec G301 - Standard directory permissions for app data
}

// ConvertIssueToDBIssue converts a GitHub API issue to database format
func ConvertIssueToDBIssue(issue *Issue) *DBIssue {
	// Convert labels to comma-separated string
	var labels string
	for i, l := range issue.Labels {
		if i > 0 {
			labels += ","
		}
		labels += l.Name
	}

	// Convert assignees to comma-separated string
	var assignees string
	for i, a := range issue.Assignees {
		if i > 0 {
			assignees += ","
		}
		assignees += a.Login
	}

	return &DBIssue{
		ID:        issue.ID,
		Number:    issue.Number,
		Title:     issue.Title,
		Body:      issue.Body,
		State:     issue.State,
		Labels:    labels,
		Assignees: assignees,
		CreatedAt: issue.CreatedAt,
		UpdatedAt: issue.UpdatedAt,
		ClosedAt:  issue.ClosedAt,
	}
}
