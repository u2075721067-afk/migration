package rules

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ActionFunc defines the signature for action functions
type ActionFunc func(params map[string]interface{}, ctx Context) (map[string]interface{}, error)

// registerDefaultActions registers all default actions
func (e *Engine) registerDefaultActions() {
	e.actions[ActionSetVar] = actionSetVar
	e.actions[ActionRetry] = actionRetry
	e.actions[ActionHTTPCall] = actionHTTPCall
	e.actions[ActionSkip] = actionSkip
	e.actions[ActionLog] = actionLog
	e.actions[ActionRoute] = actionRoute
	e.actions[ActionStop] = actionStop
	e.actions[ActionTransform] = actionTransform
}

// actionSetVar sets a variable in the context
func actionSetVar(params map[string]interface{}, ctx Context) (map[string]interface{}, error) {
	variable, ok := params["variable"].(string)
	if !ok {
		return nil, fmt.Errorf("set_var action requires 'variable' parameter")
	}

	value, exists := params["value"]
	if !exists {
		return nil, fmt.Errorf("set_var action requires 'value' parameter")
	}

	// Set the variable in context
	ctx.Variables[variable] = value

	return map[string]interface{}{
		"variable": variable,
		"value":    value,
		"set_at":   time.Now(),
	}, nil
}

// actionRetry triggers a retry with specified profile
func actionRetry(params map[string]interface{}, ctx Context) (map[string]interface{}, error) {
	profile, ok := params["profile"].(string)
	if !ok {
		profile = "default"
	}

	maxAttempts, ok := params["max_attempts"].(float64)
	if !ok {
		maxAttempts = 3
	}

	delay, ok := params["delay"].(float64)
	if !ok {
		delay = 1000 // milliseconds
	}

	return map[string]interface{}{
		"action":       "retry",
		"profile":      profile,
		"max_attempts": int(maxAttempts),
		"delay_ms":     int(delay),
		"triggered_at": time.Now(),
	}, nil
}

// actionHTTPCall makes an HTTP request
func actionHTTPCall(params map[string]interface{}, ctx Context) (map[string]interface{}, error) {
	url, ok := params["url"].(string)
	if !ok {
		return nil, fmt.Errorf("http_call action requires 'url' parameter")
	}

	method, ok := params["method"].(string)
	if !ok {
		method = "GET"
	}

	timeout, ok := params["timeout"].(float64)
	if !ok {
		timeout = 30000 // 30 seconds in milliseconds
	}

	var body io.Reader
	if bodyData, exists := params["body"]; exists {
		if bodyStr, ok := bodyData.(string); ok {
			body = bytes.NewBufferString(bodyStr)
		} else {
			// Convert to JSON
			jsonData, err := json.Marshal(bodyData)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request body: %s", err.Error())
			}
			body = bytes.NewBuffer(jsonData)
		}
	}

	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Millisecond,
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %s", err.Error())
	}

	// Add headers
	if headers, ok := params["headers"].(map[string]interface{}); ok {
		for key, value := range headers {
			req.Header.Set(key, fmt.Sprintf("%v", value))
		}
	}

	// Set content type for POST/PUT requests with body
	if body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	startTime := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(startTime)

	if err != nil {
		return map[string]interface{}{
			"success":     false,
			"error":       err.Error(),
			"duration_ms": duration.Milliseconds(),
			"url":         url,
			"method":      method,
		}, nil
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return map[string]interface{}{
			"success":     false,
			"error":       fmt.Sprintf("failed to read response body: %s", err.Error()),
			"status_code": resp.StatusCode,
			"duration_ms": duration.Milliseconds(),
			"url":         url,
			"method":      method,
		}, nil
	}

	return map[string]interface{}{
		"success":     resp.StatusCode >= 200 && resp.StatusCode < 300,
		"status_code": resp.StatusCode,
		"headers":     resp.Header,
		"body":        string(responseBody),
		"duration_ms": duration.Milliseconds(),
		"url":         url,
		"method":      method,
	}, nil
}

// actionSkip marks the current execution for skipping
func actionSkip(params map[string]interface{}, ctx Context) (map[string]interface{}, error) {
	reason, ok := params["reason"].(string)
	if !ok {
		reason = "Rule condition matched - skipping execution"
	}

	return map[string]interface{}{
		"action":     "skip",
		"reason":     reason,
		"skipped_at": time.Now(),
	}, nil
}

// actionLog logs a message
func actionLog(params map[string]interface{}, ctx Context) (map[string]interface{}, error) {
	message, ok := params["message"].(string)
	if !ok {
		return nil, fmt.Errorf("log action requires 'message' parameter")
	}

	level, ok := params["level"].(string)
	if !ok {
		level = "info"
	}

	// In a real implementation, this would use the actual logging system
	fmt.Printf("[%s] %s: %s\n", time.Now().Format(time.RFC3339), level, message)

	return map[string]interface{}{
		"action":    "log",
		"message":   message,
		"level":     level,
		"logged_at": time.Now(),
	}, nil
}

// actionRoute routes to a different workflow
func actionRoute(params map[string]interface{}, ctx Context) (map[string]interface{}, error) {
	workflow, ok := params["workflow"].(string)
	if !ok {
		return nil, fmt.Errorf("route action requires 'workflow' parameter")
	}

	reason, ok := params["reason"].(string)
	if !ok {
		reason = "Rule condition matched - routing to different workflow"
	}

	return map[string]interface{}{
		"action":    "route",
		"workflow":  workflow,
		"reason":    reason,
		"routed_at": time.Now(),
	}, nil
}

// actionStop stops the current execution
func actionStop(params map[string]interface{}, ctx Context) (map[string]interface{}, error) {
	reason, ok := params["reason"].(string)
	if !ok {
		reason = "Rule condition matched - stopping execution"
	}

	return map[string]interface{}{
		"action":     "stop",
		"reason":     reason,
		"stopped_at": time.Now(),
	}, nil
}

// actionTransform transforms data in the context
func actionTransform(params map[string]interface{}, ctx Context) (map[string]interface{}, error) {
	transformType, ok := params["type"].(string)
	if !ok {
		return nil, fmt.Errorf("transform action requires 'type' parameter")
	}

	source, ok := params["source"].(string)
	if !ok {
		return nil, fmt.Errorf("transform action requires 'source' parameter")
	}

	target, ok := params["target"].(string)
	if !ok {
		return nil, fmt.Errorf("transform action requires 'target' parameter")
	}

	// Get source value
	sourceValue := getValueFromContext(source, ctx)
	if sourceValue == nil {
		return nil, fmt.Errorf("source field '%s' not found in context", source)
	}

	var transformedValue interface{}
	var err error

	switch transformType {
	case "uppercase":
		if str, ok := sourceValue.(string); ok {
			transformedValue = strings.ToUpper(str)
		} else {
			transformedValue = strings.ToUpper(fmt.Sprintf("%v", sourceValue))
		}
	case "lowercase":
		if str, ok := sourceValue.(string); ok {
			transformedValue = strings.ToLower(str)
		} else {
			transformedValue = strings.ToLower(fmt.Sprintf("%v", sourceValue))
		}
	case "json_parse":
		if str, ok := sourceValue.(string); ok {
			var jsonData interface{}
			err = json.Unmarshal([]byte(str), &jsonData)
			if err != nil {
				return nil, fmt.Errorf("failed to parse JSON: %s", err.Error())
			}
			transformedValue = jsonData
		} else {
			return nil, fmt.Errorf("json_parse transform requires string input")
		}
	case "json_stringify":
		jsonData, err := json.Marshal(sourceValue)
		if err != nil {
			return nil, fmt.Errorf("failed to stringify to JSON: %s", err.Error())
		}
		transformedValue = string(jsonData)
	default:
		return nil, fmt.Errorf("unknown transform type: %s", transformType)
	}

	// Set target value in context
	ctx.Variables[target] = transformedValue

	return map[string]interface{}{
		"action":            "transform",
		"type":              transformType,
		"source":            source,
		"target":            target,
		"source_value":      sourceValue,
		"transformed_value": transformedValue,
		"transformed_at":    time.Now(),
	}, nil
}

// getValueFromContext retrieves a value from context using dot notation
func getValueFromContext(field string, ctx Context) interface{} {
	// Simple implementation - can be extended for nested field access
	if val, ok := ctx.Variables[field]; ok {
		return val
	}

	if val, ok := ctx.Request[field]; ok {
		return val
	}

	if val, ok := ctx.Response[field]; ok {
		return val
	}

	if val, ok := ctx.Metadata[field]; ok {
		return val
	}

	return nil
}
