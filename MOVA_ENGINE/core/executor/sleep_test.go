package executor

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExecuteSleep(t *testing.T) {
	executor := NewExecutor()
	ctx := context.Background()

	tests := []struct {
		name           string
		action         Action
		execCtx        *ExecutionContext
		expectedStatus ActionStatus
		expectedError  string
		shouldSleep    bool
	}{
		{
			name: "sleep for 0.1 seconds",
			action: Action{
				Name: "test_sleep",
				Type: "sleep",
				Config: map[string]interface{}{
					"seconds": 0.1,
				},
			},
			execCtx:        &ExecutionContext{},
			expectedStatus: ActionStatusCompleted,
			shouldSleep:    true,
		},
		{
			name: "sleep with timeout check",
			action: Action{
				Name: "test_sleep_timeout",
				Type: "sleep",
				Config: map[string]interface{}{
					"seconds": 0.1,
				},
				Timeout: func() *int { t := 5; return &t }(),
			},
			execCtx:        &ExecutionContext{},
			expectedStatus: ActionStatusCompleted,
			shouldSleep:    true,
		},
		{
			name: "sleep exceeds timeout",
			action: Action{
				Name: "test_sleep_exceeds_timeout",
				Type: "sleep",
				Config: map[string]interface{}{
					"seconds": 2.0,
				},
				Timeout: func() *int { t := 1; return &t }(),
			},
			execCtx:        &ExecutionContext{},
			expectedStatus: ActionStatusFailed,
			expectedError:  "sleep duration 2.000000 seconds exceeds timeout 1 seconds",
			shouldSleep:    false,
		},
		{
			name: "missing config",
			action: Action{
				Name: "test_missing_config",
				Type: "sleep",
			},
			execCtx:        &ExecutionContext{},
			expectedStatus: ActionStatusFailed,
			expectedError:  "invalid config: config is required",
			shouldSleep:    false,
		},
		{
			name: "missing seconds",
			action: Action{
				Name: "test_missing_seconds",
				Type: "sleep",
				Config: map[string]interface{}{
					"other_field": "value",
				},
			},
			execCtx:        &ExecutionContext{},
			expectedStatus: ActionStatusFailed,
			expectedError:  "invalid config: 'seconds' field is required and must be number",
			shouldSleep:    false,
		},
		{
			name: "invalid seconds type",
			action: Action{
				Name: "test_invalid_seconds",
				Type: "sleep",
				Config: map[string]interface{}{
					"seconds": "not_a_number",
				},
			},
			execCtx:        &ExecutionContext{},
			expectedStatus: ActionStatusFailed,
			expectedError:  "invalid config: 'seconds' field is required and must be number",
			shouldSleep:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startTime := time.Now()
			result := executor.executeSleep(ctx, tt.execCtx, tt.action)
			endTime := time.Now()

			assert.Equal(t, tt.expectedStatus, result.Status)
			if tt.expectedError != "" {
				assert.Contains(t, result.Error, tt.expectedError)
			}

			if result.Status == ActionStatusCompleted {
				assert.NotNil(t, result.Output)
				assert.NotNil(t, result.EndTime)
				assert.Contains(t, result.Output, "seconds")
				assert.Contains(t, result.Output, "slept")
				assert.True(t, result.Output["slept"].(bool))

				// Check that sleep actually happened
				if tt.shouldSleep {
					duration := endTime.Sub(startTime)
					assert.GreaterOrEqual(t, duration, time.Duration(50*time.Millisecond))
				}
			}
		})
	}
}

func TestExecuteSleepLogging(t *testing.T) {
	executor := NewExecutor()
	ctx := context.Background()
	execCtx := &ExecutionContext{}

	action := Action{
		Name: "test_sleep_logging",
		Type: "sleep",
		Config: map[string]interface{}{
			"seconds": 0.1,
		},
	}

	initialLogCount := len(execCtx.Logs)
	result := executor.executeSleep(ctx, execCtx, action)

	assert.Equal(t, ActionStatusCompleted, result.Status)

	// Check that logs were added
	assert.Greater(t, len(execCtx.Logs), initialLogCount)

	// Check for sleep start log
	foundStartLog := false
	foundEndLog := false
	for _, log := range execCtx.Logs {
		if log.Action == "test_sleep_logging" && log.Type == "sleep" {
			if log.Message == "Sleep started" {
				foundStartLog = true
			}
			if log.Message == "Sleep completed" {
				foundEndLog = true
			}
		}
	}

	assert.True(t, foundStartLog, "Sleep start log not found")
	assert.True(t, foundEndLog, "Sleep end log not found")
}
