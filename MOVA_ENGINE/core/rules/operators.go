package rules

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// OperatorFunc defines the signature for operator functions
type OperatorFunc func(fieldValue, conditionValue interface{}) (bool, error)

// registerDefaultOperators registers all default operators
func (e *Engine) registerDefaultOperators() {
	e.operators[OpEquals] = opEquals
	e.operators[OpNotEquals] = opNotEquals
	e.operators[OpGreater] = opGreater
	e.operators[OpGreaterEq] = opGreaterEq
	e.operators[OpLess] = opLess
	e.operators[OpLessEq] = opLessEq
	e.operators[OpContains] = opContains
	e.operators[OpNotContains] = opNotContains
	e.operators[OpRegex] = opRegex
	e.operators[OpIn] = opIn
	e.operators[OpNotIn] = opNotIn
	e.operators[OpExists] = opExists
	e.operators[OpNotExists] = opNotExists
}

// opEquals checks if field value equals condition value
func opEquals(fieldValue, conditionValue interface{}) (bool, error) {
	return compareValues(fieldValue, conditionValue) == 0, nil
}

// opNotEquals checks if field value does not equal condition value
func opNotEquals(fieldValue, conditionValue interface{}) (bool, error) {
	return compareValues(fieldValue, conditionValue) != 0, nil
}

// opGreater checks if field value is greater than condition value
func opGreater(fieldValue, conditionValue interface{}) (bool, error) {
	cmp := compareValues(fieldValue, conditionValue)
	if cmp == -999 { // Error value
		return false, fmt.Errorf("cannot compare values of different types")
	}
	return cmp > 0, nil
}

// opGreaterEq checks if field value is greater than or equal to condition value
func opGreaterEq(fieldValue, conditionValue interface{}) (bool, error) {
	cmp := compareValues(fieldValue, conditionValue)
	if cmp == -999 { // Error value
		return false, fmt.Errorf("cannot compare values of different types")
	}
	return cmp >= 0, nil
}

// opLess checks if field value is less than condition value
func opLess(fieldValue, conditionValue interface{}) (bool, error) {
	cmp := compareValues(fieldValue, conditionValue)
	if cmp == -999 { // Error value
		return false, fmt.Errorf("cannot compare values of different types")
	}
	return cmp < 0, nil
}

// opLessEq checks if field value is less than or equal to condition value
func opLessEq(fieldValue, conditionValue interface{}) (bool, error) {
	cmp := compareValues(fieldValue, conditionValue)
	if cmp == -999 { // Error value
		return false, fmt.Errorf("cannot compare values of different types")
	}
	return cmp <= 0, nil
}

// opContains checks if field value contains condition value (for strings and arrays)
func opContains(fieldValue, conditionValue interface{}) (bool, error) {
	if fieldValue == nil {
		return false, nil
	}

	fieldStr := toString(fieldValue)
	conditionStr := toString(conditionValue)

	return strings.Contains(fieldStr, conditionStr), nil
}

// opNotContains checks if field value does not contain condition value
func opNotContains(fieldValue, conditionValue interface{}) (bool, error) {
	contains, err := opContains(fieldValue, conditionValue)
	return !contains, err
}

// opRegex checks if field value matches the regex pattern
func opRegex(fieldValue, conditionValue interface{}) (bool, error) {
	if fieldValue == nil {
		return false, nil
	}

	fieldStr := toString(fieldValue)
	patternStr := toString(conditionValue)

	matched, err := regexp.MatchString(patternStr, fieldStr)
	if err != nil {
		return false, fmt.Errorf("invalid regex pattern '%s': %s", patternStr, err.Error())
	}

	return matched, nil
}

// opIn checks if field value is in the list of condition values
func opIn(fieldValue, conditionValue interface{}) (bool, error) {
	if fieldValue == nil {
		return false, nil
	}

	// Convert condition value to slice
	conditionSlice, ok := conditionValue.([]interface{})
	if !ok {
		// Try to convert single value to slice
		conditionSlice = []interface{}{conditionValue}
	}

	for _, item := range conditionSlice {
		if compareValues(fieldValue, item) == 0 {
			return true, nil
		}
	}

	return false, nil
}

// opNotIn checks if field value is not in the list of condition values
func opNotIn(fieldValue, conditionValue interface{}) (bool, error) {
	in, err := opIn(fieldValue, conditionValue)
	return !in, err
}

// opExists checks if field exists (is not nil)
func opExists(fieldValue, conditionValue interface{}) (bool, error) {
	return fieldValue != nil, nil
}

// opNotExists checks if field does not exist (is nil)
func opNotExists(fieldValue, conditionValue interface{}) (bool, error) {
	return fieldValue == nil, nil
}

// compareValues compares two values of potentially different types
// Returns: -1 if a < b, 0 if a == b, 1 if a > b, -999 if incomparable
func compareValues(a, b interface{}) int {
	if a == nil && b == nil {
		return 0
	}
	if a == nil {
		return -1
	}
	if b == nil {
		return 1
	}

	// Try direct comparison first
	if reflect.DeepEqual(a, b) {
		return 0
	}

	// Convert to comparable types
	aVal := reflect.ValueOf(a)
	bVal := reflect.ValueOf(b)

	// Handle numeric comparisons
	if isNumeric(aVal) && isNumeric(bVal) {
		aFloat := toFloat64(a)
		bFloat := toFloat64(b)

		if aFloat < bFloat {
			return -1
		} else if aFloat > bFloat {
			return 1
		}
		return 0
	}

	// Handle string comparisons
	if aVal.Kind() == reflect.String && bVal.Kind() == reflect.String {
		aStr := aVal.String()
		bStr := bVal.String()

		if aStr < bStr {
			return -1
		} else if aStr > bStr {
			return 1
		}
		return 0
	}

	// Convert both to strings as fallback
	aStr := toString(a)
	bStr := toString(b)

	if aStr < bStr {
		return -1
	} else if aStr > bStr {
		return 1
	}
	return 0
}

// isNumeric checks if a value is numeric
func isNumeric(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

// toFloat64 converts a numeric value to float64
func toFloat64(v interface{}) float64 {
	val := reflect.ValueOf(v)

	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(val.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(val.Uint())
	case reflect.Float32, reflect.Float64:
		return val.Float()
	}

	// Try string conversion
	if str, ok := v.(string); ok {
		if f, err := strconv.ParseFloat(str, 64); err == nil {
			return f
		}
	}

	return 0
}

// toString converts any value to string
func toString(v interface{}) string {
	if v == nil {
		return ""
	}

	if str, ok := v.(string); ok {
		return str
	}

	return fmt.Sprintf("%v", v)
}
