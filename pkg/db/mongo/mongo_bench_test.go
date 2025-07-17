//go:build integration

//nolint:all

package mongo

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go/modules/mongodb"

	"github.com/XanderD99/disruptor/pkg/db"
)

func BenchmarkMongoDB_Operations(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	ctx := context.Background()

	// Start MongoDB container
	mongoContainer, err := mongodb.Run(ctx,
		"mongo:7.0",
		mongodb.WithUsername("benchuser"),
		mongodb.WithPassword("benchpass"),
	)
	if err != nil {
		b.Fatalf("Failed to start MongoDB container: %v", err)
	}
	defer mongoContainer.Terminate(ctx)

	// Get connection details
	endpoint, err := mongoContainer.Endpoint(ctx, "")
	if err != nil {
		b.Fatalf("Failed to get endpoint: %v", err)
	}

	// Create config and database
	config := Config{
		Hosts:    []string{endpoint},
		Database: "bench_disruptor",
		Auth: AuthConfig{
			Enabled:   true,
			Username:  "benchuser",
			Password:  "benchpass",
			Mechanism: "SCRAM-SHA-256",
			Source:    "admin",
		},
		Timeout: TimeoutConfig{
			Connect: 10 * time.Second,
			Query:   30 * time.Second,
		},
	}

	database := New(config)
	if err := database.Connect(ctx); err != nil {
		b.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer database.Disconnect()

	b.Run("Create", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			user := &TestUser{
				ID:    fmt.Sprintf("bench_user_%d", i),
				Name:  fmt.Sprintf("User %d", i),
				Email: fmt.Sprintf("user%d@example.com", i),
				Age:   20 + (i % 50),
			}
			if err := database.Create(ctx, "bench_users", user); err != nil {
				b.Errorf("Create failed: %v", err)
			}
		}
	})

	// Create some data for read benchmarks
	for i := 0; i < 1000; i++ {
		user := &TestUser{
			ID:     fmt.Sprintf("read_bench_user_%d", i),
			Name:   fmt.Sprintf("Read User %d", i),
			Email:  fmt.Sprintf("readuser%d@example.com", i),
			Age:    20 + (i % 50),
			Active: i%2 == 0,
		}
		database.Create(ctx, "read_bench_users", user)
	}

	b.Run("FindByID", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			id := fmt.Sprintf("read_bench_user_%d", i%1000)
			_, err := database.FindByID(ctx, "read_bench_users", id)
			if err != nil {
				b.Errorf("FindByID failed: %v", err)
			}
		}
	})

	b.Run("FindAll", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := database.FindAll(ctx, "read_bench_users", db.WithLimit(10))
			if err != nil {
				b.Errorf("FindAll failed: %v", err)
			}
		}
	})

	b.Run("FindAllWithFilter", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := database.FindAll(ctx, "read_bench_users",
				db.WithFilter("active", true),
				db.WithLimit(10),
			)
			if err != nil {
				b.Errorf("FindAll with filter failed: %v", err)
			}
		}
	})

	b.Run("Update", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			id := fmt.Sprintf("read_bench_user_%d", i%1000)
			user := &TestUser{
				ID:    id,
				Name:  fmt.Sprintf("Updated User %d", i),
				Email: fmt.Sprintf("updated%d@example.com", i),
				Age:   30 + (i % 40),
			}
			if err := database.Update(ctx, "read_bench_users", user); err != nil {
				// Some updates might fail if the record doesn't exist, which is expected
				continue
			}
		}
	})

	b.Run("Count", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := database.Count(ctx, "read_bench_users")
			if err != nil {
				b.Errorf("Count failed: %v", err)
			}
		}
	})
}

func BenchmarkMongoDB_ConcurrentOperations(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	ctx := context.Background()

	// Start MongoDB container
	mongoContainer, err := mongodb.Run(ctx,
		"mongo:7.0",
		mongodb.WithUsername("concurrentuser"),
		mongodb.WithPassword("concurrentpass"),
	)
	if err != nil {
		b.Fatalf("Failed to start MongoDB container: %v", err)
	}
	defer mongoContainer.Terminate(ctx)

	// Get connection details
	endpoint, err := mongoContainer.Endpoint(ctx, "")
	if err != nil {
		b.Fatalf("Failed to get endpoint: %v", err)
	}

	// Create config and database
	config := Config{
		Hosts:    []string{endpoint},
		Database: "concurrent_bench_disruptor",
		Auth: AuthConfig{
			Enabled:   true,
			Username:  "concurrentuser",
			Password:  "concurrentpass",
			Mechanism: "SCRAM-SHA-256",
			Source:    "admin",
		},
		Pool: PoolConfig{
			MinSize:       10,
			MaxSize:       100,
			MaxConnecting: 20,
			MaxIdleTime:   5 * time.Minute,
		},
		Timeout: TimeoutConfig{
			Connect: 10 * time.Second,
			Query:   30 * time.Second,
		},
	}

	database := New(config)
	if err := database.Connect(ctx); err != nil {
		b.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer database.Disconnect()

	b.Run("ConcurrentCreate", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				user := &TestUser{
					ID:    fmt.Sprintf("concurrent_user_%d_%d", b.N, i),
					Name:  fmt.Sprintf("Concurrent User %d", i),
					Email: fmt.Sprintf("concurrent%d@example.com", i),
					Age:   20 + (i % 50),
				}
				if err := database.Create(ctx, "concurrent_users", user); err != nil {
					b.Errorf("Concurrent create failed: %v", err)
				}
				i++
			}
		})
	})

	// Create some data for concurrent read benchmarks
	for i := 0; i < 500; i++ {
		user := &TestUser{
			ID:    fmt.Sprintf("concurrent_read_user_%d", i),
			Name:  fmt.Sprintf("Concurrent Read User %d", i),
			Email: fmt.Sprintf("concurrentread%d@example.com", i),
			Age:   20 + (i % 50),
		}
		database.Create(ctx, "concurrent_read_users", user)
	}

	b.Run("ConcurrentRead", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				id := fmt.Sprintf("concurrent_read_user_%d", i%500)
				_, err := database.FindByID(ctx, "concurrent_read_users", id)
				if err != nil {
					b.Errorf("Concurrent read failed: %v", err)
				}
				i++
			}
		})
	})
}
