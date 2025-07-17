package db

import "context"

// Database provides basic database operations without type constraints
type Database interface {
	// Connect establishes a connection to the database
	Connect(ctx context.Context) error

	// Disconnect closes the database connection
	Disconnect() error

	// Create inserts a new document
	Create(ctx context.Context, table string, entity any) error

	// FindOne retrieves a single document
	FindOne(ctx context.Context, table string, result any, opts ...FindOption) error

	// Find retrieves documents with optional filters
	Find(ctx context.Context, table string, result any, opts ...FindOption) error

	// Update updates an existing document
	Update(ctx context.Context, table string, entity any) error

	// Upsert creates or updates a document
	Upsert(ctx context.Context, table string, entity any) error

	// Delete removes a document by ID
	Delete(ctx context.Context, table string, id any) error

	// Count returns the number of documents matching the filters
	Count(ctx context.Context, table string, opts ...FindOption) (int64, error)
}

type Option[T any] func(*T)

// FindOption allows for flexible query configuration

type FindOption Option[FindOptions]

type FindOptions struct {
	Filters    map[string]any
	Sort       Sort
	Limit      int
	Offset     int
	Projection []string
}

type SortDirection string

const (
	SortAscending  SortDirection = "asc"
	SortDescending SortDirection = "desc"
)

type Sort map[string]SortDirection

// Helper functions for building queries
func WithFilters(filters map[string]any) FindOption {
	return func(opts *FindOptions) {
		opts.Filters = filters
	}
}

func WithIDFilter[T any](entity T) FindOption {
	id, err := GetEntityID(entity)
	if err != nil {
		return func(opts *FindOptions) {}
	}

	return func(opts *FindOptions) {
		if opts.Filters == nil {
			opts.Filters = make(map[string]any)
		}
		opts.Filters["id"] = id
	}
}

func WithFilter(field string, value any) FindOption {
	return func(opts *FindOptions) {
		if opts.Filters == nil {
			opts.Filters = make(map[string]any)
		}
		opts.Filters[field] = value
	}
}

func WithSort(field string, direction SortDirection) FindOption {
	return func(opts *FindOptions) {
		if opts.Sort == nil {
			opts.Sort = make(Sort)
		}
		opts.Sort[field] = direction
	}
}

func WithLimit(limit int) FindOption {
	return func(opts *FindOptions) {
		opts.Limit = limit
	}
}

func WithOffset(offset int) FindOption {
	return func(opts *FindOptions) {
		opts.Offset = offset
	}
}

func WithProjection(fields ...string) FindOption {
	return func(opts *FindOptions) {
		opts.Projection = fields
	}
}
