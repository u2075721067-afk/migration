package rules

import (
	"encoding/json"
	"testing"
)

func TestActionSetVar(t *testing.T) {
	ctx := Context{
		Variables: make(map[string]interface{}),
		Request:   make(map[string]interface{}),
		Response:  make(map[string]interface{}),
		Metadata:  make(map[string]interface{}),
	}

	tests := []struct {
		name     string
		params   map[string]interface{}
		hasError bool
	}{
		{
			name: "set string variable",
			params: map[string]interface{}{
				"variable": "test_var",
				"value":    "test_value",
			},
			hasError: false,
		},
		{
			name: "set number variable",
			params: map[string]interface{}{
				"variable": "count",
				"value":    42,
			},
			hasError: false,
		},
		{
			name: "missing variable parameter",
			params: map[string]interface{}{
				"value": "test_value",
			},
			hasError: true,
		},
		{
			name: "missing value parameter",
			params: map[string]interface{}{
				"variable": "test_var",
			},
			hasError: true,
		},
		{
			name: "non-string variable name",
			params: map[string]interface{}{
				"variable": 123,
				"value":    "test_value",
			},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := actionSetVar(tt.params, ctx)

			if tt.hasError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.hasError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.hasError {
				if result == nil {
					t.Error("expected result but got nil")
					return
				}

				if variable, ok := result["variable"].(string); ok {
					expectedVar := tt.params["variable"].(string)
					if variable != expectedVar {
						t.Errorf("expected variable=%s, got variable=%s", expectedVar, variable)
					}

					// Check if variable was actually set in context
					if ctx.Variables[variable] != tt.params["value"] {
						t.Errorf("variable not set correctly in context")
					}
				} else {
					t.Error("result should contain variable field")
				}
			}
		})
	}
}

func TestActionRetry(t *testing.T) {
	ctx := Context{
		Variables: make(map[string]interface{}),
		Request:   make(map[string]interface{}),
		Response:  make(map[string]interface{}),
		Metadata:  make(map[string]interface{}),
	}

	tests := []struct {
		name     string
		params   map[string]interface{}
		hasError bool
	}{
		{
			name: "retry with all parameters",
			params: map[string]interface{}{
				"profile":      "aggressive",
				"max_attempts": 5.0,
				"delay":        2000.0,
			},
			hasError: false,
		},
		{
			name:     "retry with default parameters",
			params:   map[string]interface{}{},
			hasError: false,
		},
		{
			name: "retry with partial parameters",
			params: map[string]interface{}{
				"profile": "conservative",
			},
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := actionRetry(tt.params, ctx)

			if tt.hasError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.hasError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.hasError {
				if result == nil {
					t.Error("expected result but got nil")
					return
				}

				if action, ok := result["action"].(string); !ok || action != "retry" {
					t.Error("result should contain action=retry")
				}

				if profile, ok := result["profile"].(string); ok {
					if expectedProfile, exists := tt.params["profile"].(string); exists {
						if profile != expectedProfile {
							t.Errorf("expected profile=%s, got profile=%s", expectedProfile, profile)
						}
					} else {
						if profile != "default" {
							t.Errorf("expected default profile, got profile=%s", profile)
						}
					}
				}
			}
		})
	}
}

func TestActionLog(t *testing.T) {
	ctx := Context{
		Variables: make(map[string]interface{}),
		Request:   make(map[string]interface{}),
		Response:  make(map[string]interface{}),
		Metadata:  make(map[string]interface{}),
	}

	tests := []struct {
		name     string
		params   map[string]interface{}
		hasError bool
	}{
		{
			name: "log with message and level",
			params: map[string]interface{}{
				"message": "test message",
				"level":   "error",
			},
			hasError: false,
		},
		{
			name: "log with message only",
			params: map[string]interface{}{
				"message": "test message",
			},
			hasError: false,
		},
		{
			name:     "log without message",
			params:   map[string]interface{}{},
			hasError: true,
		},
		{
			name: "log with non-string message",
			params: map[string]interface{}{
				"message": 123,
			},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := actionLog(tt.params, ctx)

			if tt.hasError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.hasError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.hasError {
				if result == nil {
					t.Error("expected result but got nil")
					return
				}

				if action, ok := result["action"].(string); !ok || action != "log" {
					t.Error("result should contain action=log")
				}

				if message, ok := result["message"].(string); ok {
					expectedMessage := tt.params["message"].(string)
					if message != expectedMessage {
						t.Errorf("expected message=%s, got message=%s", expectedMessage, message)
					}
				} else {
					t.Error("result should contain message field")
				}

				if level, ok := result["level"].(string); ok {
					if expectedLevel, exists := tt.params["level"].(string); exists {
						if level != expectedLevel {
							t.Errorf("expected level=%s, got level=%s", expectedLevel, level)
						}
					} else {
						if level != "info" {
							t.Errorf("expected default level=info, got level=%s", level)
						}
					}
				}
			}
		})
	}
}

func TestActionSkip(t *testing.T) {
	ctx := Context{
		Variables: make(map[string]interface{}),
		Request:   make(map[string]interface{}),
		Response:  make(map[string]interface{}),
		Metadata:  make(map[string]interface{}),
	}

	tests := []struct {
		name     string
		params   map[string]interface{}
		hasError bool
	}{
		{
			name: "skip with reason",
			params: map[string]interface{}{
				"reason": "maintenance window",
			},
			hasError: false,
		},
		{
			name:     "skip without reason",
			params:   map[string]interface{}{},
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := actionSkip(tt.params, ctx)

			if tt.hasError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.hasError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.hasError {
				if result == nil {
					t.Error("expected result but got nil")
					return
				}

				if action, ok := result["action"].(string); !ok || action != "skip" {
					t.Error("result should contain action=skip")
				}

				if reason, ok := result["reason"].(string); ok {
					if expectedReason, exists := tt.params["reason"].(string); exists {
						if reason != expectedReason {
							t.Errorf("expected reason=%s, got reason=%s", expectedReason, reason)
						}
					}
				}
			}
		})
	}
}

func TestActionRoute(t *testing.T) {
	ctx := Context{
		Variables: make(map[string]interface{}),
		Request:   make(map[string]interface{}),
		Response:  make(map[string]interface{}),
		Metadata:  make(map[string]interface{}),
	}

	tests := []struct {
		name     string
		params   map[string]interface{}
		hasError bool
	}{
		{
			name: "route with workflow and reason",
			params: map[string]interface{}{
				"workflow": "error_handler",
				"reason":   "critical error detected",
			},
			hasError: false,
		},
		{
			name: "route with workflow only",
			params: map[string]interface{}{
				"workflow": "error_handler",
			},
			hasError: false,
		},
		{
			name:     "route without workflow",
			params:   map[string]interface{}{},
			hasError: true,
		},
		{
			name: "route with non-string workflow",
			params: map[string]interface{}{
				"workflow": 123,
			},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := actionRoute(tt.params, ctx)

			if tt.hasError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.hasError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.hasError {
				if result == nil {
					t.Error("expected result but got nil")
					return
				}

				if action, ok := result["action"].(string); !ok || action != "route" {
					t.Error("result should contain action=route")
				}

				if workflow, ok := result["workflow"].(string); ok {
					expectedWorkflow := tt.params["workflow"].(string)
					if workflow != expectedWorkflow {
						t.Errorf("expected workflow=%s, got workflow=%s", expectedWorkflow, workflow)
					}
				} else {
					t.Error("result should contain workflow field")
				}
			}
		})
	}
}

func TestActionStop(t *testing.T) {
	ctx := Context{
		Variables: make(map[string]interface{}),
		Request:   make(map[string]interface{}),
		Response:  make(map[string]interface{}),
		Metadata:  make(map[string]interface{}),
	}

	tests := []struct {
		name     string
		params   map[string]interface{}
		hasError bool
	}{
		{
			name: "stop with reason",
			params: map[string]interface{}{
				"reason": "maximum retries exceeded",
			},
			hasError: false,
		},
		{
			name:     "stop without reason",
			params:   map[string]interface{}{},
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := actionStop(tt.params, ctx)

			if tt.hasError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.hasError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.hasError {
				if result == nil {
					t.Error("expected result but got nil")
					return
				}

				if action, ok := result["action"].(string); !ok || action != "stop" {
					t.Error("result should contain action=stop")
				}
			}
		})
	}
}

func TestActionTransform(t *testing.T) {
	ctx := Context{
		Variables: map[string]interface{}{
			"input_text": "Hello World",
			"json_str":   `{"name": "test", "value": 123}`,
			"object":     map[string]interface{}{"name": "test", "value": 123},
		},
		Request:  make(map[string]interface{}),
		Response: make(map[string]interface{}),
		Metadata: make(map[string]interface{}),
	}

	tests := []struct {
		name     string
		params   map[string]interface{}
		hasError bool
	}{
		{
			name: "uppercase transform",
			params: map[string]interface{}{
				"type":   "uppercase",
				"source": "input_text",
				"target": "output_text",
			},
			hasError: false,
		},
		{
			name: "lowercase transform",
			params: map[string]interface{}{
				"type":   "lowercase",
				"source": "input_text",
				"target": "output_text",
			},
			hasError: false,
		},
		{
			name: "json_parse transform",
			params: map[string]interface{}{
				"type":   "json_parse",
				"source": "json_str",
				"target": "parsed_json",
			},
			hasError: false,
		},
		{
			name: "json_stringify transform",
			params: map[string]interface{}{
				"type":   "json_stringify",
				"source": "object",
				"target": "json_string",
			},
			hasError: false,
		},
		{
			name: "missing type parameter",
			params: map[string]interface{}{
				"source": "input_text",
				"target": "output_text",
			},
			hasError: true,
		},
		{
			name: "missing source parameter",
			params: map[string]interface{}{
				"type":   "uppercase",
				"target": "output_text",
			},
			hasError: true,
		},
		{
			name: "missing target parameter",
			params: map[string]interface{}{
				"type":   "uppercase",
				"source": "input_text",
			},
			hasError: true,
		},
		{
			name: "unknown transform type",
			params: map[string]interface{}{
				"type":   "unknown_transform",
				"source": "input_text",
				"target": "output_text",
			},
			hasError: true,
		},
		{
			name: "source field not found",
			params: map[string]interface{}{
				"type":   "uppercase",
				"source": "nonexistent_field",
				"target": "output_text",
			},
			hasError: true,
		},
		{
			name: "json_parse with invalid json",
			params: map[string]interface{}{
				"type":   "json_parse",
				"source": "input_text", // "Hello World" is not valid JSON
				"target": "parsed_json",
			},
			hasError: true,
		},
		{
			name: "json_parse with non-string input",
			params: map[string]interface{}{
				"type":   "json_parse",
				"source": "object", // object is not a string
				"target": "parsed_json",
			},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := actionTransform(tt.params, ctx)

			if tt.hasError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.hasError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.hasError {
				if result == nil {
					t.Error("expected result but got nil")
					return
				}

				if action, ok := result["action"].(string); !ok || action != "transform" {
					t.Error("result should contain action=transform")
				}

				transformType := tt.params["type"].(string)
				target := tt.params["target"].(string)

				// Check if the transformation was applied correctly
				switch transformType {
				case "uppercase":
					if ctx.Variables[target] != "HELLO WORLD" {
						t.Errorf("uppercase transform failed, got: %v", ctx.Variables[target])
					}
				case "lowercase":
					if ctx.Variables[target] != "hello world" {
						t.Errorf("lowercase transform failed, got: %v", ctx.Variables[target])
					}
				case "json_parse":
					if parsedObj, ok := ctx.Variables[target].(map[string]interface{}); ok {
						if parsedObj["name"] != "test" || parsedObj["value"].(float64) != 123 {
							t.Errorf("json_parse transform failed, got: %v", ctx.Variables[target])
						}
					} else {
						t.Errorf("json_parse should produce map[string]interface{}, got: %T", ctx.Variables[target])
					}
				case "json_stringify":
					if jsonStr, ok := ctx.Variables[target].(string); ok {
						var parsed map[string]interface{}
						if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
							t.Errorf("json_stringify produced invalid JSON: %s", jsonStr)
						}
					} else {
						t.Errorf("json_stringify should produce string, got: %T", ctx.Variables[target])
					}
				}
			}
		})
	}
}

func TestGetValueFromContext(t *testing.T) {
	ctx := Context{
		Variables: map[string]interface{}{
			"var1": "value1",
		},
		Request: map[string]interface{}{
			"req1": "request_value",
		},
		Response: map[string]interface{}{
			"resp1": "response_value",
		},
		Metadata: map[string]interface{}{
			"meta1": "metadata_value",
		},
	}

	tests := []struct {
		name     string
		field    string
		expected interface{}
	}{
		{
			name:     "get from variables",
			field:    "var1",
			expected: "value1",
		},
		{
			name:     "get from request",
			field:    "req1",
			expected: "request_value",
		},
		{
			name:     "get from response",
			field:    "resp1",
			expected: "response_value",
		},
		{
			name:     "get from metadata",
			field:    "meta1",
			expected: "metadata_value",
		},
		{
			name:     "get nonexistent field",
			field:    "nonexistent",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getValueFromContext(tt.field, ctx)

			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
