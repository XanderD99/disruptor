package db

import (
	"context"
	"fmt"
	"reflect"
)

// Create inserts a new entity with type safety
func Create[T any](ctx context.Context, db Database, entity T) error {
	table := getTableName(entity)

	return db.Create(ctx, table, entity)
}

// FindByID retrieves an entity by ID with type safety
func FindByID[T any](ctx context.Context, db Database, id any) (T, error) {
	var zero T
	table := getTableName(zero)

	result, err := db.FindByID(ctx, table, id)
	if err != nil {
		return zero, err
	}

	// Type assertion
	typed, ok := result.(T)
	if !ok {
		return zero, fmt.Errorf("expected type %T, got %T", zero, result)
	}

	return typed, nil
}

// FindAll retrieves all entities with type safety
func FindAll[T any](ctx context.Context, db Database, opts ...FindOption) ([]T, error) {
	var zero T
	table := getTableName(zero)

	results, err := db.FindAll(ctx, table, opts...)
	if err != nil {
		return nil, err
	}

	// Convert []any to []T
	typed := make([]T, 0, len(results))
	for _, item := range results {
		if t, ok := item.(T); ok {
			typed = append(typed, t)
		} else {
			// Try to convert using reflection if direct assertion fails
			converted, err := convertToType[T](item)
			if err != nil {
				return nil, fmt.Errorf("failed to convert item to type %T: %w", *new(T), err)
			}
			typed = append(typed, converted)
		}
	}

	return typed, nil
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

// Helper function to convert types using reflection
func convertToType[T any](item any) (T, error) {
	var zero T

	if item == nil {
		return zero, fmt.Errorf("cannot convert nil to type %T", zero)
	}

	// If it's already the correct type
	if typed, ok := item.(T); ok {
		return typed, nil
	}

	// Try reflection-based conversion
	targetType := reflect.TypeOf(zero)
	sourceValue := reflect.ValueOf(item)

	if sourceValue.Type().ConvertibleTo(targetType) {
		converted := sourceValue.Convert(targetType)
		value, ok := converted.Interface().(T)
		if !ok {
			return zero, fmt.Errorf("cannot cast type %T to %T", converted.Interface(), zero)
		}

		return value, nil
	}

	return zero, fmt.Errorf("cannot convert %T to %T", item, zero)
}
