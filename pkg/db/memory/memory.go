package memory

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/XanderD99/disruptor/pkg/db"
)

type MemoryDatabase struct {
	mu     sync.RWMutex
	tables map[string]*Table
}

type Table struct {
	mu      sync.RWMutex
	records map[string]any              // key: ID as string, value: the entity
	indexes map[string]map[any][]string // field -> value -> []ids
}

func New() db.Database {
	return &MemoryDatabase{
		tables: make(map[string]*Table),
	}
}

func (m *MemoryDatabase) Connect(ctx context.Context) error {
	// No-op for in-memory database
	return nil
}

func (m *MemoryDatabase) Disconnect() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Clear all tables
	m.tables = make(map[string]*Table)
	return nil
}

func (m *MemoryDatabase) getOrCreateTable(tableName string) *Table {
	m.mu.Lock()
	defer m.mu.Unlock()

	if table, exists := m.tables[tableName]; exists {
		return table
	}

	table := &Table{
		records: make(map[string]any),
		indexes: make(map[string]map[any][]string),
	}
	m.tables[tableName] = table
	return table
}

func (m *MemoryDatabase) Create(ctx context.Context, table string, entity any) error {
	t := m.getOrCreateTable(table)

	// Extract ID from entity
	id, err := m.extractID(entity)
	if err != nil {
		return fmt.Errorf("failed to extract ID: %w", err)
	}

	idStr := fmt.Sprintf("%v", id)

	t.mu.Lock()
	defer t.mu.Unlock()

	// Check if record already exists
	if _, exists := t.records[idStr]; exists {
		return fmt.Errorf("record with ID %s already exists", idStr)
	}

	// Store the record
	t.records[idStr] = entity

	// Update indexes
	m.updateIndexes(t, idStr, entity)

	return nil
}

func (m *MemoryDatabase) FindByID(ctx context.Context, table string, id any) (any, error) {
	t := m.getOrCreateTable(table)

	idStr := fmt.Sprintf("%v", id)

	t.mu.RLock()
	defer t.mu.RUnlock()

	record, exists := t.records[idStr]
	if !exists {
		return nil, fmt.Errorf("record with ID %s not found", idStr)
	}

	return record, nil
}

func (m *MemoryDatabase) FindAll(ctx context.Context, table string, opts ...db.FindOption) ([]any, error) {
	options := &db.FindOptions{}
	for _, opt := range opts {
		opt(options)
	}

	t := m.getOrCreateTable(table)

	t.mu.RLock()
	defer t.mu.RUnlock()

	// Get all records first
	var results []any
	for _, record := range t.records {
		if m.matchesFilters(record, options.Filters) {
			results = append(results, record)
		}
	}

	// Apply sorting
	if len(options.Sort) > 0 {
		m.sortResults(results, options.Sort)
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

func (m *MemoryDatabase) Update(ctx context.Context, table string, entity any) error {
	t := m.getOrCreateTable(table)

	// Extract ID from entity
	id, err := m.extractID(entity)
	if err != nil {
		return fmt.Errorf("failed to extract ID: %w", err)
	}

	idStr := fmt.Sprintf("%v", id)

	t.mu.Lock()
	defer t.mu.Unlock()

	// Check if record exists
	if _, exists := t.records[idStr]; !exists {
		return fmt.Errorf("record with ID %s not found", idStr)
	}

	// Update the record
	t.records[idStr] = entity

	// Update indexes
	m.updateIndexes(t, idStr, entity)

	return nil
}

func (m *MemoryDatabase) Upsert(ctx context.Context, table string, entity any) error {
	t := m.getOrCreateTable(table)

	// Extract ID from entity
	id, err := m.extractID(entity)
	if err != nil {
		return fmt.Errorf("failed to extract ID: %w", err)
	}

	idStr := fmt.Sprintf("%v", id)

	t.mu.Lock()
	defer t.mu.Unlock()

	// Store/update the record
	t.records[idStr] = entity

	// Update indexes
	m.updateIndexes(t, idStr, entity)

	return nil
}

func (m *MemoryDatabase) Delete(ctx context.Context, table string, id any) error {
	t := m.getOrCreateTable(table)

	idStr := fmt.Sprintf("%v", id)

	t.mu.Lock()
	defer t.mu.Unlock()

	// Check if record exists
	if _, exists := t.records[idStr]; !exists {
		return fmt.Errorf("record with ID %s not found", idStr)
	}

	// Remove from records
	delete(t.records, idStr)

	// Remove from indexes
	m.removeFromIndexes(t, idStr)

	return nil
}

func (m *MemoryDatabase) Count(ctx context.Context, table string, opts ...db.FindOption) (int64, error) {
	options := &db.FindOptions{}
	for _, opt := range opts {
		opt(options)
	}

	t := m.getOrCreateTable(table)

	t.mu.RLock()
	defer t.mu.RUnlock()

	count := int64(0)
	for _, record := range t.records {
		if m.matchesFilters(record, options.Filters) {
			count++
		}
	}

	return count, nil
}

// Helper methods
func (m *MemoryDatabase) extractID(entity any) (any, error) {
	v := reflect.ValueOf(entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Try to find ID field
	idField := v.FieldByName("ID")
	if !idField.IsValid() {
		return nil, fmt.Errorf("ID field not found")
	}

	return idField.Interface(), nil
}

func (m *MemoryDatabase) matchesFilters(record any, filters map[string]any) bool {
	if len(filters) == 0 {
		return true
	}

	v := reflect.ValueOf(record)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	for fieldName, expectedValue := range filters {
		fieldValue := m.getFieldValue(v, fieldName)
		if !m.matchesValue(fieldValue, expectedValue) {
			return false
		}
	}

	return true
}

func (m *MemoryDatabase) getFieldValue(v reflect.Value, fieldName string) any {
	// Handle nested field names (e.g., "user.name")
	parts := strings.Split(fieldName, ".")
	current := v

	for _, part := range parts {
		if current.Kind() == reflect.Ptr {
			if current.IsNil() {
				return nil
			}
			current = current.Elem()
		}
		field := current.FieldByName(part)
		if !field.IsValid() {
			return nil
		}
		current = field
	}

	return current.Interface()
}

func (m *MemoryDatabase) matchesValue(fieldValue, expectedValue any) bool {
	// Handle special MongoDB-style operators
	if expectedMap, ok := expectedValue.(map[string]any); ok {
		for operator, operand := range expectedMap {
			switch operator {
			case "$lt":
				return m.compareValues(fieldValue, operand) < 0
			case "$lte":
				return m.compareValues(fieldValue, operand) <= 0
			case "$gt":
				return m.compareValues(fieldValue, operand) > 0
			case "$gte":
				return m.compareValues(fieldValue, operand) >= 0
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

func (m *MemoryDatabase) compareValues(a, b any) int {
	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)

	// Handle numeric comparisons
	if va.CanFloat() && vb.CanFloat() {
		fa, fb := va.Float(), vb.Float()
		if fa < fb {
			return -1
		} else if fa > fb {
			return 1
		}
		return 0
	}

	if va.CanInt() && vb.CanInt() {
		ia, ib := va.Int(), vb.Int()
		if ia < ib {
			return -1
		} else if ia > ib {
			return 1
		}
		return 0
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

func (m *MemoryDatabase) sortResults(results []any, sortOptions db.Sort) {
	sort.Slice(results, func(i, j int) bool {
		vi := reflect.ValueOf(results[i])
		vj := reflect.ValueOf(results[j])

		if vi.Kind() == reflect.Ptr {
			vi = vi.Elem()
		}
		if vj.Kind() == reflect.Ptr {
			vj = vj.Elem()
		}

		for field, direction := range sortOptions {
			fieldI := m.getFieldValue(vi, field)
			fieldJ := m.getFieldValue(vj, field)

			cmp := m.compareValues(fieldI, fieldJ)
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

func (m *MemoryDatabase) updateIndexes(t *Table, id string, entity any) {
	// Simple indexing implementation
	v := reflect.ValueOf(entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	typ := v.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldName := field.Name
		fieldValue := v.Field(i).Interface()

		// Skip unhashable types (slices, maps, functions)
		if !isHashable(fieldValue) {
			continue
		}

		if t.indexes[fieldName] == nil {
			t.indexes[fieldName] = make(map[any][]string)
		}

		// Add ID to the index for this field value
		ids := t.indexes[fieldName][fieldValue]
		ids = append(ids, id)
		t.indexes[fieldName][fieldValue] = ids
	}
}

// isHashable checks if a value can be used as a map key
func isHashable(v any) bool {
	if v == nil {
		return true
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice, reflect.Map, reflect.Func:
		return false
	case reflect.Array:
		// Arrays are hashable if their element type is hashable
		if rv.Len() == 0 {
			return true
		}
		// For arrays, we can safely check the first element if it's accessible
		if rv.Index(0).CanInterface() {
			return isHashable(rv.Index(0).Interface())
		}
		// If we can't access the element, assume it's not hashable to be safe
		return false
	case reflect.Struct:
		// Structs are hashable if all their exported fields are hashable
		for i := 0; i < rv.NumField(); i++ {
			field := rv.Field(i)
			// Only check exported fields to avoid panic
			if field.CanInterface() && !isHashable(field.Interface()) {
				return false
			}
		}
		return true
	case reflect.Ptr:
		if rv.IsNil() {
			return true
		}
		return isHashable(rv.Elem().Interface())
	}
	return true
}

func (m *MemoryDatabase) removeFromIndexes(t *Table, id string) {
	// Remove ID from all indexes
	for fieldName, fieldIndex := range t.indexes {
		for value, ids := range fieldIndex {
			newIds := make([]string, 0, len(ids))
			for _, existingID := range ids {
				if existingID != id {
					newIds = append(newIds, existingID)
				}
			}
			if len(newIds) == 0 {
				delete(fieldIndex, value)
			} else {
				t.indexes[fieldName][value] = newIds
			}
		}
	}
}
