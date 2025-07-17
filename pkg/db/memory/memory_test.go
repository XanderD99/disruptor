//nolint:all
package memory

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/XanderD99/disruptor/pkg/db"
)

// Test entities
type TestUser struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Email  string   `json:"email"`
	Age    int      `json:"age"`
	Active bool     `json:"active"`
	Score  float64  `json:"score"`
	Tags   []string `json:"tags"`
}

type TestProduct struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
	InStock     bool    `json:"in_stock"`
	Description string  `json:"description"`
}

type TestOrder struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	ProductID int       `json:"product_id"`
	Quantity  int       `json:"quantity"`
	Total     float64   `json:"total"`
	CreatedAt time.Time `json:"created_at"`
}

func TestMemoryDatabase_New(t *testing.T) {
	db := New()
	memDB, ok := db.(*MemoryDatabase)
	if !ok {
		t.Fatal("New() should return a *MemoryDatabase")
	}

	if memDB.tables == nil {
		t.Error("tables map should be initialized")
	}

	if len(memDB.tables) != 0 {
		t.Error("tables map should be empty initially")
	}
}

func TestMemoryDatabase_Connect(t *testing.T) {
	db := New()
	ctx := context.Background()

	err := db.Connect(ctx)
	if err != nil {
		t.Errorf("Connect() should not return an error for memory database, got: %v", err)
	}
}

func TestMemoryDatabase_Disconnect(t *testing.T) {
	db := New()
	memDB := db.(*MemoryDatabase)

	// Add some data first
	ctx := context.Background()
	user := &TestUser{ID: "1", Name: "John", Email: "john@example.com", Age: 30}
	err := db.Create(ctx, "users", user)
	if err != nil {
		t.Fatalf("Failed to create test data: %v", err)
	}

	// Verify data exists
	if len(memDB.tables) == 0 {
		t.Error("Expected tables to exist before disconnect")
	}

	// Disconnect should clear all data
	err = db.Disconnect()
	if err != nil {
		t.Errorf("Disconnect() should not return an error, got: %v", err)
	}

	if len(memDB.tables) != 0 {
		t.Error("Disconnect() should clear all tables")
	}
}

func TestMemoryDatabase_Create(t *testing.T) {
	tests := []struct {
		name        string
		table       string
		entity      any
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid user creation",
			table:       "users",
			entity:      &TestUser{ID: "1", Name: "John", Email: "john@example.com", Age: 30},
			expectError: false,
		},
		{
			name:        "valid product creation",
			table:       "products",
			entity:      &TestProduct{ID: 1, Name: "Widget", Price: 9.99, Category: "Tools"},
			expectError: false,
		},
		{
			name:        "duplicate ID should fail",
			table:       "users",
			entity:      &TestUser{ID: "1", Name: "Jane", Email: "jane@example.com", Age: 25},
			expectError: true,
			errorMsg:    "already exists",
		},
		{
			name:        "entity without ID field should fail",
			table:       "invalid",
			entity:      struct{ Name string }{Name: "test"},
			expectError: true,
			errorMsg:    "ID field not found",
		},
	}

	db := New()
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := db.Create(ctx, tt.table, tt.entity)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestMemoryDatabase_FindByID(t *testing.T) {
	db := New()
	ctx := context.Background()

	// Create test data
	user := &TestUser{ID: "1", Name: "John", Email: "john@example.com", Age: 30}
	err := db.Create(ctx, "users", user)
	if err != nil {
		t.Fatalf("Failed to create test data: %v", err)
	}

	tests := []struct {
		name        string
		table       string
		id          any
		expectError bool
		expectNil   bool
	}{
		{
			name:        "find existing user",
			table:       "users",
			id:          "1",
			expectError: false,
		},
		{
			name:        "find non-existent user",
			table:       "users",
			id:          "999",
			expectError: true,
		},
		{
			name:        "find from empty table",
			table:       "empty",
			id:          "1",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := db.FindByID(ctx, tt.table, tt.id)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if result != nil {
					t.Error("Expected nil result on error")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result == nil {
					t.Error("Expected non-nil result")
				}

				// Verify the returned user matches what we stored
				if foundUser, ok := result.(*TestUser); ok {
					if foundUser.ID != user.ID || foundUser.Name != user.Name {
						t.Errorf("Found user doesn't match: expected %+v, got %+v", user, foundUser)
					}
				} else {
					t.Error("Result is not a *TestUser")
				}
			}
		})
	}
}

func TestMemoryDatabase_FindAll(t *testing.T) {
	memoryDB := New()
	ctx := context.Background()

	// Create test data
	users := []*TestUser{
		{ID: "1", Name: "John", Email: "john@example.com", Age: 30, Active: true, Score: 85.5},
		{ID: "2", Name: "Jane", Email: "jane@example.com", Age: 25, Active: false, Score: 92.0},
		{ID: "3", Name: "Bob", Email: "bob@example.com", Age: 35, Active: true, Score: 78.2},
	}

	for _, user := range users {
		err := memoryDB.Create(ctx, "users", user)
		if err != nil {
			t.Fatalf("Failed to create test data: %v", err)
		}
	}

	tests := []struct {
		name        string
		table       string
		options     []db.FindOption
		expectCount int
		expectError bool
	}{
		{
			name:        "find all users",
			table:       "users",
			options:     nil,
			expectCount: 3,
		},
		{
			name:        "find with filter - active users",
			table:       "users",
			options:     []db.FindOption{db.WithFilter("Active", true)},
			expectCount: 2,
		},
		{
			name:        "find with filter - age greater than 30",
			table:       "users",
			options:     []db.FindOption{db.WithFilter("Age", map[string]any{"$gt": 30})},
			expectCount: 1,
		},
		{
			name:        "find with filter - score range",
			table:       "users",
			options:     []db.FindOption{db.WithFilter("Score", map[string]any{"$gte": 80})},
			expectCount: 2,
		},
		{
			name:        "find with limit",
			table:       "users",
			options:     []db.FindOption{db.WithLimit(2)},
			expectCount: 2,
		},
		{
			name:        "find with offset",
			table:       "users",
			options:     []db.FindOption{db.WithOffset(1), db.WithLimit(2)},
			expectCount: 2,
		},
		{
			name:        "find with sort by age ascending",
			table:       "users",
			options:     []db.FindOption{db.WithSort("Age", db.SortAscending)},
			expectCount: 3,
		},
		{
			name:        "find with sort by score descending",
			table:       "users",
			options:     []db.FindOption{db.WithSort("Score", db.SortDescending)},
			expectCount: 3,
		},
		{
			name:        "find from empty table",
			table:       "empty",
			options:     nil,
			expectCount: 0,
		},
		{
			name:  "find with multiple filters",
			table: "users",
			options: []db.FindOption{
				db.WithFilter("Active", true),
				db.WithFilter("Age", map[string]any{"$gte": 30}),
			},
			expectCount: 2, // John (30, true) and Bob (35, true)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := memoryDB.FindAll(ctx, tt.table, tt.options...)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(results) != tt.expectCount {
				t.Errorf("Expected %d results, got %d", tt.expectCount, len(results))
			}
		})
	}
}

func TestMemoryDatabase_Update(t *testing.T) {
	db := New()
	ctx := context.Background()

	// Create initial data
	user := &TestUser{ID: "1", Name: "John", Email: "john@example.com", Age: 30}
	err := db.Create(ctx, "users", user)
	if err != nil {
		t.Fatalf("Failed to create test data: %v", err)
	}

	tests := []struct {
		name        string
		table       string
		entity      any
		expectError bool
		errorMsg    string
	}{
		{
			name:        "update existing user",
			table:       "users",
			entity:      &TestUser{ID: "1", Name: "John Updated", Email: "john.updated@example.com", Age: 31},
			expectError: false,
		},
		{
			name:        "update non-existent user",
			table:       "users",
			entity:      &TestUser{ID: "999", Name: "Ghost", Email: "ghost@example.com", Age: 0},
			expectError: true,
			errorMsg:    "not found",
		},
		{
			name:        "update entity without ID",
			table:       "users",
			entity:      struct{ Name string }{Name: "test"},
			expectError: true,
			errorMsg:    "ID field not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := db.Update(ctx, tt.table, tt.entity)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Verify the update worked
				if updatedUser, ok := tt.entity.(*TestUser); ok {
					found, err := db.FindByID(ctx, tt.table, updatedUser.ID)
					if err != nil {
						t.Errorf("Failed to find updated user: %v", err)
					} else if foundUser, ok := found.(*TestUser); ok {
						if foundUser.Name != updatedUser.Name || foundUser.Email != updatedUser.Email {
							t.Errorf("User not properly updated: expected %+v, got %+v", updatedUser, foundUser)
						}
					}
				}
			}
		})
	}
}

func TestMemoryDatabase_Upsert(t *testing.T) {
	db := New()
	ctx := context.Background()

	// Create initial data
	user := &TestUser{ID: "1", Name: "John", Email: "john@example.com", Age: 30}
	err := db.Create(ctx, "users", user)
	if err != nil {
		t.Fatalf("Failed to create test data: %v", err)
	}

	tests := []struct {
		name        string
		table       string
		entity      any
		expectError bool
		isUpdate    bool
	}{
		{
			name:        "upsert existing user (update)",
			table:       "users",
			entity:      &TestUser{ID: "1", Name: "John Updated", Email: "john.updated@example.com", Age: 31},
			expectError: false,
			isUpdate:    true,
		},
		{
			name:        "upsert new user (insert)",
			table:       "users",
			entity:      &TestUser{ID: "2", Name: "Jane", Email: "jane@example.com", Age: 25},
			expectError: false,
			isUpdate:    false,
		},
		{
			name:        "upsert entity without ID",
			table:       "users",
			entity:      struct{ Name string }{Name: "test"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := db.Upsert(ctx, tt.table, tt.entity)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Verify the upsert worked
				if testUser, ok := tt.entity.(*TestUser); ok {
					found, err := db.FindByID(ctx, tt.table, testUser.ID)
					if err != nil {
						t.Errorf("Failed to find upserted user: %v", err)
					} else if foundUser, ok := found.(*TestUser); ok {
						if foundUser.Name != testUser.Name || foundUser.Email != testUser.Email {
							t.Errorf("User not properly upserted: expected %+v, got %+v", testUser, foundUser)
						}
					}
				}
			}
		})
	}
}

func TestMemoryDatabase_Delete(t *testing.T) {
	db := New()
	ctx := context.Background()

	// Create test data
	users := []*TestUser{
		{ID: "1", Name: "John", Email: "john@example.com", Age: 30},
		{ID: "2", Name: "Jane", Email: "jane@example.com", Age: 25},
	}

	for _, user := range users {
		err := db.Create(ctx, "users", user)
		if err != nil {
			t.Fatalf("Failed to create test data: %v", err)
		}
	}

	tests := []struct {
		name        string
		table       string
		id          any
		expectError bool
		errorMsg    string
	}{
		{
			name:        "delete existing user",
			table:       "users",
			id:          "1",
			expectError: false,
		},
		{
			name:        "delete non-existent user",
			table:       "users",
			id:          "999",
			expectError: true,
			errorMsg:    "not found",
		},
		{
			name:        "delete from empty table",
			table:       "empty",
			id:          "1",
			expectError: true,
			errorMsg:    "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := db.Delete(ctx, tt.table, tt.id)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Verify the record was deleted
				_, err := db.FindByID(ctx, tt.table, tt.id)
				if err == nil {
					t.Error("Expected record to be deleted")
				}
			}
		})
	}
}

func TestMemoryDatabase_Count(t *testing.T) {
	memoryDB := New()
	ctx := context.Background()

	// Create test data
	users := []*TestUser{
		{ID: "1", Name: "John", Email: "john@example.com", Age: 30, Active: true},
		{ID: "2", Name: "Jane", Email: "jane@example.com", Age: 25, Active: false},
		{ID: "3", Name: "Bob", Email: "bob@example.com", Age: 35, Active: true},
	}

	for _, user := range users {
		err := memoryDB.Create(ctx, "users", user)
		if err != nil {
			t.Fatalf("Failed to create test data: %v", err)
		}
	}

	tests := []struct {
		name          string
		table         string
		options       []db.FindOption
		expectedCount int64
	}{
		{
			name:          "count all users",
			table:         "users",
			options:       nil,
			expectedCount: 3,
		},
		{
			name:          "count active users",
			table:         "users",
			options:       []db.FindOption{db.WithFilter("Active", true)},
			expectedCount: 2,
		},
		{
			name:          "count users over 30",
			table:         "users",
			options:       []db.FindOption{db.WithFilter("Age", map[string]any{"$gt": 30})},
			expectedCount: 1,
		},
		{
			name:          "count from empty table",
			table:         "empty",
			options:       nil,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := memoryDB.Count(ctx, tt.table, tt.options...)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if count != tt.expectedCount {
				t.Errorf("Expected count %d, got %d", tt.expectedCount, count)
			}
		})
	}
}

func TestMemoryDatabase_ConcurrentAccess(t *testing.T) {
	db := New()
	ctx := context.Background()

	const numGoroutines = 100
	const recordsPerGoroutine = 10

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*recordsPerGoroutine)

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < recordsPerGoroutine; j++ {
				id := fmt.Sprintf("user_%d_%d", goroutineID, j)
				user := &TestUser{
					ID:    id,
					Name:  fmt.Sprintf("User %d-%d", goroutineID, j),
					Email: fmt.Sprintf("user%d_%d@example.com", goroutineID, j),
					Age:   20 + (goroutineID+j)%50,
				}
				if err := db.Create(ctx, "users", user); err != nil {
					errors <- err
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent write error: %v", err)
	}

	// Verify all records were created
	count, err := db.Count(ctx, "users")
	if err != nil {
		t.Fatalf("Failed to count records: %v", err)
	}

	expectedCount := int64(numGoroutines * recordsPerGoroutine)
	if count != expectedCount {
		t.Errorf("Expected %d records, got %d", expectedCount, count)
	}

	// Concurrent reads
	readErrors := make(chan error, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			results, err := db.FindAll(ctx, "users")
			if err != nil {
				readErrors <- err
				return
			}
			if len(results) != int(expectedCount) {
				readErrors <- fmt.Errorf("goroutine %d: expected %d results, got %d", goroutineID, expectedCount, len(results))
			}
		}(i)
	}

	wg.Wait()
	close(readErrors)

	// Check for read errors
	for err := range readErrors {
		t.Errorf("Concurrent read error: %v", err)
	}
}

func TestMemoryDatabase_ComplexQueries(t *testing.T) {
	memoryDB := New()
	ctx := context.Background()

	// Create test data with various types
	orders := []*TestOrder{
		{ID: "1", UserID: "user1", ProductID: 1, Quantity: 2, Total: 19.98, CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)},
		{ID: "2", UserID: "user2", ProductID: 2, Quantity: 1, Total: 29.99, CreatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)},
		{ID: "3", UserID: "user1", ProductID: 3, Quantity: 3, Total: 44.97, CreatedAt: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC)},
		{ID: "4", UserID: "user3", ProductID: 1, Quantity: 1, Total: 9.99, CreatedAt: time.Date(2023, 1, 4, 0, 0, 0, 0, time.UTC)},
	}

	for _, order := range orders {
		err := memoryDB.Create(ctx, "orders", order)
		if err != nil {
			t.Fatalf("Failed to create test data: %v", err)
		}
	}

	tests := []struct {
		name        string
		options     []db.FindOption
		expectCount int
		description string
	}{
		{
			name: "orders for user1",
			options: []db.FindOption{
				db.WithFilter("UserID", "user1"),
			},
			expectCount: 2,
			description: "Find orders for specific user",
		},
		{
			name: "high value orders",
			options: []db.FindOption{
				db.WithFilter("Total", map[string]any{"$gte": 30.0}),
			},
			expectCount: 1,
			description: "Find orders with total >= 30",
		},
		{
			name: "recent orders sorted by total",
			options: []db.FindOption{
				db.WithSort("Total", db.SortDescending),
				db.WithLimit(2),
			},
			expectCount: 2,
			description: "Get top 2 orders by total value",
		},
		{
			name: "complex filter with multiple conditions",
			options: []db.FindOption{
				db.WithFilter("Quantity", map[string]any{"$gte": 2}),
				db.WithFilter("Total", map[string]any{"$lte": 50.0}),
			},
			expectCount: 2,
			description: "Orders with quantity >= 2 and total <= 50",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := memoryDB.FindAll(ctx, "orders", tt.options...)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(results) != tt.expectCount {
				t.Errorf("%s: expected %d results, got %d", tt.description, tt.expectCount, len(results))
			}
		})
	}
}

// Helper functions
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			strings.Contains(s, substr))))
}

func verifySorting(t *testing.T, results []any, field string, direction db.SortDirection) {
	if len(results) < 2 {
		return
	}

	for i := 0; i < len(results)-1; i++ {
		vi := reflect.ValueOf(results[i])
		vj := reflect.ValueOf(results[i+1])

		if vi.Kind() == reflect.Ptr {
			vi = vi.Elem()
		}
		if vj.Kind() == reflect.Ptr {
			vj = vj.Elem()
		}

		fieldI := vi.FieldByName(field).Interface()
		fieldJ := vj.FieldByName(field).Interface()

		cmp := compareValues(fieldI, fieldJ)

		if direction == db.SortAscending && cmp > 0 {
			t.Errorf("Results not sorted ascending by %s: %v > %v", field, fieldI, fieldJ)
		} else if direction == db.SortDescending && cmp < 0 {
			t.Errorf("Results not sorted descending by %s: %v < %v", field, fieldI, fieldJ)
		}
	}
}

func compareValues(a, b any) int {
	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)

	// Handle numeric comparisons
	if va.CanFloat() && vb.CanFloat() {
		fa, fb := va.Float(), vb.Float()
		if fa < fb {
			return -1
		} else if fa > fb {
			return 1
		}
		return 0
	}

	if va.CanInt() && vb.CanInt() {
		ia, ib := va.Int(), vb.Int()
		if ia < ib {
			return -1
		} else if ia > ib {
			return 1
		}
		return 0
	}

	// String comparison
	sa, sb := fmt.Sprintf("%v", a), fmt.Sprintf("%v", b)
	if sa < sb {
		return -1
	} else if sa > sb {
		return 1
	}
	return 0
}
