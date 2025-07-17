//nolint:all
package file

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/XanderD99/disruptor/pkg/db"
)

// Test models
type TestGuild struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Interval  int       `json:"interval"`
	Chance    float64   `json:"chance"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
}

type TestUser struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	GuildID string `json:"guild_id"`
	Age     int    `json:"age"`
}

func setupTestDB(t *testing.T) (db.Database, string, func()) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "filedb_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create database
	database := New(Config{BaseDirectory: tempDir})

	// Connect
	ctx := context.Background()
	if err := database.Connect(ctx); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		database.Disconnect()
		os.RemoveAll(tempDir)
	}

	return database, tempDir, cleanup
}

func TestFileDatabase_Connect(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "filedb_connect_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	database := New(Config{BaseDirectory: tempDir})
	ctx := context.Background()

	err = database.Connect(ctx)
	if err != nil {
		t.Errorf("Connect() error = %v, want nil", err)
	}

	// Check if directory was created
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Errorf("Base directory was not created")
	}
}

func TestFileDatabase_Create(t *testing.T) {
	database, tempDir, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	guild := TestGuild{
		ID:        "123",
		Name:      "Test Guild",
		Interval:  30,
		Chance:    0.5,
		Enabled:   true,
		CreatedAt: time.Now(),
	}

	// Test successful create
	err := database.Create(ctx, "guilds", guild)
	if err != nil {
		t.Errorf("Create() error = %v, want nil", err)
	}

	// Verify file was created
	filePath := filepath.Join(tempDir, "guilds.json")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("Table file was not created")
	}

	// Verify file content
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read table file: %v", err)
	}

	var records map[string]json.RawMessage
	if err := json.Unmarshal(data, &records); err != nil {
		t.Fatalf("Failed to parse table file: %v", err)
	}

	if _, exists := records["123"]; !exists {
		t.Errorf("Record with ID 123 not found in file")
	}

	// Test duplicate create (should fail)
	err = database.Create(ctx, "guilds", guild)
	if err == nil {
		t.Errorf("Create() duplicate should fail, but got nil error")
	}
}

func TestFileDatabase_FindByID(t *testing.T) {
	database, _, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	guild := TestGuild{
		ID:       "123",
		Name:     "Test Guild",
		Interval: 30,
		Chance:   0.5,
		Enabled:  true,
	}

	// Create record first
	err := database.Create(ctx, "guilds", guild)
	if err != nil {
		t.Fatalf("Failed to create test record: %v", err)
	}

	// Test successful find
	result, err := database.FindByID(ctx, "guilds", "123")
	if err != nil {
		t.Errorf("FindByID() error = %v, want nil", err)
	}

	// Verify result
	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("FindByID() result is not map[string]any, got %T", result)
	}

	if resultMap["id"] != "123" {
		t.Errorf("FindByID() id = %v, want 123", resultMap["id"])
	}

	if resultMap["name"] != "Test Guild" {
		t.Errorf("FindByID() name = %v, want Test Guild", resultMap["name"])
	}

	// Test not found
	_, err = database.FindByID(ctx, "guilds", "999")
	if err == nil {
		t.Errorf("FindByID() should fail for non-existent ID, but got nil error")
	}
}

func TestFileDatabase_FindAll(t *testing.T) {
	database, _, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create test data
	guilds := []TestGuild{
		{ID: "1", Name: "Alpha Guild", Interval: 10, Chance: 0.1, Enabled: true},
		{ID: "2", Name: "Beta Guild", Interval: 20, Chance: 0.3, Enabled: false},
		{ID: "3", Name: "Gamma Guild", Interval: 30, Chance: 0.5, Enabled: true},
		{ID: "4", Name: "Delta Guild", Interval: 40, Chance: 0.7, Enabled: true},
	}

	for _, guild := range guilds {
		err := database.Create(ctx, "guilds", guild)
		if err != nil {
			t.Fatalf("Failed to create test guild %s: %v", guild.ID, err)
		}
	}

	tests := []struct {
		name    string
		opts    []db.FindOption
		wantLen int
		checkFn func([]any) bool
	}{
		{
			name:    "find all",
			opts:    []db.FindOption{},
			wantLen: 4,
			checkFn: func(results []any) bool { return len(results) == 4 },
		},
		{
			name:    "find enabled only",
			opts:    []db.FindOption{db.WithFilters(map[string]any{"enabled": true})},
			wantLen: 3,
			checkFn: func(results []any) bool {
				for _, result := range results {
					if record := result.(map[string]any); !record["enabled"].(bool) {
						return false
					}
				}
				return true
			},
		},
		{
			name:    "find with limit",
			opts:    []db.FindOption{db.WithLimit(2)},
			wantLen: 2,
			checkFn: func(results []any) bool { return len(results) == 2 },
		},
		{
			name:    "find with offset",
			opts:    []db.FindOption{db.WithOffset(2)},
			wantLen: 2,
			checkFn: func(results []any) bool { return len(results) == 2 },
		},
		{
			name:    "find with sort ascending",
			opts:    []db.FindOption{db.WithSort("name", db.SortAscending)},
			wantLen: 4,
			checkFn: func(results []any) bool {
				if len(results) < 2 {
					return false
				}
				first := results[0].(map[string]any)["name"].(string)
				return first == "Alpha Guild"
			},
		},
		{
			name:    "find with sort descending",
			opts:    []db.FindOption{db.WithSort("name", db.SortDescending)},
			wantLen: 4,
			checkFn: func(results []any) bool {
				if len(results) < 2 {
					return false
				}
				first := results[0].(map[string]any)["name"].(string)
				return first == "Gamma Guild"
			},
		},
		{
			name: "find with range filter",
			opts: []db.FindOption{
				db.WithFilters(map[string]any{
					"chance": map[string]any{"$gt": 0.2},
				}),
			},
			wantLen: 3,
			checkFn: func(results []any) bool {
				for _, result := range results {
					chance := result.(map[string]any)["chance"].(float64)
					if chance <= 0.2 {
						return false
					}
				}
				return true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := database.FindAll(ctx, "guilds", tt.opts...)
			if err != nil {
				t.Errorf("FindAll() error = %v, want nil", err)
				return
			}

			if len(results) != tt.wantLen {
				t.Errorf("FindAll() length = %d, want %d", len(results), tt.wantLen)
				return
			}

			if tt.checkFn != nil && !tt.checkFn(results) {
				t.Errorf("FindAll() check function failed")
			}
		})
	}
}

func TestFileDatabase_Update(t *testing.T) {
	database, _, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	guild := TestGuild{
		ID:       "123",
		Name:     "Original Name",
		Interval: 30,
		Chance:   0.5,
		Enabled:  true,
	}

	// Create original record
	err := database.Create(ctx, "guilds", guild)
	if err != nil {
		t.Fatalf("Failed to create test record: %v", err)
	}

	// Update record
	guild.Name = "Updated Name"
	guild.Interval = 60
	err = database.Update(ctx, "guilds", guild)
	if err != nil {
		t.Errorf("Update() error = %v, want nil", err)
	}

	// Verify update
	result, err := database.FindByID(ctx, "guilds", "123")
	if err != nil {
		t.Fatalf("Failed to find updated record: %v", err)
	}

	resultMap := result.(map[string]any)
	if resultMap["name"] != "Updated Name" {
		t.Errorf("Update() name = %v, want Updated Name", resultMap["name"])
	}

	if int(resultMap["interval"].(float64)) != 60 {
		t.Errorf("Update() interval = %v, want 60", resultMap["interval"])
	}

	// Test update non-existent record
	nonExistent := TestGuild{ID: "999", Name: "Non-existent"}
	err = database.Update(ctx, "guilds", nonExistent)
	if err == nil {
		t.Errorf("Update() should fail for non-existent record, but got nil error")
	}
}

func TestFileDatabase_Upsert(t *testing.T) {
	database, _, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	guild := TestGuild{
		ID:       "123",
		Name:     "Test Guild",
		Interval: 30,
		Chance:   0.5,
		Enabled:  true,
	}

	// Test upsert (insert)
	err := database.Upsert(ctx, "guilds", guild)
	if err != nil {
		t.Errorf("Upsert() insert error = %v, want nil", err)
	}

	// Verify insert
	result, err := database.FindByID(ctx, "guilds", "123")
	if err != nil {
		t.Fatalf("Failed to find upserted record: %v", err)
	}

	resultMap := result.(map[string]any)
	if resultMap["name"] != "Test Guild" {
		t.Errorf("Upsert() insert name = %v, want Test Guild", resultMap["name"])
	}

	// Test upsert (update)
	guild.Name = "Updated Guild"
	err = database.Upsert(ctx, "guilds", guild)
	if err != nil {
		t.Errorf("Upsert() update error = %v, want nil", err)
	}

	// Verify update
	result, err = database.FindByID(ctx, "guilds", "123")
	if err != nil {
		t.Fatalf("Failed to find updated record: %v", err)
	}

	resultMap = result.(map[string]any)
	if resultMap["name"] != "Updated Guild" {
		t.Errorf("Upsert() update name = %v, want Updated Guild", resultMap["name"])
	}
}

func TestFileDatabase_Delete(t *testing.T) {
	database, _, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	guild := TestGuild{
		ID:       "123",
		Name:     "Test Guild",
		Interval: 30,
		Chance:   0.5,
		Enabled:  true,
	}

	// Create record
	err := database.Create(ctx, "guilds", guild)
	if err != nil {
		t.Fatalf("Failed to create test record: %v", err)
	}

	// Delete record
	err = database.Delete(ctx, "guilds", "123")
	if err != nil {
		t.Errorf("Delete() error = %v, want nil", err)
	}

	// Verify deletion
	_, err = database.FindByID(ctx, "guilds", "123")
	if err == nil {
		t.Errorf("Delete() record should be deleted, but FindByID succeeded")
	}

	// Test delete non-existent record
	err = database.Delete(ctx, "guilds", "999")
	if err == nil {
		t.Errorf("Delete() should fail for non-existent record, but got nil error")
	}
}

func TestFileDatabase_Count(t *testing.T) {
	database, _, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create test data
	guilds := []TestGuild{
		{ID: "1", Name: "Guild 1", Enabled: true, Chance: 0.1},
		{ID: "2", Name: "Guild 2", Enabled: false, Chance: 0.3},
		{ID: "3", Name: "Guild 3", Enabled: true, Chance: 0.5},
	}

	for _, guild := range guilds {
		err := database.Create(ctx, "guilds", guild)
		if err != nil {
			t.Fatalf("Failed to create test guild: %v", err)
		}
	}

	tests := []struct {
		name      string
		opts      []db.FindOption
		wantCount int64
	}{
		{
			name:      "count all",
			opts:      []db.FindOption{},
			wantCount: 3,
		},
		{
			name:      "count enabled",
			opts:      []db.FindOption{db.WithFilters(map[string]any{"enabled": true})},
			wantCount: 2,
		},
		{
			name: "count with range",
			opts: []db.FindOption{
				db.WithFilters(map[string]any{
					"chance": map[string]any{"$gt": 0.2},
				}),
			},
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := database.Count(ctx, "guilds", tt.opts...)
			if err != nil {
				t.Errorf("Count() error = %v, want nil", err)
				return
			}

			if count != tt.wantCount {
				t.Errorf("Count() = %d, want %d", count, tt.wantCount)
			}
		})
	}
}

func TestFileDatabase_MultipleTables(t *testing.T) {
	database, _, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create records in different tables
	guild := TestGuild{ID: "1", Name: "Test Guild"}
	user := TestUser{ID: "1", Name: "Test User", GuildID: "1"}

	err := database.Create(ctx, "guilds", guild)
	if err != nil {
		t.Fatalf("Failed to create guild: %v", err)
	}

	err = database.Create(ctx, "users", user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Verify both tables exist independently
	guildResult, err := database.FindByID(ctx, "guilds", "1")
	if err != nil {
		t.Errorf("Failed to find guild: %v", err)
	}

	userResult, err := database.FindByID(ctx, "users", "1")
	if err != nil {
		t.Errorf("Failed to find user: %v", err)
	}

	// Verify data integrity
	guildMap := guildResult.(map[string]any)
	userMap := userResult.(map[string]any)

	if guildMap["name"] != "Test Guild" {
		t.Errorf("Guild name = %v, want Test Guild", guildMap["name"])
	}

	if userMap["name"] != "Test User" {
		t.Errorf("User name = %v, want Test User", userMap["name"])
	}
}

func TestFileDatabase_Persistence(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "filedb_persistence_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	ctx := context.Background()
	guild := TestGuild{
		ID:       "123",
		Name:     "Persistent Guild",
		Interval: 30,
		Chance:   0.5,
		Enabled:  true,
	}

	// Create first database instance
	db1 := New(Config{BaseDirectory: tempDir})
	err = db1.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect db1: %v", err)
	}

	// Create record
	err = db1.Create(ctx, "guilds", guild)
	if err != nil {
		t.Fatalf("Failed to create record: %v", err)
	}

	// Disconnect first instance
	err = db1.Disconnect()
	if err != nil {
		t.Fatalf("Failed to disconnect db1: %v", err)
	}

	// Create second database instance
	db2 := New(Config{BaseDirectory: tempDir})
	err = db2.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect db2: %v", err)
	}
	defer db2.Disconnect()

	// Try to find the record
	result, err := db2.FindByID(ctx, "guilds", "123")
	if err != nil {
		t.Errorf("Failed to find persisted record: %v", err)
	}

	resultMap := result.(map[string]any)
	if resultMap["name"] != "Persistent Guild" {
		t.Errorf("Persisted record name = %v, want Persistent Guild", resultMap["name"])
	}
}

func TestFileDatabase_ConcurrentAccess(t *testing.T) {
	database, _, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	numGoroutines := 10
	numRecords := 5

	// Channel to collect errors
	errChan := make(chan error, numGoroutines*numRecords)

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			for j := 0; j < numRecords; j++ {
				guild := TestGuild{
					ID:       fmt.Sprintf("%d-%d", goroutineID, j),
					Name:     fmt.Sprintf("Guild %d-%d", goroutineID, j),
					Interval: 30,
					Chance:   0.5,
					Enabled:  true,
				}

				err := database.Create(ctx, "guilds", guild)
				if err != nil {
					errChan <- err
					return
				}
			}
		}(i)
	}

	// Wait a bit for goroutines to complete
	time.Sleep(100 * time.Millisecond)

	// Check for errors
	close(errChan)
	for err := range errChan {
		t.Errorf("Concurrent access error: %v", err)
	}

	// Verify all records were created
	results, err := database.FindAll(ctx, "guilds")
	if err != nil {
		t.Fatalf("Failed to find all records: %v", err)
	}

	expectedCount := numGoroutines * numRecords
	if len(results) != expectedCount {
		t.Errorf("Expected %d records, got %d", expectedCount, len(results))
	}
}

func TestFileDatabase_FilterOperators(t *testing.T) {
	database, _, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create test data with various values
	users := []TestUser{
		{ID: "1", Name: "Alice", Age: 25, GuildID: "guild1"},
		{ID: "2", Name: "Bob", Age: 30, GuildID: "guild1"},
		{ID: "3", Name: "Charlie", Age: 35, GuildID: "guild2"},
		{ID: "4", Name: "David", Age: 20, GuildID: "guild2"},
	}

	for _, user := range users {
		err := database.Create(ctx, "users", user)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}

	tests := []struct {
		name      string
		filters   map[string]any
		wantCount int
		checkFn   func([]any) bool
	}{
		{
			name:      "greater than",
			filters:   map[string]any{"age": map[string]any{"$gt": 25}},
			wantCount: 2,
			checkFn: func(results []any) bool {
				for _, result := range results {
					age := result.(map[string]any)["age"].(float64)
					if age <= 25 {
						return false
					}
				}
				return true
			},
		},
		{
			name:      "less than or equal",
			filters:   map[string]any{"age": map[string]any{"$lte": 30}},
			wantCount: 3,
			checkFn: func(results []any) bool {
				for _, result := range results {
					age := result.(map[string]any)["age"].(float64)
					if age > 30 {
						return false
					}
				}
				return true
			},
		},
		{
			name:      "not equal",
			filters:   map[string]any{"name": map[string]any{"$ne": "Alice"}},
			wantCount: 3,
			checkFn: func(results []any) bool {
				for _, result := range results {
					name := result.(map[string]any)["name"].(string)
					if name == "Alice" {
						return false
					}
				}
				return true
			},
		},
		{
			name:      "in array",
			filters:   map[string]any{"guild_id": map[string]any{"$in": []any{"guild1"}}},
			wantCount: 2,
			checkFn: func(results []any) bool {
				for _, result := range results {
					guildID := result.(map[string]any)["guild_id"].(string)
					if guildID != "guild1" {
						return false
					}
				}
				return true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := database.FindAll(ctx, "users", db.WithFilters(tt.filters))
			if err != nil {
				t.Errorf("FindAll() error = %v, want nil", err)
				return
			}

			if len(results) != tt.wantCount {
				t.Errorf("FindAll() count = %d, want %d", len(results), tt.wantCount)
				return
			}

			if tt.checkFn != nil && !tt.checkFn(results) {
				t.Errorf("FindAll() check function failed for filter: %v", tt.filters)
			}
		})
	}
}

func TestFileDatabase_InvalidData(t *testing.T) {
	database, _, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Test entity without ID field
	type InvalidEntity struct {
		Name string `json:"name"`
	}

	invalidEntity := InvalidEntity{Name: "No ID"}
	err := database.Create(ctx, "invalid", invalidEntity)
	if err == nil {
		t.Errorf("Create() should fail for entity without ID field, but got nil error")
	}
}
