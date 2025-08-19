package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/PaesslerAG/jsonpath"
	"github.com/google/uuid"
	"github.com/mova-engine/mova-engine/config"
	"github.com/sirupsen/logrus"
)

// ExecutionContext represents the runtime context for workflow execution
type ExecutionContext struct {
	RunID      string                  `json:"run_id"`
	WorkflowID string                  `json:"workflow_id"`
	StartTime  time.Time               `json:"start_time"`
	EndTime    *time.Time              `json:"end_time,omitempty"`
	Status     ExecutionStatus         `json:"status"`
	Variables  map[string]interface{}  `json:"variables"`
	Results    map[string]ActionResult `json:"results"`
	Logs       []ExecutionLog          `json:"logs"`
}

// ExecutionStatus represents the current status of workflow execution
type ExecutionStatus string

const (
	StatusPending   ExecutionStatus = "pending"
	StatusRunning   ExecutionStatus = "running"
	StatusCompleted ExecutionStatus = "completed"
	StatusFailed    ExecutionStatus = "failed"
	StatusCancelled ExecutionStatus = "cancelled"
)

// ActionResult represents the result of a single action execution
type ActionResult struct {
	ActionName string                 `json:"action_name"`
	Status     ActionStatus           `json:"status"`
	StartTime  time.Time              `json:"start_time"`
	EndTime    *time.Time             `json:"end_time,omitempty"`
	Output     map[string]interface{} `json:"output,omitempty"`
	Error      string                 `json:"error,omitempty"`
	Attempts   int                    `json:"attempts"`
}

// ActionStatus represents the status of an individual action
type ActionStatus string

const (
	ActionStatusPending   ActionStatus = "pending"
	ActionStatusRunning   ActionStatus = "running"
	ActionStatusCompleted ActionStatus = "completed"
	ActionStatusFailed    ActionStatus = "failed"
	ActionStatusSkipped   ActionStatus = "skipped"
)

// ExecutionLog represents a log entry during execution
type ExecutionLog struct {
	Timestamp      time.Time              `json:"timestamp"`
	Level          string                 `json:"level"`
	Step           string                 `json:"step"`
	Type           string                 `json:"type"`
	Action         string                 `json:"action,omitempty"`
	Message        string                 `json:"message"`
	ParamsRedacted map[string]interface{} `json:"params_redacted,omitempty"`
	Status         string                 `json:"status"`
	Data           map[string]interface{} `json:"data,omitempty"`
}

// MOVAEnvelope represents the MOVA v3.1 workflow envelope
type MOVAEnvelope struct {
	Intent    Intent                 `json:"intent"`
	Payload   map[string]interface{} `json:"payload"`
	Actions   []Action               `json:"actions"`
	Variables map[string]interface{} `json:"variables,omitempty"`
	Secrets   map[string]string      `json:"secrets,omitempty"`
}

// Intent represents workflow metadata and configuration
type Intent struct {
	Name        string             `json:"name"`
	Version     string             `json:"version"`
	Description string             `json:"description"`
	Author      string             `json:"author,omitempty"`
	Tags        []string           `json:"tags,omitempty"`
	Timeout     *int               `json:"timeout,omitempty"`
	Retry       *RetryPolicy       `json:"retry,omitempty"`
	Budget      *BudgetConstraints `json:"budget,omitempty"`
}

// Action represents a single executable action
type Action struct {
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Enabled     *bool                  `json:"enabled,omitempty"`
	Timeout     *int                   `json:"timeout,omitempty"`
	Retry       *RetryPolicy           `json:"retry,omitempty"`
	OnSuccess   []string               `json:"on_success,omitempty"`
	OnFailure   []string               `json:"on_failure,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
}

// RetryPolicy defines retry behavior for actions
type RetryPolicy struct {
	MaxAttempts int    `json:"max_attempts"`
	Backoff     string `json:"backoff"`
	Delay       int    `json:"delay"`
}

// BudgetConstraints define resource limits
type BudgetConstraints struct {
	MaxMemoryMB   int `json:"max_memory_mb"`
	MaxCPUPercent int `json:"max_cpu_percent"`
}

// Executor is the main workflow execution engine
type Executor struct {
	logger         *logrus.Logger
	securityConfig *config.SecurityConfig
}

// NewExecutor creates a new MOVA executor instance
func NewExecutor() *Executor {
	return &Executor{
		logger:         logrus.New(),
		securityConfig: config.DefaultSecurityConfig(),
	}
}

// NewExecutorWithConfig creates a new MOVA executor instance with custom security config
func NewExecutorWithConfig(securityConfig *config.SecurityConfig) *Executor {
	return &Executor{
		logger:         logrus.New(),
		securityConfig: securityConfig,
	}
}

// Execute runs a MOVA workflow envelope with security controls
func (e *Executor) Execute(ctx context.Context, envelope MOVAEnvelope) (*ExecutionContext, error) {
	// Apply workflow timeout from security config
	workflowTimeout := e.securityConfig.Timeouts.WorkflowTimeout
	workflowCtx, cancel := context.WithTimeout(ctx, workflowTimeout)
	defer cancel()

	execCtx := &ExecutionContext{
		RunID:      uuid.New().String(),
		WorkflowID: envelope.Intent.Name,
		StartTime:  time.Now(),
		Status:     StatusRunning,
		Variables:  make(map[string]interface{}),
		Results:    make(map[string]ActionResult),
		Logs:       make([]ExecutionLog, 0),
	}

	// Initialize variables
	if envelope.Variables != nil {
		for k, v := range envelope.Variables {
			execCtx.Variables[k] = v
		}
	}

	// Add initial payload to variables
	execCtx.Variables["payload"] = envelope.Payload

	e.logExecution(execCtx, "workflow", "workflow", "workflow", "Workflow execution started", map[string]interface{}{
		"workflow_name": envelope.Intent.Name,
		"version":       envelope.Intent.Version,
		"action_count":  len(envelope.Actions),
	}, nil)

	// Execute actions sequentially for MVP
	for _, action := range envelope.Actions {
		// Check if workflow timeout has been exceeded
		select {
		case <-workflowCtx.Done():
			execCtx.Status = StatusFailed
			now := time.Now()
			execCtx.EndTime = &now
			e.logExecution(execCtx, "workflow", "timeout", "workflow", "Workflow execution timed out", map[string]interface{}{
				"timeout": workflowTimeout.String(),
			}, nil)
			return execCtx, fmt.Errorf("workflow timeout after %v", workflowTimeout)
		default:
			// Continue with action execution
		}

		if action.Enabled != nil && !*action.Enabled {
			e.logExecution(execCtx, "action", "skip", action.Name, "Action skipped (disabled)", nil, nil)
			continue
		}

		result := e.executeAction(workflowCtx, execCtx, action)
		execCtx.Results[action.Name] = result

		if result.Status == ActionStatusFailed {
			execCtx.Status = StatusFailed
			now := time.Now()
			execCtx.EndTime = &now
			e.logExecution(execCtx, "workflow", "error", "workflow", "Workflow execution failed", map[string]interface{}{
				"failed_action": action.Name,
				"error":         result.Error,
			}, nil)
			return execCtx, fmt.Errorf("action '%s' failed: %s", action.Name, result.Error)
		}
	}

	execCtx.Status = StatusCompleted
	now := time.Now()
	execCtx.EndTime = &now
	e.logExecution(execCtx, "workflow", "success", "workflow", "Workflow execution completed successfully", nil, nil)

	return execCtx, nil
}

// executeAction executes a single action
func (e *Executor) executeAction(ctx context.Context, execCtx *ExecutionContext, action Action) ActionResult {
	startTime := time.Now()
	result := ActionResult{
		ActionName: action.Name,
		Status:     ActionStatusRunning,
		StartTime:  startTime,
		Attempts:   1,
	}

	// Apply action timeout from security config
	actionTimeout := e.securityConfig.Timeouts.ActionTimeout
	actionCtx, cancel := context.WithTimeout(ctx, actionTimeout)
	defer cancel()

	e.logExecution(execCtx, "action", "start", action.Name, "Action execution started", map[string]interface{}{
		"action_type": action.Type,
		"config":      action.Config,
		"timeout":     actionTimeout.String(),
	}, nil)

	// Execute action with timeout control
	switch action.Type {
	case "http_fetch":
		result = e.executeHTTPFetch(actionCtx, execCtx, action)
	case "set":
		result = e.executeSet(actionCtx, execCtx, action)
	case "if":
		result = e.executeIf(actionCtx, execCtx, action)
	case "repeat":
		result = e.executeRepeat(actionCtx, execCtx, action)
	case "print":
		result = e.executePrint(actionCtx, execCtx, action)
	case "call":
		result = e.executeCall(actionCtx, execCtx, action)
	case "parse_json":
		result = e.executeParseJSON(actionCtx, execCtx, action)
	case "sleep":
		result = e.executeSleep(actionCtx, execCtx, action)
	default:
		result.Status = ActionStatusFailed
		result.Error = fmt.Sprintf("unsupported action type: %s", action.Type)
	}

	// Check if action was cancelled due to timeout
	if actionCtx.Err() == context.DeadlineExceeded {
		result.Status = ActionStatusFailed
		result.Error = fmt.Sprintf("action timeout after %v", actionTimeout)

		e.logExecution(execCtx, "action", "timeout", action.Name, "Action execution timed out", map[string]interface{}{
			"action_type": action.Type,
			"timeout":     actionTimeout.String(),
		}, nil)
	}

	endTime := time.Now()
	result.EndTime = &endTime

	if result.Status == ActionStatusCompleted {
		e.logExecution(execCtx, "action", "success", action.Name, "Action completed successfully", nil, result.Output)
	} else {
		e.logExecution(execCtx, "action", "error", action.Name, "Action failed", map[string]interface{}{
			"error": result.Error,
		}, nil)
	}

	return result
}

// executeHTTPFetch executes an HTTP request action with security controls
func (e *Executor) executeHTTPFetch(ctx context.Context, execCtx *ExecutionContext, action Action) ActionResult {
	startTime := time.Now()
	result := ActionResult{
		ActionName: action.Name,
		Status:     ActionStatusRunning,
		StartTime:  startTime,
		Attempts:   1,
	}

	// Extract configuration
	config := action.Config
	if config == nil {
		result.Status = ActionStatusFailed
		result.Error = "invalid config: config is required"
		endTime := time.Now()
		result.EndTime = &endTime
		return result
	}

	// Get URL and method
	url, _ := config["url"].(string)
	if url == "" {
		result.Status = ActionStatusFailed
		result.Error = "invalid config: 'url' field is required and must be string"
		endTime := time.Now()
		result.EndTime = &endTime
		return result
	}

	method, _ := config["method"].(string)
	if method == "" {
		method = "GET" // default method
	}

	// Security validation: Check if URL is allowed
	if err := e.securityConfig.HTTP.ValidateURL(url); err != nil {
		result.Status = ActionStatusFailed
		result.Error = fmt.Sprintf("security validation failed: %v", err)
		endTime := time.Now()
		result.EndTime = &endTime

		e.logExecution(execCtx, "action", "http_fetch", action.Name, "HTTP request blocked by security policy", map[string]interface{}{
			"url":    url,
			"method": method,
			"error":  result.Error,
		}, nil)
		return result
	}

	// Get headers
	var headers map[string]interface{}
	if h, ok := config["headers"].(map[string]interface{}); ok {
		headers = h
	}

	// Get body
	var body interface{}
	if b, exists := config["body"]; exists {
		body = b
	}

	// Get timeout
	timeoutMs, _ := config["timeout_ms"].(float64)
	if timeoutMs == 0 {
		timeoutMs = float64(e.securityConfig.Timeouts.HTTPTimeout.Milliseconds())
	}

	// Apply security timeout limits
	maxTimeoutMs := float64(e.securityConfig.Timeouts.HTTPTimeout.Milliseconds())
	if timeoutMs > maxTimeoutMs {
		timeoutMs = maxTimeoutMs
	}

	// Log the HTTP request (with redacted sensitive data)
	e.logExecution(execCtx, "action", "http_fetch", action.Name, "HTTP request started", map[string]interface{}{
		"url":     url,
		"method":  method,
		"timeout": timeoutMs,
		"headers": headers, // Will be redacted by logExecution
	}, nil)

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: time.Duration(timeoutMs) * time.Millisecond,
	}

	// Don't follow redirects if disabled in security config
	if !e.securityConfig.HTTP.FollowRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	// Prepare request body
	var bodyReader io.Reader
	var contentType string
	if body != nil {
		switch v := body.(type) {
		case string:
			bodyReader = bytes.NewBufferString(v)
			contentType = "text/plain"
		case map[string]interface{}:
			jsonData, err := json.Marshal(v)
			if err != nil {
				result.Status = ActionStatusFailed
				result.Error = fmt.Sprintf("failed to marshal JSON body: %v", err)
				endTime := time.Now()
				result.EndTime = &endTime
				return result
			}
			bodyReader = bytes.NewBuffer(jsonData)
			contentType = "application/json"
		default:
			// Try to marshal as JSON
			jsonData, err := json.Marshal(v)
			if err != nil {
				result.Status = ActionStatusFailed
				result.Error = fmt.Sprintf("unsupported body type: %T", v)
				endTime := time.Now()
				result.EndTime = &endTime
				return result
			}
			bodyReader = bytes.NewBuffer(jsonData)
			contentType = "application/json"
		}
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		result.Status = ActionStatusFailed
		result.Error = fmt.Sprintf("failed to create HTTP request: %v", err)
		endTime := time.Now()
		result.EndTime = &endTime
		return result
	}

	// Set User-Agent
	req.Header.Set("User-Agent", e.securityConfig.HTTP.UserAgent)

	// Set Content-Type if we have a body
	if bodyReader != nil && contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	// Add custom headers
	if headers != nil {
		for key, value := range headers {
			if strValue, ok := value.(string); ok {
				req.Header.Set(key, strValue)
			}
		}
	}

	// Execute HTTP request
	resp, err := client.Do(req)
	if err != nil {
		result.Status = ActionStatusFailed
		result.Error = fmt.Sprintf("HTTP request failed: %v", err)
		endTime := time.Now()
		result.EndTime = &endTime

		e.logExecution(execCtx, "action", "http_fetch", action.Name, "HTTP request failed", map[string]interface{}{
			"url":    url,
			"method": method,
			"error":  result.Error,
		}, nil)
		return result
	}
	defer resp.Body.Close()

	// Check response size limit
	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, e.securityConfig.HTTP.MaxResponseSize))
	if err != nil {
		result.Status = ActionStatusFailed
		result.Error = fmt.Sprintf("failed to read response body: %v", err)
		endTime := time.Now()
		result.EndTime = &endTime
		return result
	}

	// Parse response body as JSON if possible
	var responseData interface{}
	if len(responseBody) > 0 {
		// Try to parse as JSON first
		if err := json.Unmarshal(responseBody, &responseData); err != nil {
			// If JSON parsing fails, use as string
			responseData = string(responseBody)
		}
	}

	// Prepare response headers (without sensitive data)
	responseHeaders := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			responseHeaders[key] = values[0]
		}
	}

	// Create response data
	output := map[string]interface{}{
		"status_code": resp.StatusCode,
		"status":      resp.Status,
		"headers":     RedactSecrets(responseHeaders),
		"body":        responseData,
		"url":         url,
		"method":      method,
	}

	result.Status = ActionStatusCompleted
	result.Output = output
	endTime := time.Now()
	result.EndTime = &endTime

	e.logExecution(execCtx, "action", "http_fetch", action.Name, "HTTP request completed", map[string]interface{}{
		"url":         url,
		"method":      method,
		"status_code": resp.StatusCode,
		"body_size":   len(responseBody),
	}, output)

	return result
}

func (e *Executor) executeSet(ctx context.Context, execCtx *ExecutionContext, action Action) ActionResult {
	startTime := time.Now()
	result := ActionResult{
		ActionName: action.Name,
		Status:     ActionStatusRunning,
		StartTime:  startTime,
		Attempts:   1,
	}

	// Extract configuration
	config := action.Config
	if config == nil {
		result.Status = ActionStatusFailed
		result.Error = "invalid config: config is required"
		endTime := time.Now()
		result.EndTime = &endTime
		return result
	}

	// Get variable name and value
	varName, ok := config["variable"].(string)
	if !ok {
		result.Status = ActionStatusFailed
		result.Error = "invalid config: 'variable' field is required and must be string"
		endTime := time.Now()
		result.EndTime = &endTime
		return result
	}

	value, exists := config["value"]
	if !exists {
		result.Status = ActionStatusFailed
		result.Error = "invalid config: 'value' field is required"
		endTime := time.Now()
		result.EndTime = &endTime
		return result
	}

	// Set variable in context
	execCtx.Variables[varName] = value

	// Log the action
	e.logExecution(execCtx, "action", "set", action.Name, "Variable set successfully", map[string]interface{}{
		"variable": varName,
		"value":    value,
	}, map[string]interface{}{
		"variable": varName,
		"value":    value,
	})

	result.Status = ActionStatusCompleted
	result.Output = map[string]interface{}{
		"variable": varName,
		"value":    value,
	}
	endTime := time.Now()
	result.EndTime = &endTime

	return result
}

func (e *Executor) executeIf(ctx context.Context, execCtx *ExecutionContext, action Action) ActionResult {
	startTime := time.Now()
	result := ActionResult{
		ActionName: action.Name,
		Status:     ActionStatusRunning,
		StartTime:  startTime,
		Attempts:   1,
	}

	// Extract configuration
	config := action.Config
	if config == nil {
		result.Status = ActionStatusFailed
		result.Error = "invalid config: config is required"
		endTime := time.Now()
		result.EndTime = &endTime
		return result
	}

	// Get condition
	condition, _ := config["condition"]
	if condition == nil {
		result.Status = ActionStatusFailed
		result.Error = "invalid config: 'condition' field is required"
		endTime := time.Now()
		result.EndTime = &endTime
		return result
	}

	// Evaluate condition
	var conditionResult bool
	switch v := condition.(type) {
	case bool:
		conditionResult = v
	case string:
		// Simple string evaluation - non-empty is true
		conditionResult = v != ""
	case float64:
		// Number evaluation - non-zero is true
		conditionResult = v != 0
	default:
		conditionResult = condition != nil
	}

	// Get then/else actions
	thenActions, _ := config["then"].([]string)
	elseActions, _ := config["else"].([]string)

	// Log the condition evaluation
	e.logExecution(execCtx, "action", "if", action.Name, "Condition evaluated", map[string]interface{}{
		"condition": condition,
		"result":    conditionResult,
		"then":      thenActions,
		"else":      elseActions,
	}, map[string]interface{}{
		"condition": condition,
		"result":    conditionResult,
		"then":      thenActions,
		"else":      elseActions,
	})

	result.Status = ActionStatusCompleted
	result.Output = map[string]interface{}{
		"condition": condition,
		"result":    conditionResult,
		"then":      thenActions,
		"else":      elseActions,
	}
	endTime := time.Now()
	result.EndTime = &endTime

	return result
}

func (e *Executor) executeRepeat(ctx context.Context, execCtx *ExecutionContext, action Action) ActionResult {
	startTime := time.Now()
	result := ActionResult{
		ActionName: action.Name,
		Status:     ActionStatusRunning,
		StartTime:  startTime,
		Attempts:   1,
	}

	// Extract configuration
	config := action.Config
	if config == nil {
		result.Status = ActionStatusFailed
		result.Error = "invalid config: config is required"
		endTime := time.Now()
		result.EndTime = &endTime
		return result
	}

	// Get iteration data
	iterations, _ := config["iterations"]
	if iterations == nil {
		result.Status = ActionStatusFailed
		result.Error = "invalid config: 'iterations' field is required"
		endTime := time.Now()
		result.EndTime = &endTime
		return result
	}

	// Handle different iteration types
	var iterationCount int
	var iterationData []interface{}

	switch v := iterations.(type) {
	case float64:
		iterationCount = int(v)
		iterationData = make([]interface{}, iterationCount)
		for i := 0; i < iterationCount; i++ {
			iterationData[i] = i
		}
	case []interface{}:
		iterationData = v
		iterationCount = len(v)
	case string:
		// Try to parse as number
		if count, err := strconv.Atoi(v); err == nil {
			iterationCount = count
			iterationData = make([]interface{}, iterationCount)
			for i := 0; i < iterationCount; i++ {
				iterationData[i] = i
			}
		} else {
			// Treat as single item array
			iterationData = []interface{}{v}
			iterationCount = 1
		}
	default:
		// Single item
		iterationData = []interface{}{iterations}
		iterationCount = 1
	}

	// Get actions to repeat
	repeatActions, _ := config["actions"].([]string)

	// Log the repeat action
	e.logExecution(execCtx, "action", "repeat", action.Name, "Repeat action configured", map[string]interface{}{
		"iterations":      iterations,
		"iteration_count": iterationCount,
		"actions":         repeatActions,
	}, map[string]interface{}{
		"iterations":      iterations,
		"iteration_count": iterationCount,
		"actions":         repeatActions,
	})

	result.Status = ActionStatusCompleted
	result.Output = map[string]interface{}{
		"iterations":      iterations,
		"iteration_count": iterationCount,
		"actions":         repeatActions,
	}
	endTime := time.Now()
	result.EndTime = &endTime

	return result
}

func (e *Executor) executePrint(ctx context.Context, execCtx *ExecutionContext, action Action) ActionResult {
	startTime := time.Now()
	result := ActionResult{
		ActionName: action.Name,
		Status:     ActionStatusRunning,
		StartTime:  startTime,
		Attempts:   1,
	}

	// Extract configuration
	config := action.Config
	if config == nil {
		result.Status = ActionStatusFailed
		result.Error = "invalid config: config is required"
		endTime := time.Now()
		result.EndTime = &endTime
		return result
	}

	// Get message and data
	message, _ := config["message"].(string)
	if message == "" {
		message = "Print action executed"
	}

	data, _ := config["data"]

	// Log the print action
	e.logExecution(execCtx, "action", "print", action.Name, message, map[string]interface{}{
		"message": message,
		"data":    data,
	}, map[string]interface{}{
		"message": message,
		"data":    data,
	})

	result.Status = ActionStatusCompleted
	result.Output = map[string]interface{}{
		"message": message,
		"data":    data,
	}
	endTime := time.Now()
	result.EndTime = &endTime

	return result
}

// executeCall executes a function call action
func (e *Executor) executeCall(ctx context.Context, execCtx *ExecutionContext, action Action) ActionResult {
	startTime := time.Now()
	result := ActionResult{
		ActionName: action.Name,
		Status:     ActionStatusRunning,
		StartTime:  startTime,
		Attempts:   1,
	}

	// Extract configuration
	config := action.Config
	if config == nil {
		result.Status = ActionStatusFailed
		result.Error = "invalid config: config is required"
		endTime := time.Now()
		result.EndTime = &endTime
		return result
	}

	// Get function name and parameters
	funcName, _ := config["function"].(string)
	if funcName == "" {
		result.Status = ActionStatusFailed
		result.Error = "invalid config: 'function' field is required and must be string"
		endTime := time.Now()
		result.EndTime = &endTime
		return result
	}

	params, _ := config["params"].(map[string]interface{})
	if params == nil {
		params = make(map[string]interface{})
	}

	// Log the function call
	e.logExecution(execCtx, "action", "call", action.Name, "Function call configured", map[string]interface{}{
		"function": funcName,
		"params":   params,
	}, map[string]interface{}{
		"function": funcName,
		"params":   params,
	})

	result.Status = ActionStatusCompleted
	result.Output = map[string]interface{}{
		"function": funcName,
		"params":   params,
		"status":   "function call prepared",
	}
	endTime := time.Now()
	result.EndTime = &endTime

	return result
}

// executeParseJSON executes a JSON parsing action using JSONPath
func (e *Executor) executeParseJSON(ctx context.Context, execCtx *ExecutionContext, action Action) ActionResult {
	startTime := time.Now()
	result := ActionResult{
		ActionName: action.Name,
		Status:     ActionStatusRunning,
		StartTime:  startTime,
		Attempts:   1,
	}

	// Extract configuration
	config := action.Config
	if config == nil {
		result.Status = ActionStatusFailed
		result.Error = "invalid config: config is required"
		endTime := time.Now()
		result.EndTime = &endTime
		return result
	}

	// Get JSONPath expression and source
	jsonPath, _ := config["jsonpath"].(string)
	if jsonPath == "" {
		result.Status = ActionStatusFailed
		result.Error = "invalid config: 'jsonpath' field is required and must be string"
		endTime := time.Now()
		result.EndTime = &endTime
		return result
	}

	source, _ := config["source"].(string)
	if source == "" {
		source = "last_result" // default to last result
	}

	// Get source data
	var sourceData interface{}
	switch source {
	case "last_result":
		// Get last action result
		if len(execCtx.Results) > 0 {
			// Find last completed action
			for _, res := range execCtx.Results {
				if res.Status == ActionStatusCompleted {
					sourceData = res.Output
					break
				}
			}
		}
	case "payload":
		sourceData = execCtx.Variables["payload"]
	default:
		// Try to get from variables
		sourceData = execCtx.Variables[source]
	}

	if sourceData == nil {
		result.Status = ActionStatusFailed
		result.Error = fmt.Sprintf("source data not found: %s", source)
		endTime := time.Now()
		result.EndTime = &endTime
		return result
	}

	// Execute JSONPath
	parsedData, err := jsonpath.Get(jsonPath, sourceData)
	if err != nil {
		result.Status = ActionStatusFailed
		result.Error = fmt.Sprintf("JSONPath evaluation failed: %s", err.Error())
		endTime := time.Now()
		result.EndTime = &endTime
		return result
	}

	// Store result in variable if specified
	varName, _ := config["variable"].(string)
	if varName != "" {
		execCtx.Variables[varName] = parsedData
	}

	// Log the parse action
	e.logExecution(execCtx, "action", "parse_json", action.Name, "JSON parsed successfully", map[string]interface{}{
		"jsonpath": jsonPath,
		"source":   source,
		"variable": varName,
	}, map[string]interface{}{
		"jsonpath": jsonPath,
		"source":   source,
		"variable": varName,
		"result":   parsedData,
	})

	result.Status = ActionStatusCompleted
	result.Output = map[string]interface{}{
		"jsonpath": jsonPath,
		"source":   source,
		"variable": varName,
		"result":   parsedData,
	}
	endTime := time.Now()
	result.EndTime = &endTime

	return result
}

// executeSleep executes a sleep action
func (e *Executor) executeSleep(ctx context.Context, execCtx *ExecutionContext, action Action) ActionResult {
	startTime := time.Now()
	result := ActionResult{
		ActionName: action.Name,
		Status:     ActionStatusRunning,
		StartTime:  startTime,
		Attempts:   1,
	}

	// Extract configuration
	config := action.Config
	if config == nil {
		result.Status = ActionStatusFailed
		result.Error = "invalid config: config is required"
		endTime := time.Now()
		result.EndTime = &endTime
		return result
	}

	// Get sleep duration
	seconds, ok := config["seconds"].(float64)
	if !ok {
		result.Status = ActionStatusFailed
		result.Error = "invalid config: 'seconds' field is required and must be number"
		endTime := time.Now()
		result.EndTime = &endTime
		return result
	}

	// Check timeout from meta
	timeout := action.Timeout
	if timeout != nil && *timeout > 0 {
		if int(seconds) > *timeout {
			result.Status = ActionStatusFailed
			result.Error = fmt.Sprintf("sleep duration %f seconds exceeds timeout %d seconds", seconds, *timeout)
			endTime := time.Now()
			result.EndTime = &endTime
			return result
		}
	}

	// Log the sleep action
	e.logExecution(execCtx, "action", "sleep", action.Name, "Sleep started", map[string]interface{}{
		"seconds": seconds,
		"timeout": timeout,
	}, nil)

	// Sleep
	time.Sleep(time.Duration(seconds * float64(time.Second)))

	// Log completion
	e.logExecution(execCtx, "action", "sleep", action.Name, "Sleep completed", map[string]interface{}{
		"seconds": seconds,
	}, nil)

	result.Status = ActionStatusCompleted
	result.Output = map[string]interface{}{
		"seconds": seconds,
		"slept":   true,
	}
	endTime := time.Now()
	result.EndTime = &endTime

	return result
}

// logExecution adds a log entry to the execution context with secret redaction
func (e *Executor) logExecution(execCtx *ExecutionContext, step, actionType, action, message string, params map[string]interface{}, data map[string]interface{}) {
	// Redact sensitive information from message
	redactedMessage := RedactSecretsInString(message)

	// Redact sensitive information from params and data
	var redactedParams map[string]interface{}
	var redactedData map[string]interface{}

	if params != nil {
		redactedParams = RedactSecretsInterface(params)
	}

	if data != nil {
		redactedData = RedactSecretsInterface(data)
	}

	log := ExecutionLog{
		Timestamp:      time.Now(),
		Level:          "info",
		Step:           step,
		Type:           actionType,
		Action:         action,
		Message:        redactedMessage,
		ParamsRedacted: redactedParams,
		Status:         "success",
		Data:           redactedData,
	}
	execCtx.Logs = append(execCtx.Logs, log)
}

// GetExecutionStatus retrieves the current execution status
func (e *Executor) GetExecutionStatus(runID string) (*ExecutionContext, error) {
	// TODO: Implement status retrieval from persistent storage
	return nil, fmt.Errorf("not implemented")
}

// CancelExecution cancels a running workflow execution
func (e *Executor) CancelExecution(runID string) error {
	// TODO: Implement execution cancellation
	return fmt.Errorf("not implemented")
}

// GetExecutionLogs retrieves execution logs for a specific run
func (e *Executor) GetExecutionLogs(runID string) ([]string, error) {
	// TODO: Implement log retrieval from persistent storage
	// For now, return a placeholder implementation
	return []string{
		fmt.Sprintf(`{"timestamp":"%s","run_id":"%s","message":"Logs not implemented yet"}`,
			time.Now().Format(time.RFC3339), runID),
	}, nil
}
