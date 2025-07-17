package db

import (
	"context"
)

// Create inserts a new entity with type safety
func Create[T any](ctx context.Context, db Database, entity T) error {
	table := getTableName(entity)

	return db.Create(ctx, table, entity)
}

// FindByID retrieves an entity by ID with type safety
func FindByID[T any](ctx context.Context, db Database, id any) (T, error) {
	var result T
	table := getTableName(result)

	err := db.FindByID(ctx, table, id, &result)
	if err != nil {
		return result, err
	}

	return result, nil
}

// FindAll retrieves all entities with type safety
func FindAll[T any](ctx context.Context, db Database, opts ...FindOption) ([]T, error) {
	var zero T
	table := getTableName(zero)

	var results []T
	err := db.FindAll(ctx, table, &results, opts...)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// Update updates an existing entity with type safety
func Update[T any](ctx context.Context, db Database, entity T) error {
	table := getTableName(entity)
	return db.Update(ctx, table, entity)
}

// Upsert creates or updates an entity with type safety
func Upsert[T any](ctx context.Context, db Database, entity T) error {
	table := getTableName(entity)

	return db.Upsert(ctx, table, entity)
}

// Delete removes an entity by ID
func Delete[T any](ctx context.Context, db Database, entity T) error {
	var zero T
	table := getTableName(zero)

	id, err := GetEntityID(entity)
	if err != nil {
		return err
	}

	return db.Delete(ctx, table, id)
}

// Count returns the number of entities matching the filters
func Count[T any](ctx context.Context, db Database, opts ...FindOption) (int64, error) {
	var zero T
	table := getTableName(zero)

	return db.Count(ctx, table, opts...)
}
