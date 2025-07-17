package file

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/XanderD99/disruptor/pkg/db"
)

type FileDatabase struct {
	mu      sync.RWMutex
	baseDir string
	cache   map[string]*TableCache
}

type TableCache struct {
	mu      sync.RWMutex
	records map[string]json.RawMessage // cached records
	dirty   bool                       // needs to be written to disk
}

type Config struct {
	BaseDirectory string `env:"FILE_DB_DIR" default:"./data"`
}

func New(config Config) db.Database {
	return &FileDatabase{
		baseDir: config.BaseDirectory,
		cache:   make(map[string]*TableCache),
	}
}

func (f *FileDatabase) Connect(ctx context.Context) error {
	// Ensure base directory exists
	if err := os.MkdirAll(f.baseDir, 0755); err != nil {
		return fmt.Errorf("failed to create base directory: %w", err)
	}

	return nil
}

func (f *FileDatabase) Disconnect() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Flush all dirty caches to disk
	for tableName, cache := range f.cache {
		if cache.dirty {
			if err := f.flushTableToDisk(tableName, cache); err != nil {
				return fmt.Errorf("failed to flush table %s: %w", tableName, err)
			}
		}
	}

	// Clear cache
	f.cache = make(map[string]*TableCache)

	return nil
}

func (f *FileDatabase) getTablePath(tableName string) string {
	return filepath.Join(f.baseDir, tableName+".json")
}

func (f *FileDatabase) loadTableFromDisk(tableName string) (*TableCache, error) {
	cache := &TableCache{
		records: make(map[string]json.RawMessage),
		dirty:   false,
	}

	tablePath := f.getTablePath(tableName)

	// Check if file exists
	if _, err := os.Stat(tablePath); os.IsNotExist(err) {
		return cache, nil // Return empty cache
	}

	// Read file
	data, err := os.ReadFile(tablePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read table file: %w", err)
	}

	// Parse JSON
	if len(data) > 0 {
		if err := json.Unmarshal(data, &cache.records); err != nil {
			return nil, fmt.Errorf("failed to parse table file: %w", err)
		}
	}

	return cache, nil
}

func (f *FileDatabase) flushTableToDisk(tableName string, cache *TableCache) error {
	tablePath := f.getTablePath(tableName)

	// Marshal records to JSON
	data, err := json.MarshalIndent(cache.records, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal records: %w", err)
	}

	// Write to file
	if err := os.WriteFile(tablePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write table file: %w", err)
	}

	cache.dirty = false
	return nil
}

func (f *FileDatabase) getOrLoadTable(tableName string) (*TableCache, error) {
	f.mu.RLock()
	if cache, exists := f.cache[tableName]; exists {
		f.mu.RUnlock()
		return cache, nil
	}
	f.mu.RUnlock()

	f.mu.Lock()
	defer f.mu.Unlock()

	// Double-check after acquiring write lock
	if cache, exists := f.cache[tableName]; exists {
		return cache, nil
	}

	// Load from disk
	cache, err := f.loadTableFromDisk(tableName)
	if err != nil {
		return nil, err
	}

	f.cache[tableName] = cache
	return cache, nil
}

func (f *FileDatabase) Create(ctx context.Context, table string, entity any) error {
	cache, err := f.getOrLoadTable(table)
	if err != nil {
		return err
	}

	// Extract ID from entity
	id, err := f.extractID(entity)
	if err != nil {
		return fmt.Errorf("failed to extract ID: %w", err)
	}

	idStr := fmt.Sprintf("%v", id)

	cache.mu.Lock()
	defer cache.mu.Unlock()

	// Check if record already exists
	if _, exists := cache.records[idStr]; exists {
		return fmt.Errorf("record with ID %s already exists", idStr)
	}

	// Marshal entity to JSON
	data, err := json.Marshal(entity)
	if err != nil {
		return fmt.Errorf("failed to marshal entity: %w", err)
	}

	// Store in cache
	cache.records[idStr] = data
	cache.dirty = true

	// Optionally flush immediately for durability
	return f.flushTableToDisk(table, cache)
}

func (f *FileDatabase) FindByID(ctx context.Context, table string, id any) (any, error) {
	cache, err := f.getOrLoadTable(table)
	if err != nil {
		return nil, err
	}

	idStr := fmt.Sprintf("%v", id)

	cache.mu.RLock()
	defer cache.mu.RUnlock()

	data, exists := cache.records[idStr]
	if !exists {
		return nil, fmt.Errorf("record with ID %s not found", idStr)
	}

	// Return raw JSON data - will be converted by operations layer
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal record: %w", err)
	}

	return result, nil
}

func (f *FileDatabase) FindAll(ctx context.Context, table string, opts ...db.FindOption) ([]any, error) {
	options := &db.FindOptions{}
	for _, opt := range opts {
		opt(options)
	}

	cache, err := f.getOrLoadTable(table)
	if err != nil {
		return nil, err
	}

	cache.mu.RLock()
	defer cache.mu.RUnlock()

	var results []any

	// Convert all records
	for _, data := range cache.records {
		var record map[string]any
		if err := json.Unmarshal(data, &record); err != nil {
			continue // Skip invalid records
		}

		if f.matchesFilters(record, options.Filters) {
			results = append(results, record)
		}
	}

	// Apply sorting
	if len(options.Sort) > 0 {
		f.sortResults(results, options.Sort)
	}

	// Apply pagination
	start := options.Offset
	if start < 0 {
		start = 0
	}
	if start >= len(results) {
		return []any{}, nil
	}

	end := len(results)
	if options.Limit > 0 {
		end = start + options.Limit
		if end > len(results) {
			end = len(results)
		}
	}

	return results[start:end], nil
}

func (f *FileDatabase) Update(ctx context.Context, table string, entity any) error {
	cache, err := f.getOrLoadTable(table)
	if err != nil {
		return err
	}

	// Extract ID from entity
	id, err := f.extractID(entity)
	if err != nil {
		return fmt.Errorf("failed to extract ID: %w", err)
	}

	idStr := fmt.Sprintf("%v", id)

	cache.mu.Lock()
	defer cache.mu.Unlock()

	// Check if record exists
	if _, exists := cache.records[idStr]; !exists {
		return fmt.Errorf("record with ID %s not found", idStr)
	}

	// Marshal entity to JSON
	data, err := json.Marshal(entity)
	if err != nil {
		return fmt.Errorf("failed to marshal entity: %w", err)
	}

	// Update in cache
	cache.records[idStr] = data
	cache.dirty = true

	// Flush to disk
	return f.flushTableToDisk(table, cache)
}

func (f *FileDatabase) Upsert(ctx context.Context, table string, entity any) error {
	cache, err := f.getOrLoadTable(table)
	if err != nil {
		return err
	}

	// Extract ID from entity
	id, err := f.extractID(entity)
	if err != nil {
		return fmt.Errorf("failed to extract ID: %w", err)
	}

	idStr := fmt.Sprintf("%v", id)

	cache.mu.Lock()
	defer cache.mu.Unlock()

	// Marshal entity to JSON
	data, err := json.Marshal(entity)
	if err != nil {
		return fmt.Errorf("failed to marshal entity: %w", err)
	}

	// Store/update in cache
	cache.records[idStr] = data
	cache.dirty = true

	// Flush to disk
	return f.flushTableToDisk(table, cache)
}

func (f *FileDatabase) Delete(ctx context.Context, table string, id any) error {
	cache, err := f.getOrLoadTable(table)
	if err != nil {
		return err
	}

	idStr := fmt.Sprintf("%v", id)

	cache.mu.Lock()
	defer cache.mu.Unlock()

	// Check if record exists
	if _, exists := cache.records[idStr]; !exists {
		return fmt.Errorf("record with ID %s not found", idStr)
	}

	// Remove from cache
	delete(cache.records, idStr)
	cache.dirty = true

	// Flush to disk
	return f.flushTableToDisk(table, cache)
}

func (f *FileDatabase) Count(ctx context.Context, table string, opts ...db.FindOption) (int64, error) {
	options := &db.FindOptions{}
	for _, opt := range opts {
		opt(options)
	}

	cache, err := f.getOrLoadTable(table)
	if err != nil {
		return 0, err
	}

	cache.mu.RLock()
	defer cache.mu.RUnlock()

	count := int64(0)

	for _, data := range cache.records {
		var record map[string]any
		if err := json.Unmarshal(data, &record); err != nil {
			continue // Skip invalid records
		}

		if f.matchesFilters(record, options.Filters) {
			count++
		}
	}

	return count, nil
}

// Helper methods (similar to memory database)
func (f *FileDatabase) extractID(entity any) (any, error) {
	v := reflect.ValueOf(entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	idField := v.FieldByName("ID")
	if !idField.IsValid() {
		return nil, fmt.Errorf("ID field not found")
	}

	return idField.Interface(), nil
}

func (f *FileDatabase) matchesFilters(record map[string]any, filters map[string]any) bool {
	if len(filters) == 0 {
		return true
	}

	for fieldName, expectedValue := range filters {
		fieldValue := f.getFieldValue(record, fieldName)
		if !f.matchesValue(fieldValue, expectedValue) {
			return false
		}
	}

	return true
}

func (f *FileDatabase) getFieldValue(record map[string]any, fieldName string) any {
	// Handle nested field names
	parts := strings.Split(fieldName, ".")
	current := record

	for i, part := range parts {
		if i == len(parts)-1 {
			return current[part]
		}

		if next, ok := current[part].(map[string]any); ok {
			current = next
		} else {
			return nil
		}
	}

	return nil
}

func (f *FileDatabase) matchesValue(fieldValue, expectedValue any) bool {
	// Handle MongoDB-style operators (similar to memory implementation)
	if expectedMap, ok := expectedValue.(map[string]any); ok {
		for operator, operand := range expectedMap {
			switch operator {
			case "$lt":
				return f.compareValues(fieldValue, operand) < 0
			case "$lte":
				return f.compareValues(fieldValue, operand) <= 0
			case "$gt":
				return f.compareValues(fieldValue, operand) > 0
			case "$gte":
				return f.compareValues(fieldValue, operand) >= 0
			case "$ne":
				return !reflect.DeepEqual(fieldValue, operand)
			case "$in":
				if slice, ok := operand.([]any); ok {
					for _, item := range slice {
						if reflect.DeepEqual(fieldValue, item) {
							return true
						}
					}
					return false
				}
			}
		}
		return false
	}

	return reflect.DeepEqual(fieldValue, expectedValue)
}

func (f *FileDatabase) compareValues(a, b any) int {
	// Convert to float64 for numeric comparison
	if fa, ok := f.toFloat64(a); ok {
		if fb, ok := f.toFloat64(b); ok {
			if fa < fb {
				return -1
			} else if fa > fb {
				return 1
			}
			return 0
		}
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

func (f *FileDatabase) toFloat64(v any) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case int32:
		return float64(val), true
	default:
		return 0, false
	}
}

func (f *FileDatabase) sortResults(results []any, sortOptions db.Sort) {
	sort.Slice(results, func(i, j int) bool {
		recordI, okI := results[i].(map[string]any)
		recordJ, okJ := results[j].(map[string]any)

		if !okI || !okJ {
			return false
		}

		for field, direction := range sortOptions {
			fieldI := f.getFieldValue(recordI, field)
			fieldJ := f.getFieldValue(recordJ, field)

			cmp := f.compareValues(fieldI, fieldJ)
			if cmp != 0 {
				if direction == db.SortAscending {
					return cmp < 0
				}
				return cmp > 0
			}
		}

		return false
	})
}
