package executor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecuteHTTPFetch(t *testing.T) {
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
			name: "basic GET request",
			action: Action{
				Name: "test_get",
				Type: "http_fetch",
				Config: map[string]interface{}{
					"url":     "https://api.example.com/users",
					"method":  "GET",
					"timeout": 30000,
				},
			},
			execCtx:        &ExecutionContext{},
			expectedStatus: ActionStatusCompleted,
		},
		{
			name: "POST request with body",
			action: Action{
				Name: "test_post",
				Type: "http_fetch",
				Config: map[string]interface{}{
					"url":     "https://api.example.com/users",
					"method":  "POST",
					"body":    `{"name": "John", "email": "john@example.com"}`,
					"timeout": 30000,
				},
			},
			execCtx:        &ExecutionContext{},
			expectedStatus: ActionStatusCompleted,
		},
		{
			name: "request with headers",
			action: Action{
				Name: "test_with_headers",
				Type: "http_fetch",
				Config: map[string]interface{}{
					"url":     "https://api.example.com/users",
					"method":  "GET",
					"headers": map[string]interface{}{"Authorization": "Bearer token123"},
					"timeout": 30000,
				},
			},
			execCtx:        &ExecutionContext{},
			expectedStatus: ActionStatusCompleted,
		},
		{
			name: "default method GET",
			action: Action{
				Name: "test_default_method",
				Type: "http_fetch",
				Config: map[string]interface{}{
					"url":     "https://api.example.com/users",
					"timeout": 30000,
				},
			},
			execCtx:        &ExecutionContext{},
			expectedStatus: ActionStatusCompleted,
		},
		{
			name: "default timeout",
			action: Action{
				Name: "test_default_timeout",
				Type: "http_fetch",
				Config: map[string]interface{}{
					"url":    "https://api.example.com/users",
					"method": "GET",
				},
			},
			execCtx:        &ExecutionContext{},
			expectedStatus: ActionStatusCompleted,
		},
		{
			name: "missing config",
			action: Action{
				Name: "test_missing_config",
				Type: "http_fetch",
			},
			execCtx:        &ExecutionContext{},
			expectedStatus: ActionStatusFailed,
			expectedError:  "invalid config: config is required",
		},
		{
			name: "missing URL",
			action: Action{
				Name: "test_missing_url",
				Type: "http_fetch",
				Config: map[string]interface{}{
					"method": "GET",
				},
			},
			execCtx:        &ExecutionContext{},
			expectedStatus: ActionStatusFailed,
			expectedError:  "invalid config: 'url' field is required and must be string",
		},
		{
			name: "empty URL",
			action: Action{
				Name: "test_empty_url",
				Type: "http_fetch",
				Config: map[string]interface{}{
					"url":    "",
					"method": "GET",
				},
			},
			execCtx:        &ExecutionContext{},
			expectedStatus: ActionStatusFailed,
			expectedError:  "invalid config: 'url' field is required and must be string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := executor.executeHTTPFetch(ctx, tt.execCtx, tt.action)

			assert.Equal(t, tt.expectedStatus, result.Status)
			if tt.expectedError != "" {
				assert.Contains(t, result.Error, tt.expectedError)
			}

			if result.Status == ActionStatusCompleted {
				assert.NotNil(t, result.Output)
				assert.NotNil(t, result.EndTime)
				assert.Contains(t, result.Output, "status_code")
				assert.Contains(t, result.Output, "url")
				assert.Contains(t, result.Output, "method")
				assert.Contains(t, result.Output, "response")

				// Check simulated response
				assert.Equal(t, 200, result.Output["status_code"])
				assert.Contains(t, result.Output["response"], "HTTP request simulated successfully")
			}
		})
	}
}

func TestExecuteHTTPFetchLogging(t *testing.T) {
	executor := NewExecutor()
	ctx := context.Background()
	execCtx := &ExecutionContext{}

	action := Action{
		Name: "test_http_logging",
		Type: "http_fetch",
		Config: map[string]interface{}{
			"url":     "https://api.example.com/test",
			"method":  "POST",
			"timeout": 15000,
		},
	}

	initialLogCount := len(execCtx.Logs)
	result := executor.executeHTTPFetch(ctx, execCtx, action)

	assert.Equal(t, ActionStatusCompleted, result.Status)

	// Check that logs were added
	assert.Greater(t, len(execCtx.Logs), initialLogCount)

	// Check for HTTP request log
	foundHTTPLog := false
	for _, log := range execCtx.Logs {
		if log.Action == "test_http_logging" && log.Type == "http_fetch" {
			if log.Message == "HTTP request started" {
				foundHTTPLog = true
				// Check log parameters
				assert.Contains(t, log.ParamsRedacted, "url")
				assert.Contains(t, log.ParamsRedacted, "method")
				assert.Contains(t, log.ParamsRedacted, "timeout")
			}
		}
	}

	assert.True(t, foundHTTPLog, "HTTP request log not found")
}

func TestExecuteHTTPFetchOutputStructure(t *testing.T) {
	executor := NewExecutor()
	ctx := context.Background()
	execCtx := &ExecutionContext{}

	action := Action{
		Name: "test_output_structure",
		Type: "http_fetch",
		Config: map[string]interface{}{
			"url":     "https://api.example.com/users/123",
			"method":  "GET",
			"timeout": 25000,
		},
	}

	result := executor.executeHTTPFetch(ctx, execCtx, action)

	assert.Equal(t, ActionStatusCompleted, result.Status)
	assert.NotNil(t, result.Output)

	// Verify output structure
	output := result.Output
	assert.IsType(t, 0, output["status_code"])
	assert.IsType(t, "", output["url"])
	assert.IsType(t, "", output["method"])
	assert.IsType(t, "", output["response"])

	// Verify specific values
	assert.Equal(t, 200, output["status_code"])
	assert.Equal(t, "https://api.example.com/users/123", output["url"])
	assert.Equal(t, "GET", output["method"])
	assert.Contains(t, output["response"], "simulated successfully")
}
