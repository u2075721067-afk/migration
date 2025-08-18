package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mova-engine/mova-engine/core/executor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestServer() *gin.Engine {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a new router for testing
	router := gin.New()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(loggingMiddleware())
	router.Use(errorHandlingMiddleware())

	// Create mock executor and validator
	exec := &MockExecutor{}
	v := &MockValidator{}

	// API routes
	v1 := router.Group("/v1")
	{
		v1.POST("/execute", handleExecute(exec))
		v1.GET("/runs/:id", handleGetRun(exec))
		v1.GET("/runs/:id/logs", handleGetRunLogs(exec))
		v1.DELETE("/runs/:id", handleCancelRun(exec))
		v1.POST("/validate", handleValidate(v))
		v1.GET("/schemas", handleGetSchemas)
		v1.GET("/schemas/:name", handleGetSchema)
		v1.GET("/introspect", handleIntrospect)
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": "2024-01-01T00:00:00Z",
			"version":   "1.0.0",
		})
	})

	return router
}

// MockExecutor implements a mock executor for testing
type MockExecutor struct{}

func (m *MockExecutor) Execute(ctx context.Context, envelope executor.MOVAEnvelope) (*executor.ExecutionContext, error) {
	// Mock successful execution
	return &executor.ExecutionContext{
		RunID:      "test_run_123",
		WorkflowID: envelope.Intent.Name,
		Status:     executor.StatusCompleted,
		Variables: map[string]interface{}{
			"result": "success",
		},
	}, nil
}

func (m *MockExecutor) GetExecutionStatus(runID string) (*executor.ExecutionContext, error) {
	if runID == "test_run_123" {
		return &executor.ExecutionContext{
			RunID:      runID,
			WorkflowID: "test_workflow",
			Status:     executor.StatusCompleted,
		}, nil
	}
	return nil, fmt.Errorf("run not found")
}

func (m *MockExecutor) GetExecutionLogs(runID string) ([]string, error) {
	if runID == "test_run_123" {
		return []string{
			`{"timestamp":"2024-01-01T00:00:00Z","run_id":"test_run_123","message":"Test log entry"}`,
		}, nil
	}
	return nil, fmt.Errorf("run not found")
}

func (m *MockExecutor) CancelExecution(runID string) error {
	if runID == "test_run_123" {
		return nil
	}
	return fmt.Errorf("run not found")
}

// MockValidator implements a mock validator for testing
type MockValidator struct{}

func (m *MockValidator) ValidateEnvelope(file string) (bool, []error) {
	// Mock validation - always return valid for testing
	return true, nil
}

func TestHealthEndpoint(t *testing.T) {
	router := setupTestServer()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "1.0.0", response["version"])
}

func TestIntrospectEndpoint(t *testing.T) {
	router := setupTestServer()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/introspect", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "MOVA Automation Engine API", response["name"])
	assert.Equal(t, "1.0.0", response["version"])
	assert.Equal(t, "3.1", response["mova_version"])

	endpoints, ok := response["endpoints"].([]interface{})
	assert.True(t, ok)
	assert.Greater(t, len(endpoints), 0)
}

func TestSchemasEndpoint(t *testing.T) {
	router := setupTestServer()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/schemas", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	schemas, ok := response["schemas"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(schemas))
}

func TestValidateEndpoint(t *testing.T) {
	router := setupTestServer()

	// Test valid envelope
	validEnvelope := map[string]interface{}{
		"mova_version": "3.1",
		"intent": map[string]interface{}{
			"name": "test",
		},
		"payload": map[string]interface{}{},
		"actions": []interface{}{},
	}

	body, _ := json.Marshal(validEnvelope)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/validate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, true, response["valid"])
}

func TestExecuteEndpointSynchronous(t *testing.T) {
	router := setupTestServer()

	envelope := map[string]interface{}{
		"mova_version": "3.1",
		"intent": map[string]interface{}{
			"name": "test_workflow",
		},
		"payload": map[string]interface{}{},
		"actions": []interface{}{},
	}

	body, _ := json.Marshal(envelope)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/execute?wait=true", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "test_run_123", response["run_id"])
	assert.Equal(t, "completed", response["status"])
}

func TestExecuteEndpointAsynchronous(t *testing.T) {
	router := setupTestServer()

	envelope := map[string]interface{}{
		"mova_version": "3.1",
		"intent": map[string]interface{}{
			"name": "test_workflow",
		},
		"payload": map[string]interface{}{},
		"actions": []interface{}{},
	}

	body, _ := json.Marshal(envelope)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/execute", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "accepted", response["status"])
	assert.Contains(t, response, "run_id")
}

func TestGetRunEndpoint(t *testing.T) {
	router := setupTestServer()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/runs/test_run_123", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "test_run_123", response["run_id"])
	assert.Equal(t, "completed", response["status"])
}

func TestGetRunLogsEndpoint(t *testing.T) {
	router := setupTestServer()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/runs/test_run_123/logs", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/jsonl", w.Header().Get("Content-Type"))

	body := w.Body.String()
	assert.Contains(t, body, "test_run_123")
	assert.Contains(t, body, "Test log entry")
}

func TestCancelRunEndpoint(t *testing.T) {
	router := setupTestServer()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/v1/runs/test_run_123", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Execution cancelled successfully", response["message"])
	assert.Equal(t, "test_run_123", response["run_id"])
}

func TestErrorHandling(t *testing.T) {
	router := setupTestServer()

	// Test invalid JSON
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/validate", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test non-existent run
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/v1/runs/nonexistent", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestMiddleware(t *testing.T) {
	router := setupTestServer()

	// Test logging middleware
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	req.Header.Set("User-Agent", "test-agent")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Test error handling middleware
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/v1/runs/invalid", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
