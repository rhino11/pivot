package internal

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// TestMultiProjectDatabaseSchema tests the enhanced database schema for multi-project support
func TestMultiProjectDatabaseSchema(t *testing.T) {
	tempDB := "test_multiproject.db"
	defer os.Remove(tempDB)

	db, err := sql.Open("sqlite3", tempDB)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Initialize the enhanced schema
	err = InitMultiProjectDB(db)
	if err != nil {
		t.Fatalf("Failed to initialize multi-project database: %v", err)
	}

	// Verify projects table exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='projects'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check for projects table: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected projects table to exist")
	}

	// Verify issues table has project_id column
	rows, err := db.Query("PRAGMA table_info(issues)")
	if err != nil {
		t.Fatalf("Failed to get issues table info: %v", err)
	}
	defer rows.Close()

	foundProjectID := false
	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, pk int
		var defaultValue interface{}

		err = rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk)
		if err != nil {
			t.Fatalf("Failed to scan column info: %v", err)
		}

		if name == "project_id" {
			foundProjectID = true
			break
		}
	}

	if !foundProjectID {
		t.Error("Expected issues table to have project_id column")
	}
}

// TestProjectManagement tests CRUD operations for projects
func TestProjectManagement(t *testing.T) {
	tempDB := "test_project_crud.db"
	defer os.Remove(tempDB)

	db, err := sql.Open("sqlite3", tempDB)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Initialize schema
	err = InitMultiProjectDB(db)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Test creating a project
	projectConfig := &ProjectConfig{
		Owner: "testowner",
		Repo:  "testrepo",
		Path:  "/path/to/project",
		Token: "ghp_test_token",
	}

	projectID, err := CreateProject(db, projectConfig)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	if projectID <= 0 {
		t.Errorf("Expected positive project ID, got %d", projectID)
	}

	// Test finding project by path
	foundProject, err := FindProjectByPath(db, "/path/to/project")
	if err != nil {
		t.Fatalf("Failed to find project by path: %v", err)
	}

	if foundProject.Owner != "testowner" || foundProject.Repo != "testrepo" {
		t.Errorf("Found project doesn't match: %+v", foundProject)
	}

	// Test finding project by owner/repo
	foundProject2, err := FindProjectByOwnerRepo(db, "testowner", "testrepo")
	if err != nil {
		t.Fatalf("Failed to find project by owner/repo: %v", err)
	}

	if int64(foundProject2.ID) != projectID {
		t.Errorf("Expected same project ID, got %d vs %d", foundProject2.ID, projectID)
	}

	// Test listing all projects
	projects, err := ListProjects(db)
	if err != nil {
		t.Fatalf("Failed to list projects: %v", err)
	}

	if len(projects) != 1 {
		t.Errorf("Expected 1 project, got %d", len(projects))
	}
}

// TestIssueProjectAssociation tests that issues are properly associated with projects
func TestIssueProjectAssociation(t *testing.T) {
	tempDB := "test_issue_project.db"
	defer os.Remove(tempDB)

	db, err := sql.Open("sqlite3", tempDB)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Initialize schema
	err = InitMultiProjectDB(db)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Create two projects
	project1Config := &ProjectConfig{
		Owner: "owner1",
		Repo:  "repo1",
		Path:  "/path/to/project1",
	}
	project1ID, err := CreateProject(db, project1Config)
	if err != nil {
		t.Fatalf("Failed to create project 1: %v", err)
	}

	project2Config := &ProjectConfig{
		Owner: "owner2",
		Repo:  "repo2",
		Path:  "/path/to/project2",
	}
	project2ID, err := CreateProject(db, project2Config)
	if err != nil {
		t.Fatalf("Failed to create project 2: %v", err)
	}

	// Create issues for each project
	issue1 := &DBIssue{
		ID:     1001,
		Number: 1,
		Title:  "Issue for project 1",
		State:  "open",
	}
	err = SaveIssue(db, project1ID, issue1)
	if err != nil {
		t.Fatalf("Failed to save issue for project 1: %v", err)
	}

	issue2 := &DBIssue{
		ID:     2001,
		Number: 1, // Same number but different project
		Title:  "Issue for project 2",
		State:  "closed",
	}
	err = SaveIssue(db, project2ID, issue2)
	if err != nil {
		t.Fatalf("Failed to save issue for project 2: %v", err)
	}

	// Test getting issues for each project
	project1Issues, err := GetIssuesForProject(db, project1ID)
	if err != nil {
		t.Fatalf("Failed to get issues for project 1: %v", err)
	}
	if len(project1Issues) != 1 {
		t.Errorf("Expected 1 issue for project 1, got %d", len(project1Issues))
	}
	if project1Issues[0].Title != "Issue for project 1" {
		t.Errorf("Wrong issue title for project 1: %s", project1Issues[0].Title)
	}

	project2Issues, err := GetIssuesForProject(db, project2ID)
	if err != nil {
		t.Fatalf("Failed to get issues for project 2: %v", err)
	}
	if len(project2Issues) != 1 {
		t.Errorf("Expected 1 issue for project 2, got %d", len(project2Issues))
	}
	if project2Issues[0].Title != "Issue for project 2" {
		t.Errorf("Wrong issue title for project 2: %s", project2Issues[0].Title)
	}
}

// TestDatabaseMigration tests upgrading from single-project to multi-project database
func TestDatabaseMigration(t *testing.T) {
	tempDB := "test_migration.db"
	defer os.Remove(tempDB)

	// First, create a legacy single-project database
	db, err := sql.Open("sqlite3", tempDB)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create legacy schema
	legacySchema := `
	CREATE TABLE IF NOT EXISTS issues (
		github_id INTEGER PRIMARY KEY,
		number INTEGER,
		title TEXT,
		body TEXT,
		state TEXT,
		labels TEXT,
		assignees TEXT,
		created_at TEXT,
		updated_at TEXT,
		closed_at TEXT
	);`
	_, err = db.Exec(legacySchema)
	if err != nil {
		t.Fatalf("Failed to create legacy schema: %v", err)
	}

	// Insert some legacy data
	_, err = db.Exec(`INSERT INTO issues (github_id, number, title, state) VALUES (1001, 1, 'Legacy Issue', 'open')`)
	if err != nil {
		t.Fatalf("Failed to insert legacy data: %v", err)
	}

	db.Close()

	// Now test migration
	db, err = sql.Open("sqlite3", tempDB)
	if err != nil {
		t.Fatalf("Failed to reopen database: %v", err)
	}
	defer db.Close()

	// Run migration
	err = MigrateToMultiProject(db, "legacyowner", "legacyrepo", "/legacy/path")
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Verify projects table was created and has the legacy project
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM projects WHERE owner = 'legacyowner' AND repo = 'legacyrepo'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count migrated projects: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 migrated project, got %d", count)
	}

	// Verify issues table has project_id column and legacy data
	var projectID int
	var title string
	err = db.QueryRow(`
		SELECT p.id, i.title 
		FROM projects p 
		JOIN issues i ON i.project_id = p.id 
		WHERE p.owner = 'legacyowner' AND p.repo = 'legacyrepo'
	`).Scan(&projectID, &title)
	if err != nil {
		t.Fatalf("Failed to verify migrated data: %v", err)
	}

	if title != "Legacy Issue" {
		t.Errorf("Expected legacy issue title, got '%s'", title)
	}
}

// TestGetProjectID tests the getProjectID function
func TestGetProjectID(t *testing.T) {
	tempDB := "test_getprojectid.db"
	defer os.Remove(tempDB)

	db, err := sql.Open("sqlite3", tempDB)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	err = InitMultiProjectDB(db)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Create a test project
	project := &ProjectConfig{
		Owner:    "testowner",
		Repo:     "testrepo",
		Path:     "/test/path",
		Token:    "token123",
		Database: "",
	}
	projectID, err := CreateProject(db, project)
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	// Test successful retrieval
	retrievedID, err := getProjectID(db, "testowner", "testrepo")
	if err != nil {
		t.Fatalf("getProjectID failed: %v", err)
	}

	if retrievedID != projectID {
		t.Errorf("Expected project ID %d, got %d", projectID, retrievedID)
	}

	// Test non-existent project
	_, err = getProjectID(db, "nonexistent", "project")
	if err == nil {
		t.Error("Expected error for non-existent project, got nil")
	}
}

// TestInitMultiProjectDBFromPath tests the InitMultiProjectDBFromPath function
func TestInitMultiProjectDBFromPath(t *testing.T) {
	// Test with temporary path
	tempDir := t.TempDir()
	dbPath := tempDir + "/test_multiproject.db"

	db, err := InitMultiProjectDBFromPath(dbPath)
	if err != nil {
		t.Fatalf("InitMultiProjectDBFromPath failed: %v", err)
	}
	defer db.Close()

	// Verify database was created and initialized
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='projects'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check for projects table: %v", err)
	}
	if count != 1 {
		t.Error("Expected projects table to exist")
	}

	// Test with home directory expansion
	// We need to test path resolution separately as we can't write to actual home

	// Test invalid path (should fail gracefully)
	_, err = InitMultiProjectDBFromPath("/invalid/readonly/path/db.sqlite")
	if err == nil {
		t.Error("Expected error for invalid path, got nil")
	}
}

// TestEnsureDirectoryExists tests the ensureDirectoryExists function
func TestEnsureDirectoryExists(t *testing.T) {
	tempDir := t.TempDir()

	// Test creating nested directories
	nestedPath := tempDir + "/nested/deep/path/file.db"
	err := ensureDirectoryExists(nestedPath)
	if err != nil {
		t.Fatalf("ensureDirectoryExists failed: %v", err)
	}

	// Verify directory was created
	dir := tempDir + "/nested/deep/path"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("Expected directory to be created")
	}

	// Test with file in current directory (no directory component)
	err = ensureDirectoryExists("file.db")
	if err != nil {
		t.Fatalf("ensureDirectoryExists failed for current directory: %v", err)
	}

	// Test with existing directory
	err = ensureDirectoryExists(nestedPath)
	if err != nil {
		t.Fatalf("ensureDirectoryExists failed for existing directory: %v", err)
	}
}

// TestConvertIssueToDBIssue tests the ConvertIssueToDBIssue function
func TestConvertIssueToDBIssue(t *testing.T) {
	// Test with complete issue
	issue := &Issue{
		ID:        123,
		Number:    456,
		Title:     "Test Issue",
		Body:      "Test body",
		State:     "open",
		CreatedAt: "2025-01-01T00:00:00Z",
		UpdatedAt: "2025-01-02T00:00:00Z",
		ClosedAt:  "2025-01-03T00:00:00Z",
		Labels: []struct {
			Name string `json:"name"`
		}{
			{Name: "bug"},
			{Name: "enhancement"},
			{Name: "high-priority"},
		},
		Assignees: []struct {
			Login string `json:"login"`
		}{
			{Login: "user1"},
			{Login: "user2"},
		},
	}

	dbIssue := ConvertIssueToDBIssue(issue)

	// Verify all fields are converted correctly
	if dbIssue.ID != 123 {
		t.Errorf("Expected ID 123, got %d", dbIssue.ID)
	}
	if dbIssue.Number != 456 {
		t.Errorf("Expected Number 456, got %d", dbIssue.Number)
	}
	if dbIssue.Title != "Test Issue" {
		t.Errorf("Expected title 'Test Issue', got '%s'", dbIssue.Title)
	}
	if dbIssue.Body != "Test body" {
		t.Errorf("Expected body 'Test body', got '%s'", dbIssue.Body)
	}
	if dbIssue.State != "open" {
		t.Errorf("Expected state 'open', got '%s'", dbIssue.State)
	}
	if dbIssue.Labels != "bug,enhancement,high-priority" {
		t.Errorf("Expected labels 'bug,enhancement,high-priority', got '%s'", dbIssue.Labels)
	}
	if dbIssue.Assignees != "user1,user2" {
		t.Errorf("Expected assignees 'user1,user2', got '%s'", dbIssue.Assignees)
	}
	if dbIssue.CreatedAt != "2025-01-01T00:00:00Z" {
		t.Errorf("Expected CreatedAt '2025-01-01T00:00:00Z', got '%s'", dbIssue.CreatedAt)
	}
	if dbIssue.UpdatedAt != "2025-01-02T00:00:00Z" {
		t.Errorf("Expected UpdatedAt '2025-01-02T00:00:00Z', got '%s'", dbIssue.UpdatedAt)
	}
	if dbIssue.ClosedAt != "2025-01-03T00:00:00Z" {
		t.Errorf("Expected ClosedAt '2025-01-03T00:00:00Z', got '%s'", dbIssue.ClosedAt)
	}
}

// TestConvertIssueToDBIssue_EmptyLabelsAndAssignees tests conversion with empty labels and assignees
func TestConvertIssueToDBIssue_EmptyLabelsAndAssignees(t *testing.T) {
	issue := &Issue{
		ID:        789,
		Number:    101,
		Title:     "Simple Issue",
		Body:      "",
		State:     "closed",
		CreatedAt: "2025-01-01T00:00:00Z",
		UpdatedAt: "2025-01-01T00:00:00Z",
		ClosedAt:  "",
		Labels: []struct {
			Name string `json:"name"`
		}{},
		Assignees: []struct {
			Login string `json:"login"`
		}{},
	}

	dbIssue := ConvertIssueToDBIssue(issue)

	if dbIssue.Labels != "" {
		t.Errorf("Expected empty labels, got '%s'", dbIssue.Labels)
	}
	if dbIssue.Assignees != "" {
		t.Errorf("Expected empty assignees, got '%s'", dbIssue.Assignees)
	}
}

// TestConvertIssueToDBIssue_SingleLabelAndAssignee tests conversion with single label and assignee
func TestConvertIssueToDBIssue_SingleLabelAndAssignee(t *testing.T) {
	issue := &Issue{
		ID:     1,
		Number: 1,
		Title:  "Single",
		Labels: []struct {
			Name string `json:"name"`
		}{
			{Name: "solo-label"},
		},
		Assignees: []struct {
			Login string `json:"login"`
		}{
			{Login: "solo-user"},
		},
	}

	dbIssue := ConvertIssueToDBIssue(issue)

	if dbIssue.Labels != "solo-label" {
		t.Errorf("Expected labels 'solo-label', got '%s'", dbIssue.Labels)
	}
	if dbIssue.Assignees != "solo-user" {
		t.Errorf("Expected assignees 'solo-user', got '%s'", dbIssue.Assignees)
	}
}
