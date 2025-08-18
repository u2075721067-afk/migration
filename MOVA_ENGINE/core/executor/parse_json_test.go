package executor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecuteParseJSON(t *testing.T) {
	executor := NewExecutor()
	ctx := context.Background()

	tests := []struct {
		name           string
		action         Action
		execCtx        *ExecutionContext
		expectedStatus ActionStatus
		expectedError  string
	}{
		{
			name: "parse simple JSONPath",
			action: Action{
				Name: "test_parse",
				Type: "parse_json",
				Config: map[string]interface{}{
					"jsonpath": "$.name",
					"source":   "test_data",
					"variable": "result",
				},
			},
			execCtx: &ExecutionContext{
				Variables: map[string]interface{}{
					"test_data": map[string]interface{}{
						"name": "John",
						"age":  30,
					},
				},
			},
			expectedStatus: ActionStatusCompleted,
		},
		{
			name: "parse array with wildcard",
			action: Action{
				Name: "test_parse_array",
				Type: "parse_json",
				Config: map[string]interface{}{
					"jsonpath": "$.items[*].name",
					"source":   "test_data",
					"variable": "names",
				},
			},
			execCtx: &ExecutionContext{
				Variables: map[string]interface{}{
					"test_data": map[string]interface{}{
						"items": []interface{}{
							map[string]interface{}{"name": "Item1", "id": 1},
							map[string]interface{}{"name": "Item2", "id": 2},
						},
					},
				},
			},
			expectedStatus: ActionStatusCompleted,
		},
		{
			name: "missing config",
			action: Action{
				Name: "test_missing_config",
				Type: "parse_json",
			},
			execCtx:        &ExecutionContext{},
			expectedStatus: ActionStatusFailed,
			expectedError:  "invalid config: config is required",
		},
		{
			name: "missing jsonpath",
			action: Action{
				Name: "test_missing_jsonpath",
				Type: "parse_json",
				Config: map[string]interface{}{
					"source": "test_data",
				},
			},
			execCtx:        &ExecutionContext{},
			expectedStatus: ActionStatusFailed,
			expectedError:  "invalid config: 'jsonpath' field is required and must be string",
		},
		{
			name: "invalid JSONPath",
			action: Action{
				Name: "test_invalid_jsonpath",
				Type: "parse_json",
				Config: map[string]interface{}{
					"jsonpath": "$[invalid]",
					"source":   "test_data",
				},
			},
			execCtx: &ExecutionContext{
				Variables: map[string]interface{}{
					"test_data": map[string]interface{}{"name": "test"},
				},
			},
			expectedStatus: ActionStatusFailed,
			expectedError:  "JSONPath evaluation failed:",
		},
		{
			name: "source not found",
			action: Action{
				Name: "test_source_not_found",
				Type: "parse_json",
				Config: map[string]interface{}{
					"jsonpath": "$.name",
					"source":   "nonexistent",
				},
			},
			execCtx:        &ExecutionContext{},
			expectedStatus: ActionStatusFailed,
			expectedError:  "source data not found: nonexistent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := executor.executeParseJSON(ctx, tt.execCtx, tt.action)

			assert.Equal(t, tt.expectedStatus, result.Status)
			if tt.expectedError != "" {
				assert.Contains(t, result.Error, tt.expectedError)
			}

			if result.Status == ActionStatusCompleted {
				assert.NotNil(t, result.Output)
				assert.NotNil(t, result.EndTime)
			}
		})
	}
}

func TestExecuteParseJSONWithLastResult(t *testing.T) {
	executor := NewExecutor()
	ctx := context.Background()

	// Create execution context with previous results
	execCtx := &ExecutionContext{
		Variables: map[string]interface{}{},
		Results: map[string]ActionResult{
			"previous_action": {
				ActionName: "previous_action",
				Status:     ActionStatusCompleted,
				Output: map[string]interface{}{
					"data": map[string]interface{}{
						"users": []interface{}{
							map[string]interface{}{"name": "Alice", "id": 1},
							map[string]interface{}{"name": "Bob", "id": 2},
						},
					},
				},
			},
		},
	}

	action := Action{
		Name: "parse_from_last_result",
		Type: "parse_json",
		Config: map[string]interface{}{
			"jsonpath": "$.data.users[*].name",
			"variable": "user_names",
		},
	}

	result := executor.executeParseJSON(ctx, execCtx, action)

	assert.Equal(t, ActionStatusCompleted, result.Status)
	assert.NotNil(t, result.Output)
	assert.Contains(t, result.Output, "result")

	// Check that variable was set
	userNames, exists := execCtx.Variables["user_names"]
	assert.True(t, exists)
	assert.NotNil(t, userNames)
}
