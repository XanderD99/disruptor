//nolint:all
package memory

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/XanderD99/disruptor/pkg/db"
)

type TestGuild struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Interval  int       `json:"interval"`
	Chance    float64   `json:"chance"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
}

func BenchmarkMemoryDatabase_Operations(b *testing.B) {
	database := New()
	ctx := context.Background()

	if err := database.Connect(ctx); err != nil {
		b.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Disconnect()

	b.Run("Create", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			user := TestUser{
				ID:     fmt.Sprintf("bench_user_%d_%s", i, uuid.New()),
				Name:   fmt.Sprintf("Benchmark User %d", i),
				Email:  fmt.Sprintf("user%d@example.com", i),
				Age:    20 + (i % 60),
				Active: i%2 == 0,
				Score:  float64(i%100) / 10.0,
				Tags:   []string{"tag1", "tag2"},
			}
			if err := database.Create(ctx, "bench_users", user); err != nil {
				b.Errorf("Create failed: %v", err)
			}
		}
	})

	// Create some data for read benchmarks
	for i := 0; i < 1000; i++ {
		user := TestGuild{
			ID:       fmt.Sprintf("read_bench_guild_%d", i),
			Name:     fmt.Sprintf("Read Guild %d", i),
			Interval: 30 + (i % 60),
			Chance:   0.1 + float64(i%9)/10.0,
		}
		database.Create(ctx, "read_bench_guilds", user)
	}

	b.Run("FindByID", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			id := fmt.Sprintf("read_bench_guild_%d", i%1000)
			_, err := database.FindByID(ctx, "read_bench_guilds", id)
			if err != nil {
				b.Errorf("FindByID failed: %v", err)
			}
		}
	})

	b.Run("FindAll", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := database.FindAll(ctx, "read_bench_guilds", db.WithLimit(10))
			if err != nil {
				b.Errorf("FindAll failed: %v", err)
			}
		}
	})

	b.Run("FindAllWithFilter", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := database.FindAll(ctx, "read_bench_guilds",
				db.WithFilter("enabled", true),
				db.WithLimit(10),
			)
			if err != nil {
				b.Errorf("FindAll with filter failed: %v", err)
			}
		}
	})

	b.Run("FindAllWithComplexFilter", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := database.FindAll(ctx, "read_bench_guilds",
				db.WithFilters(map[string]any{
					"enabled":  true,
					"interval": map[string]any{"$gt": 40},
				}),
				db.WithLimit(10),
			)
			if err != nil {
				b.Errorf("FindAll with complex filter failed: %v", err)
			}
		}
	})

	b.Run("Update", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			id := fmt.Sprintf("read_bench_guild_%d", i%1000)
			guild := TestGuild{
				ID:        id,
				Name:      fmt.Sprintf("Updated Guild %d", i),
				Interval:  60 + (i % 30),
				Chance:    0.2 + float64(i%8)/10.0,
				Enabled:   i%3 == 0,
				CreatedAt: time.Now(),
			}
			if err := database.Update(ctx, "read_bench_guilds", guild); err != nil {
				// Some updates might fail if the record doesn't exist, which is expected
				continue
			}
		}
	})

	b.Run("Upsert", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			guild := TestGuild{
				ID:        fmt.Sprintf("upsert_bench_guild_%d", i%500),
				Name:      fmt.Sprintf("Upsert Guild %d", i),
				Interval:  45 + (i % 45),
				Chance:    0.3 + float64(i%7)/10.0,
				Enabled:   i%4 == 0,
				CreatedAt: time.Now(),
			}
			if err := database.Upsert(ctx, "upsert_bench_guilds", guild); err != nil {
				b.Errorf("Upsert failed: %v", err)
			}
		}
	})

	b.Run("Count", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := database.Count(ctx, "read_bench_guilds")
			if err != nil {
				b.Errorf("Count failed: %v", err)
			}
		}
	})

	b.Run("CountWithFilter", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := database.Count(ctx, "read_bench_guilds",
				db.WithFilter("enabled", true),
			)
			if err != nil {
				b.Errorf("Count with filter failed: %v", err)
			}
		}
	})

	b.Run("Delete", func(b *testing.B) {
		// Create data to delete
		for i := 0; i < b.N; i++ {
			guild := TestGuild{
				ID:        fmt.Sprintf("delete_bench_guild_%d", i),
				Name:      fmt.Sprintf("Delete Guild %d", i),
				Interval:  30,
				Chance:    0.5,
				Enabled:   true,
				CreatedAt: time.Now(),
			}
			database.Create(ctx, "delete_bench_guilds", guild)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			id := fmt.Sprintf("delete_bench_guild_%d", i)
			if err := database.Delete(ctx, "delete_bench_guilds", id); err != nil {
				b.Errorf("Delete failed: %v", err)
			}
		}
	})
}

func BenchmarkMemoryDatabase_ConcurrentOperations(b *testing.B) {
	database := New()
	ctx := context.Background()

	if err := database.Connect(ctx); err != nil {
		b.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Disconnect()

	b.Run("ConcurrentCreate", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				guild := TestGuild{
					ID:        fmt.Sprintf("concurrent_guild_%d_%d", b.N, i),
					Name:      fmt.Sprintf("Concurrent Guild %d", i),
					Interval:  30 + (i % 60),
					Chance:    0.1 + float64(i%9)/10.0,
					Enabled:   i%2 == 0,
					CreatedAt: time.Now(),
				}
				if err := database.Create(ctx, "concurrent_guilds", guild); err != nil {
					b.Errorf("Concurrent create failed: %v", err)
				}
				i++
			}
		})
	})

	// Create some data for concurrent read benchmarks
	for i := 0; i < 500; i++ {
		guild := TestGuild{
			ID:        fmt.Sprintf("concurrent_read_guild_%d", i),
			Name:      fmt.Sprintf("Concurrent Read Guild %d", i),
			Interval:  30 + (i % 60),
			Chance:    0.1 + float64(i%9)/10.0,
			Enabled:   i%2 == 0,
			CreatedAt: time.Now(),
		}
		database.Create(ctx, "concurrent_read_guilds", guild)
	}

	b.Run("ConcurrentRead", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				id := fmt.Sprintf("concurrent_read_guild_%d", i%500)
				_, err := database.FindByID(ctx, "concurrent_read_guilds", id)
				if err != nil {
					b.Errorf("Concurrent read failed: %v", err)
				}
				i++
			}
		})
	})

	b.Run("ConcurrentUpdate", func(b *testing.B) {
		// Create data to update
		for i := 0; i < 100; i++ {
			guild := TestGuild{
				ID:        fmt.Sprintf("concurrent_update_guild_%d", i),
				Name:      fmt.Sprintf("Update Guild %d", i),
				Interval:  30,
				Chance:    0.5,
				Enabled:   true,
				CreatedAt: time.Now(),
			}
			database.Create(ctx, "concurrent_update_guilds", guild)
		}

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				id := fmt.Sprintf("concurrent_update_guild_%d", i%100)
				guild := TestGuild{
					ID:        id,
					Name:      fmt.Sprintf("Concurrent Updated Guild %d", i),
					Interval:  60 + (i % 30),
					Chance:    0.7,
					Enabled:   i%2 == 0,
					CreatedAt: time.Now(),
				}
				if err := database.Update(ctx, "concurrent_update_guilds", guild); err != nil {
					// Some updates might fail due to concurrent access, which is acceptable
					continue
				}
				i++
			}
		})
	})

	b.Run("ConcurrentFindAll", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, err := database.FindAll(ctx, "concurrent_read_guilds",
					db.WithLimit(10),
					db.WithFilter("enabled", true),
				)
				if err != nil {
					b.Errorf("Concurrent FindAll failed: %v", err)
				}
			}
		})
	})

	b.Run("ConcurrentMixed", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				switch i % 4 {
				case 0: // Create
					guild := TestGuild{
						ID:        fmt.Sprintf("mixed_guild_%d_%d", b.N, i),
						Name:      fmt.Sprintf("Mixed Guild %d", i),
						Interval:  30,
						Chance:    0.5,
						Enabled:   true,
						CreatedAt: time.Now(),
					}
					database.Create(ctx, "mixed_guilds", guild)
				case 1: // Read
					id := fmt.Sprintf("concurrent_read_guild_%d", i%500)
					database.FindByID(ctx, "concurrent_read_guilds", id)
				case 2: // Update
					id := fmt.Sprintf("concurrent_update_guild_%d", i%100)
					guild := TestGuild{
						ID:       id,
						Name:     fmt.Sprintf("Mixed Updated Guild %d", i),
						Interval: 45,
						Chance:   0.6,
						Enabled:  false,
					}
					database.Update(ctx, "concurrent_update_guilds", guild)
				case 3: // FindAll
					database.FindAll(ctx, "concurrent_read_guilds", db.WithLimit(5))
				}
				i++
			}
		})
	})
}

func BenchmarkMemoryDatabase_LargeDataset(b *testing.B) {
	database := New()
	ctx := context.Background()

	if err := database.Connect(ctx); err != nil {
		b.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Disconnect()

	// Create large dataset
	const datasetSize = 10000
	b.Logf("Creating dataset of %d records", datasetSize)

	for i := 0; i < datasetSize; i++ {
		guild := TestGuild{
			ID:        fmt.Sprintf("large_guild_%d", i),
			Name:      fmt.Sprintf("Large Dataset Guild %d", i),
			Interval:  30 + (i % 120),
			Chance:    float64(i%100) / 100.0,
			Enabled:   i%3 == 0,
			CreatedAt: time.Now().Add(-time.Duration(i) * time.Minute),
		}
		if err := database.Create(ctx, "large_guilds", guild); err != nil {
			b.Fatalf("Failed to create large dataset: %v", err)
		}
	}

	b.Run("FindAllLargeDataset", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			results, err := database.FindAll(ctx, "large_guilds")
			if err != nil {
				b.Errorf("FindAll large dataset failed: %v", err)
			}
			if len(results) != datasetSize {
				b.Errorf("Expected %d results, got %d", datasetSize, len(results))
			}
		}
	})

	b.Run("FindAllWithFilterLargeDataset", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := database.FindAll(ctx, "large_guilds",
				db.WithFilter("enabled", true),
			)
			if err != nil {
				b.Errorf("FindAll with filter large dataset failed: %v", err)
			}
		}
	})

	b.Run("FindAllWithComplexFilterLargeDataset", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := database.FindAll(ctx, "large_guilds",
				db.WithFilters(map[string]any{
					"enabled":  true,
					"interval": map[string]any{"$gte": 50},
					"chance":   map[string]any{"$lt": 0.8},
				}),
			)
			if err != nil {
				b.Errorf("FindAll with complex filter large dataset failed: %v", err)
			}
		}
	})

	b.Run("FindAllWithSortLargeDataset", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := database.FindAll(ctx, "large_guilds",
				db.WithSort("interval", db.SortAscending),
				db.WithLimit(100),
			)
			if err != nil {
				b.Errorf("FindAll with sort large dataset failed: %v", err)
			}
		}
	})

	b.Run("CountLargeDataset", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			count, err := database.Count(ctx, "large_guilds")
			if err != nil {
				b.Errorf("Count large dataset failed: %v", err)
			}
			if count != datasetSize {
				b.Errorf("Expected count %d, got %d", datasetSize, count)
			}
		}
	})

	b.Run("CountWithFilterLargeDataset", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := database.Count(ctx, "large_guilds",
				db.WithFilter("enabled", true),
			)
			if err != nil {
				b.Errorf("Count with filter large dataset failed: %v", err)
			}
		}
	})

	b.Run("RandomAccessLargeDataset", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			id := fmt.Sprintf("large_guild_%d", i%datasetSize)
			_, err := database.FindByID(ctx, "large_guilds", id)
			if err != nil {
				b.Errorf("Random access large dataset failed: %v", err)
			}
		}
	})

	b.Run("IndexedLookupLargeDataset", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Test indexed lookup performance
			_, err := database.FindAll(ctx, "large_guilds",
				db.WithFilter("interval", 60+i%60),
				db.WithLimit(1),
			)
			if err != nil {
				b.Errorf("Indexed lookup large dataset failed: %v", err)
			}
		}
	})
}

func BenchmarkMemoryDatabase_IndexPerformance(b *testing.B) {
	database := New()
	ctx := context.Background()

	if err := database.Connect(ctx); err != nil {
		b.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Disconnect()

	// Create data with varied indexable fields
	const datasetSize = 5000
	for i := 0; i < datasetSize; i++ {
		guild := TestGuild{
			ID:        fmt.Sprintf("index_guild_%d", i),
			Name:      fmt.Sprintf("Index Guild %d", i),
			Interval:  30 + (i % 100),       // 100 different interval values
			Chance:    float64(i%50) / 50.0, // 50 different chance values
			Enabled:   i%2 == 0,             // 2 different enabled values
			CreatedAt: time.Now().Add(-time.Duration(i) * time.Hour),
		}
		database.Create(ctx, "index_guilds", guild)
	}

	b.Run("IndexedFieldLookup", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			interval := 30 + (i % 100)
			_, err := database.FindAll(ctx, "index_guilds",
				db.WithFilter("interval", interval),
			)
			if err != nil {
				b.Errorf("Indexed field lookup failed: %v", err)
			}
		}
	})

	b.Run("NonIndexedFieldLookup", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			name := fmt.Sprintf("Index Guild %d", i%datasetSize)
			_, err := database.FindAll(ctx, "index_guilds",
				db.WithFilter("name", name),
			)
			if err != nil {
				b.Errorf("Non-indexed field lookup failed: %v", err)
			}
		}
	})

	b.Run("BooleanIndexLookup", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			enabled := i%2 == 0
			_, err := database.FindAll(ctx, "index_guilds",
				db.WithFilter("enabled", enabled),
			)
			if err != nil {
				b.Errorf("Boolean index lookup failed: %v", err)
			}
		}
	})

	b.Run("RangeQueryOnIndexedField", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			minInterval := 30 + (i % 50)
			_, err := database.FindAll(ctx, "index_guilds",
				db.WithFilters(map[string]any{
					"interval": map[string]any{"$gte": minInterval},
				}),
				db.WithLimit(10),
			)
			if err != nil {
				b.Errorf("Range query on indexed field failed: %v", err)
			}
		}
	})
}
