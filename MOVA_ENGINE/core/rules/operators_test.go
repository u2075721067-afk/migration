package rules

import (
	"testing"
)

func TestOperators(t *testing.T) {
	tests := []struct {
		name           string
		operator       OperatorFunc
		fieldValue     interface{}
		conditionValue interface{}
		expected       bool
		hasError       bool
	}{
		// Equals operator tests
		{
			name:           "equals string match",
			operator:       opEquals,
			fieldValue:     "test",
			conditionValue: "test",
			expected:       true,
			hasError:       false,
		},
		{
			name:           "equals string no match",
			operator:       opEquals,
			fieldValue:     "test",
			conditionValue: "other",
			expected:       false,
			hasError:       false,
		},
		{
			name:           "equals number match",
			operator:       opEquals,
			fieldValue:     42,
			conditionValue: 42,
			expected:       true,
			hasError:       false,
		},
		{
			name:           "equals nil values",
			operator:       opEquals,
			fieldValue:     nil,
			conditionValue: nil,
			expected:       true,
			hasError:       false,
		},

		// Not equals operator tests
		{
			name:           "not equals string match",
			operator:       opNotEquals,
			fieldValue:     "test",
			conditionValue: "other",
			expected:       true,
			hasError:       false,
		},
		{
			name:           "not equals string no match",
			operator:       opNotEquals,
			fieldValue:     "test",
			conditionValue: "test",
			expected:       false,
			hasError:       false,
		},

		// Greater than operator tests
		{
			name:           "greater than number true",
			operator:       opGreater,
			fieldValue:     10,
			conditionValue: 5,
			expected:       true,
			hasError:       false,
		},
		{
			name:           "greater than number false",
			operator:       opGreater,
			fieldValue:     5,
			conditionValue: 10,
			expected:       false,
			hasError:       false,
		},
		{
			name:           "greater than equal false",
			operator:       opGreater,
			fieldValue:     5,
			conditionValue: 5,
			expected:       false,
			hasError:       false,
		},
		{
			name:           "greater than string true",
			operator:       opGreater,
			fieldValue:     "zebra",
			conditionValue: "apple",
			expected:       true,
			hasError:       false,
		},

		// Greater than or equal operator tests
		{
			name:           "greater than or equal number true",
			operator:       opGreaterEq,
			fieldValue:     10,
			conditionValue: 5,
			expected:       true,
			hasError:       false,
		},
		{
			name:           "greater than or equal equal true",
			operator:       opGreaterEq,
			fieldValue:     5,
			conditionValue: 5,
			expected:       true,
			hasError:       false,
		},
		{
			name:           "greater than or equal false",
			operator:       opGreaterEq,
			fieldValue:     3,
			conditionValue: 5,
			expected:       false,
			hasError:       false,
		},

		// Less than operator tests
		{
			name:           "less than number true",
			operator:       opLess,
			fieldValue:     5,
			conditionValue: 10,
			expected:       true,
			hasError:       false,
		},
		{
			name:           "less than number false",
			operator:       opLess,
			fieldValue:     10,
			conditionValue: 5,
			expected:       false,
			hasError:       false,
		},
		{
			name:           "less than equal false",
			operator:       opLess,
			fieldValue:     5,
			conditionValue: 5,
			expected:       false,
			hasError:       false,
		},

		// Less than or equal operator tests
		{
			name:           "less than or equal number true",
			operator:       opLessEq,
			fieldValue:     5,
			conditionValue: 10,
			expected:       true,
			hasError:       false,
		},
		{
			name:           "less than or equal equal true",
			operator:       opLessEq,
			fieldValue:     5,
			conditionValue: 5,
			expected:       true,
			hasError:       false,
		},
		{
			name:           "less than or equal false",
			operator:       opLessEq,
			fieldValue:     10,
			conditionValue: 5,
			expected:       false,
			hasError:       false,
		},

		// Contains operator tests
		{
			name:           "contains string true",
			operator:       opContains,
			fieldValue:     "hello world",
			conditionValue: "world",
			expected:       true,
			hasError:       false,
		},
		{
			name:           "contains string false",
			operator:       opContains,
			fieldValue:     "hello world",
			conditionValue: "xyz",
			expected:       false,
			hasError:       false,
		},
		{
			name:           "contains nil field",
			operator:       opContains,
			fieldValue:     nil,
			conditionValue: "test",
			expected:       false,
			hasError:       false,
		},

		// Not contains operator tests
		{
			name:           "not contains string true",
			operator:       opNotContains,
			fieldValue:     "hello world",
			conditionValue: "xyz",
			expected:       true,
			hasError:       false,
		},
		{
			name:           "not contains string false",
			operator:       opNotContains,
			fieldValue:     "hello world",
			conditionValue: "world",
			expected:       false,
			hasError:       false,
		},

		// Regex operator tests
		{
			name:           "regex match true",
			operator:       opRegex,
			fieldValue:     "test@example.com",
			conditionValue: `^[^@]+@[^@]+\.[^@]+$`,
			expected:       true,
			hasError:       false,
		},
		{
			name:           "regex match false",
			operator:       opRegex,
			fieldValue:     "invalid-email",
			conditionValue: `^[^@]+@[^@]+\.[^@]+$`,
			expected:       false,
			hasError:       false,
		},
		{
			name:           "regex invalid pattern",
			operator:       opRegex,
			fieldValue:     "test",
			conditionValue: `[`,
			expected:       false,
			hasError:       true,
		},
		{
			name:           "regex nil field",
			operator:       opRegex,
			fieldValue:     nil,
			conditionValue: `test`,
			expected:       false,
			hasError:       false,
		},

		// In operator tests
		{
			name:           "in list true",
			operator:       opIn,
			fieldValue:     "apple",
			conditionValue: []interface{}{"apple", "banana", "cherry"},
			expected:       true,
			hasError:       false,
		},
		{
			name:           "in list false",
			operator:       opIn,
			fieldValue:     "grape",
			conditionValue: []interface{}{"apple", "banana", "cherry"},
			expected:       false,
			hasError:       false,
		},
		{
			name:           "in single value true",
			operator:       opIn,
			fieldValue:     "apple",
			conditionValue: "apple",
			expected:       true,
			hasError:       false,
		},
		{
			name:           "in nil field",
			operator:       opIn,
			fieldValue:     nil,
			conditionValue: []interface{}{"apple", "banana"},
			expected:       false,
			hasError:       false,
		},

		// Not in operator tests
		{
			name:           "not in list true",
			operator:       opNotIn,
			fieldValue:     "grape",
			conditionValue: []interface{}{"apple", "banana", "cherry"},
			expected:       true,
			hasError:       false,
		},
		{
			name:           "not in list false",
			operator:       opNotIn,
			fieldValue:     "apple",
			conditionValue: []interface{}{"apple", "banana", "cherry"},
			expected:       false,
			hasError:       false,
		},

		// Exists operator tests
		{
			name:           "exists true",
			operator:       opExists,
			fieldValue:     "some value",
			conditionValue: nil,
			expected:       true,
			hasError:       false,
		},
		{
			name:           "exists false",
			operator:       opExists,
			fieldValue:     nil,
			conditionValue: nil,
			expected:       false,
			hasError:       false,
		},

		// Not exists operator tests
		{
			name:           "not exists true",
			operator:       opNotExists,
			fieldValue:     nil,
			conditionValue: nil,
			expected:       true,
			hasError:       false,
		},
		{
			name:           "not exists false",
			operator:       opNotExists,
			fieldValue:     "some value",
			conditionValue: nil,
			expected:       false,
			hasError:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.operator(tt.fieldValue, tt.conditionValue)

			if tt.hasError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.hasError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestCompareValues(t *testing.T) {
	tests := []struct {
		name     string
		a        interface{}
		b        interface{}
		expected int
	}{
		{
			name:     "nil values equal",
			a:        nil,
			b:        nil,
			expected: 0,
		},
		{
			name:     "nil vs value",
			a:        nil,
			b:        "test",
			expected: -1,
		},
		{
			name:     "value vs nil",
			a:        "test",
			b:        nil,
			expected: 1,
		},
		{
			name:     "equal strings",
			a:        "test",
			b:        "test",
			expected: 0,
		},
		{
			name:     "string a < b",
			a:        "apple",
			b:        "banana",
			expected: -1,
		},
		{
			name:     "string a > b",
			a:        "zebra",
			b:        "apple",
			expected: 1,
		},
		{
			name:     "equal integers",
			a:        42,
			b:        42,
			expected: 0,
		},
		{
			name:     "integer a < b",
			a:        10,
			b:        20,
			expected: -1,
		},
		{
			name:     "integer a > b",
			a:        20,
			b:        10,
			expected: 1,
		},
		{
			name:     "mixed numeric types",
			a:        10,
			b:        10.0,
			expected: 0,
		},
		{
			name:     "float comparison",
			a:        3.14,
			b:        2.71,
			expected: 1,
		},
		{
			name:     "string comparison",
			a:        "10",
			b:        "5",
			expected: -1, // "10" < "5" lexicographically
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareValues(tt.a, tt.b)

			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestToString(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "nil value",
			input:    nil,
			expected: "",
		},
		{
			name:     "string value",
			input:    "test",
			expected: "test",
		},
		{
			name:     "integer value",
			input:    42,
			expected: "42",
		},
		{
			name:     "float value",
			input:    3.14,
			expected: "3.14",
		},
		{
			name:     "boolean value",
			input:    true,
			expected: "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toString(tt.input)

			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestToFloat64(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected float64
	}{
		{
			name:     "integer value",
			input:    42,
			expected: 42.0,
		},
		{
			name:     "float32 value",
			input:    float32(3.14),
			expected: 3.140000104904175, // float32 precision
		},
		{
			name:     "float64 value",
			input:    3.14159,
			expected: 3.14159,
		},
		{
			name:     "string number",
			input:    "42.5",
			expected: 42.5,
		},
		{
			name:     "invalid string",
			input:    "not a number",
			expected: 0,
		},
		{
			name:     "nil value",
			input:    nil,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toFloat64(tt.input)

			// Use a small epsilon for float comparison
			epsilon := 1e-10
			if (result-tt.expected) > epsilon || (tt.expected-result) > epsilon {
				t.Errorf("expected %f, got %f", tt.expected, result)
			}
		})
	}
}
