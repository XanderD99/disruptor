package db

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// Identifiable interface for fast ID access
type Identifiable interface {
	GetID() any
}

type TableNamer interface {
	GetTable() string
}

// IDInfo contains both the ID value and the key to store it under
type IDInfo struct {
	Value any
	Key   string
}

// Sync pools for reducing allocations (only where beneficial)
var (
	tagKeysMapPool = sync.Pool{
		New: func() interface{} {
			return make(map[string]string, 4) // Pre-allocate for common tags
		},
	}

	// Only use string builder pool for complex operations
	stringBuilderPool = sync.Pool{
		New: func() interface{} {
			sb := &strings.Builder{}
			sb.Grow(32) // Pre-allocate reasonable capacity
			return sb
		},
	}
)

const (
	defaultIDKey = "id"

	// Threshold for when to use string builder pool vs simple concatenation
	maxSimpleStringLength = 20
)

// Auto-detect table name from type
func GetTableName[T any](entity T) string {
	// Fast path for interface-implementing entities (no pool needed)
	if tn, ok := any(entity).(TableNamer); ok {
		return tn.GetTable()
	}

	// Auto-generate from type name
	t := reflect.TypeOf(entity)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Handle slices - extract the element type
	if t.Kind() == reflect.Slice {
		t = t.Elem()
		// If it's a slice of pointers, get the underlying type
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
	}

	name := strings.ToLower(t.Name())

	// For simple cases, avoid pool overhead and use direct string operations
	if len(name) <= maxSimpleStringLength {
		if strings.HasSuffix(name, "s") {
			return name
		}
		return name + "s"
	}

	// For longer strings, use pool to avoid larger allocations
	return buildTableNameWithPool(name)
}

// buildTableNameWithPool uses string builder pool for complex table name generation
func buildTableNameWithPool(name string) string {
	sb, ok := stringBuilderPool.Get().(*strings.Builder)
	if !ok {
		// If pool retrieval failed, fallback to a new instance
		sb = &strings.Builder{}
		sb.Grow(32)
	}
	sb.Reset()
	defer stringBuilderPool.Put(sb)

	sb.WriteString(name)
	if !strings.HasSuffix(name, "s") {
		sb.WriteString("s")
	}

	return sb.String()
}

// GetEntityID extracts ID value and key from entity
// tagName is optional - if provided, it uses the specified struct tag to determine the key
// Supported tags: "bson", "json", "db", etc.
func GetEntityID(entity any, tagName ...string) (string, any, error) {
	// Fast path for interface-implementing entities
	if e, ok := entity.(Identifiable); ok {
		// For interface entities, we still need to determine the key
		var key string
		if len(tagName) > 0 && tagName[0] != "" {
			var err error
			key, err = getIDKeyFromTag(entity, tagName[0])
			if err != nil {
				// Fallback to default key
				key = defaultIDKey
			}
		} else {
			key = defaultIDKey
		}

		return key, e.GetID(), nil
	}

	// Use cached reflection with tag support
	return getEntityIDCachedWithTag(entity, tagName...)
}

// Cached reflection as fallback
var typeInfoCache = sync.Map{}

type cachedTypeInfo struct {
	idFieldIndex int
	hasIDField   bool
	defaultKey   string
	tagKeys      map[string]string // tag name -> key value
}

func getEntityIDCachedWithTag(entity any, tagName ...string) (string, any, error) {
	v, t := getReflectionInfo(entity)
	requestedTag := getRequestedTag(tagName)

	info, err := getOrCreateTypeInfo(t)
	if err != nil {
		return "", nil, err
	}

	if !info.hasIDField {
		return "", nil, fmt.Errorf("ID field not found")
	}

	idValue := v.Field(info.idFieldIndex).Interface()
	key, err := determineIDKey(info, requestedTag)
	if err != nil {
		return "", nil, err
	}

	return key, idValue, nil
}

// getReflectionInfo extracts reflection values from entity
func getReflectionInfo(entity any) (reflect.Value, reflect.Type) {
	v := reflect.ValueOf(entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v, v.Type()
}

// getRequestedTag extracts the requested tag name from variadic parameters
func getRequestedTag(tagName []string) string {
	if len(tagName) > 0 {
		return tagName[0]
	}
	return ""
}

// getOrCreateTypeInfo retrieves cached type info or creates new one
func getOrCreateTypeInfo(t reflect.Type) (cachedTypeInfo, error) {
	cacheKey := t.String()

	if cached, ok := typeInfoCache.Load(cacheKey); ok {
		if cachedInfo, ok := cached.(cachedTypeInfo); ok {
			return cachedInfo, nil
		}
	}

	info := buildTypeInfo(t)
	typeInfoCache.Store(cacheKey, info)
	return info, nil
}

// buildTypeInfo creates type information by scanning fields
func buildTypeInfo(t reflect.Type) cachedTypeInfo {
	// Get a clean map from the pool
	tagKeysMap, ok := tagKeysMapPool.Get().(map[string]string)
	if !ok {
		// If pool retrieval failed, fallback to a new map
		tagKeysMap = make(map[string]string, 4) // Pre-allocate for common tags
	}
	// Clear the map (in case it has leftover data)
	for k := range tagKeysMap {
		delete(tagKeysMap, k)
	}

	info := cachedTypeInfo{
		tagKeys: tagKeysMap, // Use pooled map
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if isIDField(field) {
			info.idFieldIndex = i
			info.hasIDField = true
			info.defaultKey = defaultIDKey
			extractTagKeys(&info, field)
			setDefaultKey(&info)
			break
		}
	}

	// Important: Don't put the map back to pool since it's stored in cache
	// The map will be reused when the cache entry is accessed

	return info
}

// isIDField checks if a field is an ID field
func isIDField(field reflect.StructField) bool {
	return field.Name == "ID" || strings.ToLower(field.Name) == defaultIDKey
}

// extractTagKeys extracts tag information from the ID field
func extractTagKeys(info *cachedTypeInfo, field reflect.StructField) {
	// For common tags, avoid pool overhead with static slice
	commonTags := [4]string{"bson", "json", "xml", "db"}

	for _, tagName := range commonTags {
		if tag := field.Tag.Get(tagName); tag != "" {
			info.tagKeys[tagName] = strings.Split(tag, ",")[0]
		}
	}
}

// setDefaultKey sets the default key based on available tags
func setDefaultKey(info *cachedTypeInfo) {
	// Static priority order for better performance
	priorityOrder := [4]string{"bson", "json", "db", "xml"}

	for _, tagName := range priorityOrder {
		if key, exists := info.tagKeys[tagName]; exists && key != "" {
			info.defaultKey = key
			return
		}
	}
}

// determineIDKey determines which key to use based on requested tag
func determineIDKey(info cachedTypeInfo, requestedTag string) (string, error) {
	if requestedTag == "" {
		return info.defaultKey, nil
	}

	if tagKey, exists := info.tagKeys[requestedTag]; exists && tagKey != "" {
		return tagKey, nil
	}

	return "", fmt.Errorf("tag '%s' not found or empty for ID field", requestedTag)
}

// Helper function to get ID key from a specific tag
func getIDKeyFromTag(entity any, tagName string) (string, error) {
	t := reflect.TypeOf(entity)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Name == "ID" || strings.ToLower(field.Name) == defaultIDKey {
			if tag := field.Tag.Get(tagName); tag != "" {
				// Extract the field name from tag (before any comma)
				return strings.Split(tag, ",")[0], nil
			}
		}
	}

	return "", fmt.Errorf("tag '%s' not found for ID field", tagName)
}

// Helper functions for common use cases
func GetEntityBSONID(entity any) (string, any, error) {
	return GetEntityID(entity, "bson")
}

func GetEntityJSONID(entity any) (string, any, error) {
	return GetEntityID(entity, "json")
}
