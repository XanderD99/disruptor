package database

import "context"

type Pagination struct {
	Limit  int
	Offset int
}

type SortDirection string

const (
	Asc  SortDirection = "asc"
	Desc SortDirection = "desc"
)

type Sort struct {
	Field     string
	Direction SortDirection
}

type Database interface {
	Open(ctx context.Context) error
	Close() error

	Create(ctx context.Context, entity any) error
	Update(ctx context.Context, entity any) error
	Upsert(ctx context.Context, entity any) error
	Delete(ctx context.Context, id string, entity any) error

	FindAll(ctx context.Context, entity any, opts ...FindAllOption) (any, error)
	FindByID(ctx context.Context, id string, entity any) (any, error)
}

type FindAllOptions struct {
	Filters    map[string]any
	Pagination Pagination
	Sort       []Sort
}

type FindAllOption func(f *FindAllOptions)

func WithFilters(filters map[string]any) FindAllOption {
	return func(f *FindAllOptions) {
		f.Filters = filters
	}
}

func WithPagination(pagination Pagination) FindAllOption {
	return func(f *FindAllOptions) {
		f.Pagination = pagination
	}
}

func WithSort(sort []Sort) FindAllOption {
	return func(f *FindAllOptions) {
		f.Sort = sort
	}
}
