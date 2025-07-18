package db

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

// Test entities
type TestUser struct {
	ID     string    `json:"id" bson:"_id" db:"user_id" xml:"user_id"`
	Name   string    `json:"name" bson:"name" db:"username" xml:"name"`
	Email  string    `json:"email" bson:"email" db:"email_address" xml:"email"`
	Active bool      `json:"active" bson:"active" db:"is_active" xml:"active"`
	Score  float64   `json:"score" bson:"score" db:"user_score" xml:"score"`
	Tags   []string  `json:"tags" bson:"tags" db:"user_tags" xml:"tags"`
	Date   time.Time `json:"created_at" bson:"createdAt" db:"created_at" xml:"created_at"`
	Age    int       `json:"age" bson:"age" db:"user_age" xml:"age"`
}

type TestGuild struct {
	ID       string  `json:"id" bson:"guild_id"`
	Name     string  `json:"name" bson:"name"`
	Interval int     `json:"interval" bson:"interval"`
	Chance   float64 `json:"chance" bson:"chance"`
	Enabled  bool    `json:"enabled" bson:"enabled"`
}

type TestProduct struct {
	ID          int     `json:"id" bson:"_id" db:"product_id"`
	Name        string  `json:"name" bson:"name" db:"product_name"`
	Price       float64 `json:"price" bson:"price" db:"product_price"`
	Category    string  `json:"category" bson:"category" db:"product_category"`
	InStock     bool    `json:"in_stock" bson:"inStock" db:"in_stock"`
	Description string  `json:"description" bson:"description" db:"product_description"`
}

// Test entity with custom table name
type TestCustomTable struct {
	ID   string `json:"id" bson:"_id"`
	Name string `json:"name" bson:"name"`
}

func (t TestCustomTable) GetTable() string {
	return "custom_table_name"
}

// Test entity implementing Identifiable
type TestIdentifiable struct {
	UserID string `json:"user_id" bson:"_id" db:"id"`
	Name   string `json:"name" bson:"name"`
}

func (t TestIdentifiable) GetID() any {
	return t.UserID
}

// Test entity with no ID field
type TestNoID struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Test entity with lowercase id field
type TestLowercaseID struct {
	Id   string `json:"id" bson:"_id"` // Made exported
	Name string `json:"name"`
}

func TestGetTableName(t *testing.T) {
	tests := []struct {
		name     string
		entity   any
		expected string
	}{
		{
			name:     "simple struct",
			entity:   TestUser{},
			expected: "testusers",
		},
		{
			name:     "pointer to struct",
			entity:   &TestUser{},
			expected: "testusers",
		},
		{
			name:     "slice of structs",
			entity:   []TestUser{},
			expected: "testusers",
		},
		{
			name:     "slice of pointers",
			entity:   []*TestUser{},
			expected: "testusers",
		},
		{
			name:     "struct with custom table name",
			entity:   TestCustomTable{},
			expected: "custom_table_name",
		},
		{
			name:     "slice with custom table name",
			entity:   []TestCustomTable{},
			expected: "testcustomtables", // slice doesn't use TableNamer interface
		},
		{
			name:     "struct already plural",
			entity:   TestGuild{},
			expected: "testguilds",
		},
		{
			name:     "product struct",
			entity:   TestProduct{},
			expected: "testproducts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetTableName(tt.entity)
			if result != tt.expected {
				t.Errorf("GetTableName() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetEntityID(t *testing.T) {
	tests := []struct {
		name        string
		entity      any
		tagName     []string
		expectedKey string
		expectedVal any
		expectError bool
	}{
		{
			name:        "user with default tag (bson priority)",
			entity:      TestUser{ID: "user123"},
			tagName:     []string{},
			expectedKey: "_id", // bson tag has priority
			expectedVal: "user123",
		},
		{
			name:        "user with json tag",
			entity:      TestUser{ID: "user123"},
			tagName:     []string{"json"},
			expectedKey: "id",
			expectedVal: "user123",
		},
		{
			name:        "user with bson tag",
			entity:      TestUser{ID: "user123"},
			tagName:     []string{"bson"},
			expectedKey: "_id",
			expectedVal: "user123",
		},
		{
			name:        "user with db tag",
			entity:      TestUser{ID: "user123"},
			tagName:     []string{"db"},
			expectedKey: "user_id",
			expectedVal: "user123",
		},
		{
			name:        "user with xml tag",
			entity:      TestUser{ID: "user123"},
			tagName:     []string{"xml"},
			expectedKey: "user_id",
			expectedVal: "user123",
		},
		{
			name:        "guild with default (bson priority)",
			entity:      TestGuild{ID: "guild456"},
			tagName:     []string{},
			expectedKey: "guild_id", // bson tag
			expectedVal: "guild456",
		},
		{
			name:        "product with int ID",
			entity:      TestProduct{ID: 789},
			tagName:     []string{"json"},
			expectedKey: "id",
			expectedVal: 789,
		},
		{
			name:        "identifiable entity",
			entity:      TestIdentifiable{UserID: "ident123"},
			tagName:     []string{},
			expectedKey: "id", // default key for identifiable
			expectedVal: "ident123",
		},
		{
			name:        "identifiable with bson tag",
			entity:      TestIdentifiable{UserID: "ident123"},
			tagName:     []string{"bson"},
			expectedKey: "id", // identifiable uses default key fallback
			expectedVal: "ident123",
		},
		{
			name:        "pointer to entity",
			entity:      &TestUser{ID: "ptr123"},
			tagName:     []string{"json"},
			expectedKey: "id",
			expectedVal: "ptr123",
		},
		{
			name:        "nonexistent tag",
			entity:      TestUser{ID: "user123"},
			tagName:     []string{"nonexistent"},
			expectError: true,
		},
		{
			name:        "entity without ID field",
			entity:      TestNoID{Name: "test"},
			tagName:     []string{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, val, err := GetEntityID(tt.entity, tt.tagName...)

			if tt.expectError {
				if err == nil {
					t.Errorf("GetEntityID() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("GetEntityID() unexpected error: %v", err)
				return
			}

			if key != tt.expectedKey {
				t.Errorf("GetEntityID() key = %q, want %q", key, tt.expectedKey)
			}

			if !reflect.DeepEqual(val, tt.expectedVal) {
				t.Errorf("GetEntityID() value = %v, want %v", val, tt.expectedVal)
			}
		})
	}
}

func TestGetEntityIDHelperFunctions(t *testing.T) {
	user := TestUser{ID: "helper123"}

	tests := []struct {
		name        string
		fn          func(any) (string, any, error)
		expectedKey string
	}{
		{
			name:        "GetEntityBSONID",
			fn:          GetEntityBSONID,
			expectedKey: "_id",
		},
		{
			name:        "GetEntityJSONID",
			fn:          GetEntityJSONID,
			expectedKey: "id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, val, err := tt.fn(user)
			if err != nil {
				t.Errorf("%s() unexpected error: %v", tt.name, err)
				return
			}

			if key != tt.expectedKey {
				t.Errorf("%s() key = %q, want %q", tt.name, key, tt.expectedKey)
			}

			if val != "helper123" {
				t.Errorf("%s() value = %v, want %q", tt.name, val, "helper123")
			}
		})
	}
}

func TestCaching(t *testing.T) {
	// Clear cache before test
	typeInfoCache = sync.Map{}

	user1 := TestUser{ID: "cache1"}
	user2 := TestUser{ID: "cache2"}

	// First call should populate cache
	key1, val1, err1 := GetEntityID(user1, "json")
	if err1 != nil {
		t.Fatalf("First GetEntityID() error: %v", err1)
	}

	// Second call should use cache
	key2, val2, err2 := GetEntityID(user2, "json")
	if err2 != nil {
		t.Fatalf("Second GetEntityID() error: %v", err2)
	}

	// Keys should be the same (from cache)
	if key1 != key2 {
		t.Errorf("Cache test: keys differ: %q vs %q", key1, key2)
	}

	// Values should be different (entity-specific)
	if val1 == val2 {
		t.Errorf("Cache test: values should differ but both are %v", val1)
	}

	if key1 != "id" {
		t.Errorf("Cache test: expected key 'id', got %q", key1)
	}
}

func TestGetEntityIDEdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		entity        any
		tagName       []string
		expectError   bool
		errorContains string
		expectedKey   string
	}{
		{
			name:          "nil entity",
			entity:        nil,
			expectError:   true,
			errorContains: "reflect",
		},
		{
			name:        "empty string tag",
			entity:      TestUser{ID: "test"},
			tagName:     []string{""},
			expectedKey: "_id", // should use default
		},
		{
			name:        "multiple tag parameters (uses first)",
			entity:      TestUser{ID: "test"},
			tagName:     []string{"json", "bson", "xml"},
			expectedKey: "id", // uses first (json)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil && !tt.expectError {
					t.Errorf("GetEntityID() unexpected panic: %v", r)
				}
			}()

			key, _, err := GetEntityID(tt.entity, tt.tagName...)

			if tt.expectError {
				if err == nil {
					t.Errorf("GetEntityID() expected error but got none")
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("GetEntityID() error %q should contain %q", err.Error(), tt.errorContains)
				}
				return
			}

			if err != nil {
				t.Errorf("GetEntityID() unexpected error: %v", err)
				return
			}

			if tt.expectedKey != "" && key != tt.expectedKey {
				t.Errorf("GetEntityID() key = %q, want %q", key, tt.expectedKey)
			}
		})
	}
}

func TestReflectionHelperFunctions(t *testing.T) {
	t.Run("getReflectionInfo", func(t *testing.T) {
		user := TestUser{ID: "test"}
		ptr := &user

		// Test with value
		v1, t1 := getReflectionInfo(user)
		if v1.Kind() != reflect.Struct {
			t.Errorf("getReflectionInfo(value) kind = %v, want %v", v1.Kind(), reflect.Struct)
		}
		if t1.Name() != "TestUser" {
			t.Errorf("getReflectionInfo(value) type = %v, want TestUser", t1.Name())
		}

		// Test with pointer
		v2, t2 := getReflectionInfo(ptr)
		if v2.Kind() != reflect.Struct {
			t.Errorf("getReflectionInfo(pointer) kind = %v, want %v", v2.Kind(), reflect.Struct)
		}
		if t2.Name() != "TestUser" {
			t.Errorf("getReflectionInfo(pointer) type = %v, want TestUser", t2.Name())
		}
	})

	t.Run("getRequestedTag", func(t *testing.T) {
		tests := []struct {
			input    []string
			expected string
		}{
			{[]string{}, ""},
			{[]string{"json"}, "json"},
			{[]string{"bson", "json"}, "bson"}, // uses first
		}

		for _, tt := range tests {
			result := getRequestedTag(tt.input)
			if result != tt.expected {
				t.Errorf("getRequestedTag(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		}
	})

	t.Run("isIDField", func(t *testing.T) {
		userType := reflect.TypeOf(TestUser{})

		tests := []struct {
			fieldName string
			expected  bool
		}{
			{"ID", true},
			{"Name", false},
			{"Email", false},
		}

		for _, tt := range tests {
			field, found := userType.FieldByName(tt.fieldName)
			if !found {
				t.Fatalf("Field %q not found", tt.fieldName)
			}

			result := isIDField(field)
			if result != tt.expected {
				t.Errorf("isIDField(%q) = %v, want %v", tt.fieldName, result, tt.expected)
			}
		}
	})
}

// Benchmark tests
func BenchmarkGetTableName(b *testing.B) {
	entities := []any{
		TestUser{},
		&TestUser{},
		[]TestUser{},
		[]*TestUser{},
		TestGuild{},
		TestProduct{},
		TestCustomTable{},
	}

	b.Run("Various entities", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			entity := entities[i%len(entities)]
			GetTableName(entity)
		}
	})

	b.Run("Single entity type", func(b *testing.B) {
		user := TestUser{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			GetTableName(user)
		}
	})

	b.Run("Custom table name", func(b *testing.B) {
		custom := TestCustomTable{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			GetTableName(custom)
		}
	})
}

func BenchmarkGetEntityID(b *testing.B) {
	user := TestUser{ID: "bench123"}
	guild := TestGuild{ID: "guild456"}
	product := TestProduct{ID: 789}
	identifiable := TestIdentifiable{UserID: "ident123"}

	b.Run("Default tag", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, err := GetEntityID(user)
			_ = err // Ignore error for benchmark
		}
	})

	b.Run("JSON tag", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, err := GetEntityID(user, "json")
			_ = err // Ignore error for benchmark
		}
	})

	b.Run("BSON tag", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, err := GetEntityID(user, "bson")
			_ = err // Ignore error for benchmark
		}
	})

	b.Run("Identifiable interface", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, err := GetEntityID(identifiable)
			_ = err // Ignore error for benchmark
		}
	})

	b.Run("Mixed entities", func(b *testing.B) {
		entities := []any{user, guild, product, identifiable}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			entity := entities[i%len(entities)]
			_, _, err := GetEntityID(entity)
			_ = err // Ignore error for benchmark
		}
	})

	b.Run("Cache performance", func(b *testing.B) {
		// This tests how well caching works with repeated calls
		users := make([]TestUser, 1000)
		for i := range users {
			users[i] = TestUser{ID: fmt.Sprintf("user%d", i)}
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			user := users[i%len(users)]
			_, _, err := GetEntityID(user, "json")
			_ = err // Ignore error for benchmark
		}
	})
}

func BenchmarkGetEntityIDHelpers(b *testing.B) {
	user := TestUser{ID: "helper123"}

	b.Run("GetEntityBSONID", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, err := GetEntityBSONID(user)
			_ = err // Ignore error for benchmark
		}
	})

	b.Run("GetEntityJSONID", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, err := GetEntityJSONID(user)
			_ = err // Ignore error for benchmark
		}
	})
}

func BenchmarkCachePerformance(b *testing.B) {
	// Test cache hit performance vs cache miss
	user := TestUser{ID: "cache123"}

	b.Run("First call (cache miss)", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Clear cache to force miss
			typeInfoCache = sync.Map{}
			_, _, err := GetEntityID(user, "json")
			_ = err // Ignore error for benchmark
		}
	})

	b.Run("Subsequent calls (cache hit)", func(b *testing.B) {
		// Warm up cache
		_, _, err := GetEntityID(user, "json")
		_ = err // Ignore error for benchmark

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, err := GetEntityID(user, "json")
			_ = err // Ignore error for benchmark
		}
	})
}

func BenchmarkReflectionHelpers(b *testing.B) {
	user := TestUser{ID: "bench123"}
	userPtr := &user
	userType := reflect.TypeOf(user)

	b.Run("getReflectionInfo_value", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			getReflectionInfo(user)
		}
	})

	b.Run("getReflectionInfo_pointer", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			getReflectionInfo(userPtr)
		}
	})

	b.Run("buildTypeInfo", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buildTypeInfo(userType)
		}
	})
}

func BenchmarkSyncPools(b *testing.B) {
	user := TestUser{
		ID:    "test-123",
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}

	b.Run("GetTableName with pools", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			GetTableName(user)
		}
	})

	b.Run("GetEntityID with pools", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, err := GetEntityID(user, "json")
			_ = err // Ignore error for benchmark
		}
	})

	b.Run("Concurrent GetTableName", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			entities := []any{
				TestUser{ID: "1", Name: "User1"},
				TestGuild{ID: "2", Name: "Guild1"},
				TestProduct{ID: 3, Name: "Product1"},
			}
			i := 0
			for pb.Next() {
				entity := entities[i%len(entities)]
				GetTableName(entity)
				i++
			}
		})
	})

	b.Run("Concurrent GetEntityID", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			entities := []any{
				TestUser{ID: "1", Name: "User1"},
				TestGuild{ID: "2", Name: "Guild1"},
				TestProduct{ID: 3, Name: "Product1"},
			}
			tags := []string{"json", "bson", "db", "xml"}
			i := 0
			for pb.Next() {
				entity := entities[i%len(entities)]
				tag := tags[i%len(tags)]
				_, _, err := GetEntityID(entity, tag)
				_ = err // Ignore error for benchmark
				i++
			}
		})
	})
}

func BenchmarkMemoryPressure(b *testing.B) {
	// Test under memory pressure to see pool benefits
	entities := make([]TestUser, 1000)
	for i := range entities {
		entities[i] = TestUser{
			ID:    fmt.Sprintf("user-%d", i),
			Name:  fmt.Sprintf("User %d", i),
			Email: fmt.Sprintf("user%d@example.com", i),
			Age:   20 + i%50,
		}
	}

	b.Run("High frequency calls", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			entity := entities[i%len(entities)]
			GetTableName(entity)
			_, _, err := GetEntityID(entity, "json")
			_ = err // Ignore error for benchmark
			_, _, err = GetEntityID(entity, "bson")
			_ = err // Ignore error for benchmark
		}
	})
}
