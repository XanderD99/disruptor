package db

import "context"

// Database provides basic database operations without type constraints
type Database interface {
	// Connect establishes a connection to the database
	Connect(ctx context.Context) error

	// Disconnect closes the database connection
	Disconnect() error

	// Create inserts a new document
	Create(ctx context.Context, result any) error

	// FindOne retrieves a single document
	FindOne(ctx context.Context, result any, opts ...FindOption) error

	// FindByID retrieves a document by its ID
	FindByID(ctx context.Context, id any, result any) error

	// Find retrieves documents with optional filters
	Find(ctx context.Context, results any, opts ...FindOption) error

	// Update updates an existing document
	Update(ctx context.Context, entity any) error

	// Upsert creates or updates a document
	Upsert(ctx context.Context, entity any) error

	// Delete removes a document by ID
	Delete(ctx context.Context, id any) error

	// Count returns the number of documents matching the filters
	Count(ctx context.Context, entity any, opts ...FindOption) (int64, error)
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
