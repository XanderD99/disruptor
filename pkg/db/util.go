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

// Auto-detect table name from type
func getTableName[T any](entity T) string {
	// Check if entity implements TableNamer
	if tn, ok := any(entity).(TableNamer); ok {
		return tn.GetTable()
	}

	// Auto-generate from type name
	t := reflect.TypeOf(entity)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Convert "Guild" -> "guilds", "User" -> "users"
	name := strings.ToLower(t.Name())
	if !strings.HasSuffix(name, "s") {
		name += "s"
	}

	return name
}

// Fast path for interface-implementing entities
func GetEntityID(entity any) (any, error) {
	if e, ok := entity.(Identifiable); ok {
		return e.GetID(), nil
	}
	return getEntityIDCached(entity)
}

// Cached reflection as fallback
var typeInfoCache = sync.Map{}

type cachedTypeInfo struct {
	idFieldIndex int
	hasIDField   bool
}

func getEntityIDCached(entity any) (any, error) {
	v := reflect.ValueOf(entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	t := v.Type()

	if cached, ok := typeInfoCache.Load(t); ok {
		info, ok := cached.(cachedTypeInfo)
		if !ok {
			return nil, fmt.Errorf("invalid cache type")
		}
		if !info.hasIDField {
			return nil, fmt.Errorf("ID field not found")
		}
		return v.Field(info.idFieldIndex).Interface(), nil
	}

	// Build cache
	info := cachedTypeInfo{}
	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).Name == "ID" {
			info.idFieldIndex = i
			info.hasIDField = true
			break
		}
	}

	typeInfoCache.Store(t, info)

	if !info.hasIDField {
		return nil, fmt.Errorf("ID field not found")
	}

	return v.Field(info.idFieldIndex).Interface(), nil
}
