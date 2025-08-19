package executor

import (
	"context"
	"testing"
	"time"

	"github.com/mova-engine/mova-engine/config"
	"github.com/stretchr/testify/assert"
)

func TestExecuteHTTPFetch(t *testing.T) {
	// Create executor with permissive security config for testing
	securityConfig := &config.SecurityConfig{
		HTTP: config.HTTPSecurityConfig{
			AllowedHosts:    []string{"api.example.com", "httpbin.org", "*.example.com"},
			DeniedHosts:     []string{},
			AllowedPorts:    []int{80, 443, 8080, 8443},
			DeniedPorts:     []int{},
			AllowedSchemes:  []string{"http", "https"},
			DeniedNetworks:  []string{},
			MaxRequestSize:  10 * 1024 * 1024,
			MaxResponseSize: 50 * 1024 * 1024,
			UserAgent:       "MOVA-Engine-Test/1.0",
			FollowRedirects: false,
			MaxRedirects:    0,
		},
		Logging: config.LoggingSecurityConfig{
			RedactSecrets:   true,
			SensitiveKeys:   []string{},
			MaxLogEntrySize: 1024 * 1024,
		},
		Timeouts: config.TimeoutSecurityConfig{
			HTTPTimeout:     30 * time.Second,
			ActionTimeout:   5 * time.Minute,
			WorkflowTimeout: 30 * time.Minute,
		},
	}
	executor := NewExecutorWithConfig(securityConfig)
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
					"url":     "https://httpbin.org/get",
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
					"url":     "https://httpbin.org/post",
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
					"url":     "https://httpbin.org/get",
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
					"url":     "https://httpbin.org/get",
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
					"url":    "https://httpbin.org/get",
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

			if result.Status != tt.expectedStatus {
				t.Logf("Test %s failed. Expected status: %s, Got: %s, Error: %s", tt.name, tt.expectedStatus, result.Status, result.Error)
			}
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
				assert.Contains(t, result.Output, "body")

				// Check real HTTP response
				assert.Equal(t, 200, result.Output["status_code"])
				assert.NotNil(t, result.Output["body"])
			}
		})
	}
}

func TestExecuteHTTPFetchLogging(t *testing.T) {
	// Create executor with permissive security config for testing
	securityConfig := &config.SecurityConfig{
		HTTP: config.HTTPSecurityConfig{
			AllowedHosts:    []string{"api.example.com", "httpbin.org", "*.example.com"},
			DeniedHosts:     []string{},
			AllowedPorts:    []int{80, 443, 8080, 8443},
			DeniedPorts:     []int{},
			AllowedSchemes:  []string{"http", "https"},
			DeniedNetworks:  []string{},
			MaxRequestSize:  10 * 1024 * 1024,
			MaxResponseSize: 50 * 1024 * 1024,
			UserAgent:       "MOVA-Engine-Test/1.0",
			FollowRedirects: false,
			MaxRedirects:    0,
		},
		Logging: config.LoggingSecurityConfig{
			RedactSecrets:   true,
			SensitiveKeys:   []string{},
			MaxLogEntrySize: 1024 * 1024,
		},
		Timeouts: config.TimeoutSecurityConfig{
			HTTPTimeout:     30 * time.Second,
			ActionTimeout:   5 * time.Minute,
			WorkflowTimeout: 30 * time.Minute,
		},
	}
	executor := NewExecutorWithConfig(securityConfig)
	ctx := context.Background()
	execCtx := &ExecutionContext{}

	action := Action{
		Name: "test_http_logging",
		Type: "http_fetch",
		Config: map[string]interface{}{
			"url":     "https://httpbin.org/post",
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
	// Create executor with permissive security config for testing
	securityConfig := &config.SecurityConfig{
		HTTP: config.HTTPSecurityConfig{
			AllowedHosts:    []string{"api.example.com", "httpbin.org", "*.example.com"},
			DeniedHosts:     []string{},
			AllowedPorts:    []int{80, 443, 8080, 8443},
			DeniedPorts:     []int{},
			AllowedSchemes:  []string{"http", "https"},
			DeniedNetworks:  []string{},
			MaxRequestSize:  10 * 1024 * 1024,
			MaxResponseSize: 50 * 1024 * 1024,
			UserAgent:       "MOVA-Engine-Test/1.0",
			FollowRedirects: false,
			MaxRedirects:    0,
		},
		Logging: config.LoggingSecurityConfig{
			RedactSecrets:   true,
			SensitiveKeys:   []string{},
			MaxLogEntrySize: 1024 * 1024,
		},
		Timeouts: config.TimeoutSecurityConfig{
			HTTPTimeout:     30 * time.Second,
			ActionTimeout:   5 * time.Minute,
			WorkflowTimeout: 30 * time.Minute,
		},
	}
	executor := NewExecutorWithConfig(securityConfig)
	ctx := context.Background()
	execCtx := &ExecutionContext{}

	action := Action{
		Name: "test_output_structure",
		Type: "http_fetch",
		Config: map[string]interface{}{
			"url":     "https://httpbin.org/get",
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
	assert.NotNil(t, output["body"])

	// Verify specific values
	assert.Equal(t, 200, output["status_code"])
	assert.Equal(t, "https://httpbin.org/get", output["url"])
	assert.Equal(t, "GET", output["method"])
	assert.NotNil(t, output["body"])
}

func TestExecuteHTTPFetchSecurity(t *testing.T) {
	// Create executor with default security config (restrictive)
	executor := NewExecutor()
	ctx := context.Background()
	execCtx := &ExecutionContext{}

	tests := []struct {
		name          string
		url           string
		expectedError string
	}{
		{
			name:          "blocked localhost",
			url:           "http://localhost:8080/api",
			expectedError: "security validation failed: host localhost is explicitly denied",
		},
		{
			name:          "blocked internal host",
			url:           "http://service.internal/api",
			expectedError: "security validation failed: host service.internal is explicitly denied",
		},
		{
			name:          "blocked port",
			url:           "http://api.github.com:22/",
			expectedError: "security validation failed: port 22 is explicitly denied",
		},
		{
			name:          "blocked scheme",
			url:           "ftp://api.github.com/file",
			expectedError: "security validation failed: scheme ftp is not allowed",
		},
		{
			name:          "blocked metadata service",
			url:           "http://169.254.169.254/metadata",
			expectedError: "security validation failed: host 169.254.169.254 is explicitly denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := Action{
				Name: "test_security_block",
				Type: "http_fetch",
				Config: map[string]interface{}{
					"url":    tt.url,
					"method": "GET",
				},
			}

			result := executor.executeHTTPFetch(ctx, execCtx, action)

			assert.Equal(t, ActionStatusFailed, result.Status)
			assert.Contains(t, result.Error, "security validation failed")
		})
	}
}
