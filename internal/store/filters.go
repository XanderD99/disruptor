package store

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Operator represents a Strapi filter operator
type Operator string

const (
	OpEqual                  Operator = "$eq"
	OpEqualInsensitive       Operator = "$eqi"
	OpNotEqual               Operator = "$ne"
	OpNotEqualInsensitive    Operator = "$nei"
	OpLessThan               Operator = "$lt"
	OpLessThanEqual          Operator = "$lte"
	OpGreaterThan            Operator = "$gt"
	OpGreaterThanEqual       Operator = "$gte"
	OpIn                     Operator = "$in"
	OpNotIn                  Operator = "$notIn"
	OpContains               Operator = "$contains"
	OpNotContains            Operator = "$notContains"
	OpContainsInsensitive    Operator = "$containsi"
	OpNotContainsInsensitive Operator = "$notContainsi"
	OpNull                   Operator = "$null"
	OpNotNull                Operator = "$notNull"
	OpBetween                Operator = "$between"
	OpStartsWith             Operator = "$startsWith"
	OpStartsWithInsensitive  Operator = "$startsWithi"
	OpEndsWith               Operator = "$endsWith"
	OpEndsWithInsensitive    Operator = "$endsWithi"
	OpOr                     Operator = "$or"
	OpAnd                    Operator = "$and"
	OpNot                    Operator = "$not"
)

// FilterCondition represents a single filter condition
type FilterCondition struct {
	Field    string
	Operator Operator
	Value    interface{}
}

// FilterGroup represents a group of conditions with logical operators
type FilterGroup struct {
	Operator   Operator // $or, $and, $not
	Conditions []FilterCondition
	Groups     []FilterGroup
}

// Filter represents the complete filter structure
type Filter struct {
	Conditions []FilterCondition
	Groups     []FilterGroup
}

type FilterOption func(filter *Filter)

// NewFilter creates a new StrapiFilter
func NewFilter(filters ...FilterOption) Filter {
	filter := Filter{
		Conditions: make([]FilterCondition, 0),
		Groups:     make([]FilterGroup, 0),
	}

	for _, opt := range filters {
		opt(&filter)
	}

	return filter
}

// WithCondition creates a filter condition for a specific field, operator, and value
func WithCondition(field string, operator Operator, value interface{}) FilterOption {
	return func(filter *Filter) {
		filter.Conditions = append(filter.Conditions, FilterCondition{
			Field:    field,
			Operator: operator,
			Value:    value,
		})
	}
}

// WithGroup creates a filter group with logical operator and conditions
func WithGroup(operator Operator, conditions []FilterCondition) FilterOption {
	return func(filter *Filter) {
		filter.Groups = append(filter.Groups, FilterGroup{
			Operator:   operator,
			Conditions: conditions,
			Groups:     make([]FilterGroup, 0),
		})
	}
}

// AddCondition adds a simple filter condition
func (f *Filter) AddCondition(field string, operator Operator, value interface{}) *Filter {
	f.Conditions = append(f.Conditions, FilterCondition{
		Field:    field,
		Operator: operator,
		Value:    value,
	})
	return f
}

// AddGroup adds a filter group with logical operator
func (f *Filter) AddGroup(operator Operator, conditions []FilterCondition) *Filter {
	f.Groups = append(f.Groups, FilterGroup{
		Operator:   operator,
		Conditions: conditions,
		Groups:     make([]FilterGroup, 0),
	})
	return f
}

// ToQueryParams converts the filter to URL query parameters
func (f *Filter) ToQueryParams() map[string]string {
	params := make(map[string]string)
	f.buildQueryParams(params, "filters", f)
	return params
}

func (f *Filter) ToQueryString() string {
	params := f.ToQueryParams()
	query := make([]string, 0, len(params))
	for key, value := range params {
		query = append(query, fmt.Sprintf("%s=%s", key, value))
	}
	return strings.Join(query, "&")
}

// buildQueryParams recursively builds query parameters
func (f *Filter) buildQueryParams(params map[string]string, prefix string, filter *Filter) {
	// Add simple conditions
	for _, condition := range filter.Conditions {
		key := fmt.Sprintf("[%s]", condition.Field)
		if split := strings.Split(condition.Field, "."); len(split) > 1 {
			key = fmt.Sprintf("[%s]", split[0])
			for _, part := range split[1:] {
				key += fmt.Sprintf("[%s]", part)
			}
		}

		key = fmt.Sprintf("%s%s[%s]", prefix, key, condition.Operator)
		params[key] = f.formatValue(condition.Value)
	}

	// Add groups
	for i, group := range filter.Groups {
		groupPrefix := fmt.Sprintf("%s[%s][%d]", prefix, group.Operator, i)
		f.buildGroupParams(params, groupPrefix, &group)
	}
}

// buildGroupParams builds parameters for a filter group
func (f *Filter) buildGroupParams(params map[string]string, prefix string, group *FilterGroup) {
	for i, condition := range group.Conditions {
		key := fmt.Sprintf("%s[%d][%s][%s]", prefix, i, condition.Field, condition.Operator)
		params[key] = f.formatValue(condition.Value)
	}

	for i, subGroup := range group.Groups {
		subPrefix := fmt.Sprintf("%s[%s][%d]", prefix, subGroup.Operator, i)
		f.buildGroupParams(params, subPrefix, &subGroup)
	}
}

// formatValue converts a value to string for query parameters
func (f *Filter) formatValue(value interface{}) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%g", v)
	case bool:
		return strconv.FormatBool(v)
	case []string:
		return strings.Join(v, ",")
	case []int:
		strs := make([]string, len(v))
		for i, val := range v {
			strs[i] = strconv.Itoa(val)
		}
		return strings.Join(strs, ",")
	default:
		// Use reflection for other slice types
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Slice {
			strs := make([]string, rv.Len())
			for i := 0; i < rv.Len(); i++ {
				strs[i] = fmt.Sprintf("%v", rv.Index(i).Interface())
			}
			return strings.Join(strs, ",")
		}
		return fmt.Sprintf("%v", value)
	}
}

// Helper methods for common operations

// Equal creates an equality filter
func Equal(field string, value interface{}) FilterCondition {
	return FilterCondition{Field: field, Operator: OpEqual, Value: value}
}

// Contains creates a contains filter
func Contains(field string, value string) FilterCondition {
	return FilterCondition{Field: field, Operator: OpContains, Value: value}
}

// In creates an "in array" filter
func In(field string, values interface{}) FilterCondition {
	return FilterCondition{Field: field, Operator: OpIn, Value: values}
}

// GreaterThan creates a greater than filter
func GreaterThan(field string, value interface{}) FilterCondition {
	return FilterCondition{Field: field, Operator: OpGreaterThan, Value: value}
}

// LessThan creates a less than filter
func LessThan(field string, value interface{}) FilterCondition {
	return FilterCondition{Field: field, Operator: OpLessThan, Value: value}
}
