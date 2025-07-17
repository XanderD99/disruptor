//go:build integration

//nolint:all
package mongo

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/XanderD99/disruptor/pkg/db"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
)

func setupMongoContainer() (db.Database, func(), error) {
	ctx := context.Background()

	// Start MongoDB container
	mongoContainer, err := mongodb.Run(ctx,
		"mongo:7.0",
		mongodb.WithUsername("testuser"),
		mongodb.WithPassword("testpass"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start MongoDB container: %w", err)
	}

	// Parse connection details
	endpoint, err := mongoContainer.Endpoint(ctx, "")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get endpoint: %w", err)
	}

	// Create config
	config := Config{
		Hosts:    []string{endpoint},
		Database: "test_disruptor",
		Auth: AuthConfig{
			Enabled:   true,
			Username:  "testuser",
			Password:  "testpass",
			Mechanism: "SCRAM-SHA-256",
			Source:    "admin",
		},
		Pool: PoolConfig{
			MinSize:       2,
			MaxSize:       10,
			MaxConnecting: 5,
			MaxIdleTime:   5 * time.Minute,
		},
		Timeout: TimeoutConfig{
			Connect: 10 * time.Second,
			Query:   30 * time.Second,
		},
	}

	// Create database instance
	database := New(config)

	// Test connection
	if err := database.Connect(ctx); err != nil {
		if err := mongoContainer.Terminate(ctx); err != nil {
			return nil, nil, fmt.Errorf("failed to terminate MongoDB container: %w", err)
		}
		return nil, nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	return database, func() {
		if err := database.Disconnect(); err != nil {
			fmt.Printf("Failed to disconnect from MongoDB: %v\n", err)
		}
		if err := mongoContainer.Terminate(ctx); err != nil {
			fmt.Printf("Failed to terminate MongoDB container: %v\n", err)
		}
	}, nil
}

var (
	sharedTestDB    db.Database
	sharedDBCleanup func()
)

func TestMain(m *testing.M) {
	// 🏃‍♂️ One-time setup
	var err error
	sharedTestDB, sharedDBCleanup, err = setupMongoContainer()
	if err != nil {
		panic("Shared DB setup failed: " + err.Error())
	}

	code := m.Run() // 🧹 Clean up when we're done
	if sharedDBCleanup != nil {
		sharedDBCleanup()
	}

	os.Exit(code)
}

type TestUser struct {
	ID      string        `json:"id" bson:"id"`
	Name    string        `json:"name" bson:"name"`
	Email   string        `json:"email" bson:"email"`
	Age     int           `json:"age" bson:"age"`
	Active  bool          `json:"active" bson:"active"`
	Score   float64       `json:"score" bson:"score"`
	Tags    []string      `json:"tags" bson:"tags"`
	Timeout time.Duration `json:"timeout" bson:"timeout"`
}

func TestMongoConnect(t *testing.T) {
	// Connection should already be established in setup
	// Test that we can perform a simple operation
	ctx := context.Background()

	user := &TestUser{ID: "connect_test", Name: "Connect Test", Email: "test@example.com", Age: 30}
	err := sharedTestDB.Create(ctx, "test_users", user)
	if err != nil {
		t.Errorf("Failed to create test document: %v", err)
	}

	// Clean up
	sharedTestDB.Delete(ctx, "test_users", "connect_test")
}

func TestMongoCreate(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		table       string
		entity      any
		expectError bool
		cleanup     string
	}{
		{
			name:    "create user",
			table:   "users",
			entity:  &TestUser{ID: "user1", Name: "John", Email: "john@example.com", Age: 30, Active: true, Score: 85.5},
			cleanup: "user1",
		},
		{
			name:        "create duplicate should succeed (MongoDB doesn't enforce unique by default)",
			table:       "users",
			entity:      &TestUser{ID: "user1", Name: "Jane", Email: "jane@example.com", Age: 25},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sharedTestDB.Create(ctx, tt.table, tt.entity)

			if tt.expectError {
				assert.Error(t, err, "Expected error but got none")
			} else {
				assert.NoError(t, err, "Unexpected error: %v", err)
			}

			// Cleanup
			if tt.cleanup != "" {
				sharedTestDB.Delete(ctx, tt.table, tt.cleanup)
			}
		})
	}
}

func TestMongoFindByID(t *testing.T) {
	ctx := context.Background()

	// Create test data
	user := &TestUser{ID: "find_test", Name: "Find Test", Email: "find@example.com", Age: 30}
	err := sharedTestDB.Create(ctx, "users", user)
	if err != nil {
		t.Fatalf("Failed to create test data: %v", err)
	}
	defer sharedTestDB.Delete(ctx, "users", "find_test")

	tests := []struct {
		name        string
		table       string
		id          string
		expectError bool
	}{
		{
			name:        "find existing user",
			table:       "users",
			id:          "find_test",
			expectError: false,
		},
		{
			name:        "find non-existent user",
			table:       "users",
			id:          "nonexistent",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result TestUser
			err := sharedTestDB.FindOne(ctx, tt.table, &result, db.WithIDFilter(TestUser{ID: tt.id}))

			if tt.expectError {
				assert.Error(t, err, "Expected error but got none")
			} else {
				assert.NoError(t, err, "Unexpected error: %v", err)
				assert.Equal(t, tt.id, result.ID, "Expected ID to match")
			}
		})
	}
}

func TestMongoFindAll(t *testing.T) {
	ctx := context.Background()

	// Create test data
	users := []*TestUser{
		{ID: "find_all_1", Name: "John", Email: "john@example.com", Age: 30, Active: true, Score: 85.5},
		{ID: "find_all_2", Name: "Jane", Email: "jane@example.com", Age: 25, Active: false, Score: 92.0},
		{ID: "find_all_3", Name: "Bob", Email: "bob@example.com", Age: 35, Active: true, Score: 78.2, Timeout: 5 * time.Second},
	}

	// Insert test data
	for _, user := range users {
		err := sharedTestDB.Create(ctx, "find_all_users", user)
		if err != nil {
			t.Fatalf("Failed to create test data: %v", err)
		}
	}

	// Cleanup
	defer func() {
		for _, user := range users {
			sharedTestDB.Delete(ctx, "find_all_users", user.ID)
		}
	}()

	tests := []struct {
		name        string
		table       string
		options     []db.FindOption
		expectCount int
		description string
	}{
		{
			name:        "find all users",
			table:       "find_all_users",
			options:     nil,
			expectCount: 3,
			description: "Should return all test users",
		},
		{
			name:        "find active users",
			table:       "find_all_users",
			options:     []db.FindOption{db.WithFilter("active", true)},
			expectCount: 2,
			description: "Should return users where active=true",
		},
		{
			name:        "find with limit",
			table:       "find_all_users",
			options:     []db.FindOption{db.WithLimit(2)},
			expectCount: 2,
			description: "Should return limited number of users",
		},
		{
			name:        "find with offset",
			table:       "find_all_users",
			options:     []db.FindOption{db.WithOffset(1), db.WithLimit(2)},
			expectCount: 2,
			description: "Should return users with offset",
		},
		{
			name:        "find with sort by age ascending",
			table:       "find_all_users",
			options:     []db.FindOption{db.WithSort("age", db.SortAscending)},
			expectCount: 3,
			description: "Should return all users sorted by age",
		},
		{
			name:        "find users with timeout of 5 seconds",
			table:       "find_all_users",
			options:     []db.FindOption{db.WithFilters(map[string]any{"timeout": time.Second * 5})},
			expectCount: 1,
			description: "Should return all users with timeout of 5 seconds",
		},
		{
			name:        "find empty table",
			table:       "empty_table",
			options:     nil,
			expectCount: 0,
			description: "Should return empty result for non-existent table",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var results []TestUser
			err := sharedTestDB.Find(ctx, tt.table, &results, tt.options...)
			assert.NoError(t, err, "Unexpected error: %v", err)
			assert.Len(t, results, tt.expectCount, "%s: expected %d results, got %d", tt.description, tt.expectCount, len(results))
		})
	}
}

func TestMongoUpdate(t *testing.T) {
	ctx := context.Background()

	// Create initial data
	user := &TestUser{ID: "update_test", Name: "Original", Email: "original@example.com", Age: 30}
	err := sharedTestDB.Create(ctx, "update_users", user)
	if err != nil {
		t.Fatalf("Failed to create test data: %v", err)
	}
	defer sharedTestDB.Delete(ctx, "update_users", "update_test")

	tests := []struct {
		name        string
		table       string
		entity      any
		expectError bool
	}{
		{
			name:        "update existing user",
			table:       "update_users",
			entity:      &TestUser{ID: "update_test", Name: "Updated", Email: "updated@example.com", Age: 31},
			expectError: false,
		},
		{
			name:        "update non-existent user",
			table:       "update_users",
			entity:      &TestUser{ID: "nonexistent", Name: "Ghost", Email: "ghost@example.com", Age: 0},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sharedTestDB.Update(ctx, tt.table, tt.entity)

			if tt.expectError {
				assert.Error(t, err, "Expected error but got none")
			} else {
				assert.NoError(t, err, "Unexpected error: %v", err)
			}
		})
	}
}

func TestMongoUpsert(t *testing.T) {
	ctx := context.Background()

	// Create initial data for update test
	user := &TestUser{ID: "upsert_test", Name: "Original", Email: "original@example.com", Age: 30}
	err := sharedTestDB.Create(ctx, "upsert_users", user)
	if err != nil {
		t.Fatalf("Failed to create test data: %v", err)
	}
	defer sharedTestDB.Delete(ctx, "upsert_users", "upsert_test")
	defer sharedTestDB.Delete(ctx, "upsert_users", "upsert_new")

	tests := []struct {
		name   string
		table  string
		entity any
		isNew  bool
	}{
		{
			name:   "upsert existing user (update)",
			table:  "upsert_users",
			entity: &TestUser{ID: "upsert_test", Name: "Updated via Upsert", Email: "upsert@example.com", Age: 32},
			isNew:  false,
		},
		{
			name:   "upsert new user (insert)",
			table:  "upsert_users",
			entity: &TestUser{ID: "upsert_new", Name: "New User", Email: "new@example.com", Age: 28},
			isNew:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sharedTestDB.Upsert(ctx, tt.table, tt.entity)
			assert.NoError(t, err, "Unexpected error: %v", err)
		})
	}
}

func TestMongoDelete(t *testing.T) {
	ctx := context.Background()

	// Create test data
	users := []*TestUser{
		{ID: "delete_test_1", Name: "Delete1", Email: "delete1@example.com", Age: 30},
		{ID: "delete_test_2", Name: "Delete2", Email: "delete2@example.com", Age: 25},
	}

	for _, user := range users {
		err := sharedTestDB.Create(ctx, "delete_users", user)
		if err != nil {
			t.Fatalf("Failed to create test data: %v", err)
		}
	}

	tests := []struct {
		name        string
		table       string
		id          string
		expectError bool
	}{
		{
			name:        "delete existing user",
			table:       "delete_users",
			id:          "delete_test_1",
			expectError: false,
		},
		{
			name:        "delete non-existent user (should not error)",
			table:       "delete_users",
			id:          "nonexistent",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sharedTestDB.Delete(ctx, tt.table, tt.id)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Verify deletion for valid IDs
				if tt.id == "delete_test_1" {
					var result any
					err := sharedTestDB.FindOne(ctx, tt.table, &result, db.WithIDFilter(TestUser{ID: tt.id}))
					if err == nil {
						t.Error("Expected error when finding deleted record")
					}
				}
			}
		})
	}

	// Clean up remaining test data
	sharedTestDB.Delete(ctx, "delete_users", "delete_test_2")
}

func TestMongoCount(t *testing.T) {
	ctx := context.Background()

	// Create test data
	users := []*TestUser{
		{ID: "count_1", Name: "Count1", Email: "count1@example.com", Age: 30, Active: true},
		{ID: "count_2", Name: "Count2", Email: "count2@example.com", Age: 25, Active: false},
		{ID: "count_3", Name: "Count3", Email: "count3@example.com", Age: 35, Active: true},
	}

	for _, user := range users {
		err := sharedTestDB.Create(ctx, "count_users", user)
		if err != nil {
			t.Fatalf("Failed to create test data: %v", err)
		}
	}

	defer func() {
		for _, user := range users {
			sharedTestDB.Delete(ctx, "count_users", user.ID)
		}
	}()

	tests := []struct {
		name          string
		table         string
		options       []db.FindOption
		expectedCount int64
	}{
		{
			name:          "count all users",
			table:         "count_users",
			options:       nil,
			expectedCount: 3,
		},
		{
			name:          "count active users",
			table:         "count_users",
			options:       []db.FindOption{db.WithFilter("active", true)},
			expectedCount: 2,
		},
		{
			name:          "count empty table",
			table:         "empty_count_table",
			options:       nil,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := sharedTestDB.Count(ctx, tt.table, tt.options...)
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

func TestMongoComplexQueries(t *testing.T) {
	ctx := context.Background()

	// Create test data with various types
	type TestOrder struct {
		ID        string    `json:"id" bson:"id"`
		UserID    string    `json:"user_id" bson:"user_id"`
		ProductID int       `json:"product_id" bson:"product_id"`
		Quantity  int       `json:"quantity" bson:"quantity"`
		Total     float64   `json:"total" bson:"total"`
		CreatedAt time.Time `json:"created_at" bson:"created_at"`
	}

	orders := []*TestOrder{
		{ID: "order_1", UserID: "user1", ProductID: 1, Quantity: 2, Total: 19.98, CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)},
		{ID: "order_2", UserID: "user2", ProductID: 2, Quantity: 1, Total: 29.99, CreatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)},
		{ID: "order_3", UserID: "user1", ProductID: 3, Quantity: 3, Total: 44.97, CreatedAt: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC)},
		{ID: "order_4", UserID: "user3", ProductID: 1, Quantity: 1, Total: 9.99, CreatedAt: time.Date(2023, 1, 4, 0, 0, 0, 0, time.UTC)},
	}

	for _, order := range orders {
		err := sharedTestDB.Create(ctx, "orders", order)
		if err != nil {
			t.Fatalf("Failed to create test data: %v", err)
		}
	}

	defer func() {
		for _, order := range orders {
			sharedTestDB.Delete(ctx, "orders", order.ID)
		}
	}()

	tests := []struct {
		name        string
		options     []db.FindOption
		expectCount int
		description string
	}{
		{
			name: "orders for user1",
			options: []db.FindOption{
				db.WithFilter("user_id", "user1"),
			},
			expectCount: 2,
			description: "Find orders for specific user",
		},
		{
			name: "high value orders",
			options: []db.FindOption{
				db.WithFilters(map[string]any{
					"total": map[string]any{"$gte": 30.0},
				}),
			},
			expectCount: 1,
			description: "Find orders with total >= 30",
		},
		{
			name: "recent orders sorted by total",
			options: []db.FindOption{
				db.WithSort("total", db.SortDescending),
				db.WithLimit(2),
			},
			expectCount: 2,
			description: "Get top 2 orders by total value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var results []TestOrder
			err := sharedTestDB.Find(ctx, "orders", &results, tt.options...)
			assert.NoError(t, err, "Unexpected error: %v", err)
			assert.Len(t, results, tt.expectCount, fmt.Sprintf("%s: expected %d results, got %d", tt.description, tt.expectCount, len(results)))
		})
	}
}

func TestMongoConfig(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		valid  bool
	}{
		{
			name: "valid config",
			config: Config{
				Hosts:    []string{"localhost:27017"},
				Database: "test",
				Auth: AuthConfig{
					Enabled:   false,
					Username:  "",
					Password:  "",
					Mechanism: "SCRAM-SHA-256",
					Source:    "admin",
				},
			},
			valid: true,
		},
		{
			name: "config with authentication",
			config: Config{
				Hosts:    []string{"localhost:27017"},
				Database: "test",
				Auth: AuthConfig{
					Enabled:   true,
					Username:  "user",
					Password:  "pass",
					Mechanism: "SCRAM-SHA-256",
					Source:    "admin",
				},
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			database := New(tt.config)
			if database == nil {
				t.Error("Expected non-nil database instance")
			}

			mongoDB, ok := database.(*MongoDB)
			if !ok {
				t.Error("Expected *MongoDB instance")
			}

			if mongoDB.config.Database != tt.config.Database {
				t.Errorf("Expected database name %s, got %s", tt.config.Database, mongoDB.config.Database)
			}
		})
	}
}

func TestMongoDisconnect(t *testing.T) {
	// Test disconnecting without connection
	config := Config{
		Hosts:    []string{"localhost:27017"},
		Database: "test",
	}

	database := New(config)
	err := database.Disconnect()
	if err != nil {
		t.Errorf("Disconnect should not error when no connection exists: %v", err)
	}
}
