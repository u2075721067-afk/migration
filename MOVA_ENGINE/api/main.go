package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mova-engine/mova-engine/core/executor"
	"github.com/mova-engine/mova-engine/core/validator"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func main() {
	// Initialize tracing
	cleanup := InitTracing()
	defer cleanup()

	// Log system startup
	LogSystemEvent("startup", logrus.Fields{
		"version":    "1.0.0",
		"go_version": runtime.Version(),
	})

	// Initialize executor and validator
	exec := executor.NewExecutor()
	v, err := validator.NewValidator("./schemas")
	if err != nil {
		LogError("main", "validator_init", err, nil)
		log.Fatalf("Failed to initialize validator: %v", err)
	}

	// Initialize DLQ and retry manager
	dlqPath := getEnv("MOVA_DLQ_PATH", "./state/deadletter")
	dlq := executor.NewDeadLetterQueue(dlqPath)
	retryManager := executor.NewRetryManager(dlq)
	dlqService := NewDLQService(dlq, retryManager, exec)

	// Set Gin mode
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	router := gin.Default()

	// Add middleware
	router.Use(StructuredLoggingMiddleware())
	router.Use(gin.Recovery())
	router.Use(errorHandlingMiddleware())
	router.Use(PrometheusMiddleware())
	router.Use(otelgin.Middleware("mova-engine"))

	// API routes
	v1 := router.Group("/v1")
	{
		// Workflow execution
		v1.POST("/execute", handleExecute(exec))
		v1.GET("/runs/:id", handleGetRun(exec))
		v1.GET("/runs/:id/logs", handleGetRunLogs(exec))
		v1.DELETE("/runs/:id", handleCancelRun(exec))

		// Validation
		v1.POST("/validate", handleValidate(v))

		// Schemas
		v1.GET("/schemas", handleGetSchemas)
		v1.GET("/schemas/:name", handleGetSchema)

		// Introspection
		v1.GET("/introspect", handleIntrospect)

		// Dead Letter Queue
		dlqGroup := v1.Group("/dlq")
		{
			dlqGroup.GET("", dlqService.handleListDLQEntries)
			dlqGroup.GET("/:id", dlqService.handleGetDLQEntry)
			dlqGroup.POST("/:id/retry", dlqService.handleRetryDLQEntry)
			dlqGroup.PUT("/:id/status", dlqService.handleUpdateDLQStatus)
			dlqGroup.POST("/:id/archive", dlqService.handleArchiveDLQEntry)
			dlqGroup.DELETE("/:id", dlqService.handleDeleteDLQEntry)
			dlqGroup.GET("/stats", dlqService.handleDLQStats)
		}
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"version":   "1.0.0",
		})
	})

	// Metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Start system metrics updater
	go updateSystemMetrics()

	// Start server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		LogSystemEvent("server_start", logrus.Fields{
			"addr": srv.Addr,
		})
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			LogError("main", "server_start", err, nil)
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	LogSystemEvent("shutdown_start", nil)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		LogError("main", "server_shutdown", err, nil)
		log.Fatal("Server forced to shutdown:", err)
	}

	LogSystemEvent("shutdown_complete", nil)
}

// handleExecute handles workflow execution requests
func handleExecute(exec executor.ExecutorInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		var envelope executor.MOVAEnvelope
		if err := c.ShouldBindJSON(&envelope); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request body",
				"details": err.Error(),
			})
			return
		}

		// Check if synchronous execution is requested
		wait := c.Query("wait") == "true"

		if wait {
			// Synchronous execution
			ctx := c.Request.Context()
			result, err := exec.Execute(ctx, envelope)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Execution failed",
					"details": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, result)
		} else {
			// Asynchronous execution
			runID := generateRunID()

			// Start execution in background
			go func() {
				ctx := context.Background()
				exec.Execute(ctx, envelope)
			}()

			c.JSON(http.StatusAccepted, gin.H{
				"run_id":  runID,
				"status":  "accepted",
				"message": "Execution started asynchronously",
			})
		}
	}
}

// handleGetRun retrieves execution status
func handleGetRun(exec executor.ExecutorInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		runID := c.Param("id")
		if runID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Run ID is required",
			})
			return
		}

		result, err := exec.GetExecutionStatus(runID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Run not found",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

// handleGetRunLogs retrieves execution logs
func handleGetRunLogs(exec executor.ExecutorInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		runID := c.Param("id")
		if runID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Run ID is required",
			})
			return
		}

		logs, err := exec.GetExecutionLogs(runID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Run logs not found",
				"details": err.Error(),
			})
			return
		}

		// Return logs as JSONL
		c.Header("Content-Type", "application/jsonl")
		for _, log := range logs {
			c.Writer.Write([]byte(log + "\n"))
		}
	}
}

// handleCancelRun cancels execution
func handleCancelRun(exec executor.ExecutorInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		runID := c.Param("id")
		if runID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Run ID is required",
			})
			return
		}

		err := exec.CancelExecution(runID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to cancel execution",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Execution cancelled successfully",
			"run_id":  runID,
		})
	}
}

// handleValidate validates MOVA envelope
func handleValidate(v validator.ValidatorInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		var envelope map[string]interface{}
		if err := c.ShouldBindJSON(&envelope); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request body",
				"details": err.Error(),
			})
			return
		}

		// Convert to JSON string for validation
		envelopeBytes, err := json.Marshal(envelope)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to process envelope",
				"details": err.Error(),
			})
			return
		}

		// Create temporary file for validation
		tmpFile, err := os.CreateTemp("", "mova-validate-*.json")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to create temporary file",
				"details": err.Error(),
			})
			return
		}
		defer os.Remove(tmpFile.Name())

		if _, err := tmpFile.Write(envelopeBytes); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to write temporary file",
				"details": err.Error(),
			})
			return
		}
		tmpFile.Close()

		// Validate envelope
		valid, errors := v.ValidateEnvelope(tmpFile.Name())

		if valid {
			c.JSON(http.StatusOK, gin.H{
				"valid":   true,
				"message": "Envelope is valid",
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"valid":   false,
				"message": "Envelope validation failed",
				"errors":  errors,
			})
		}
	}
}

// handleGetSchemas returns available schemas
func handleGetSchemas(c *gin.Context) {
	schemas := []gin.H{
		{
			"name":        "envelope",
			"version":     "3.1",
			"description": "MOVA v3.1 envelope schema",
			"url":         "/v1/schemas/envelope",
		},
		{
			"name":        "action",
			"version":     "3.1",
			"description": "MOVA v3.1 action schema",
			"url":         "/v1/schemas/action",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"schemas": schemas,
	})
}

// handleGetSchema returns specific schema
func handleGetSchema(c *gin.Context) {
	schemaName := c.Param("name")

	var schemaPath string
	switch schemaName {
	case "envelope":
		schemaPath = "../../schemas/envelope.json"
	case "action":
		schemaPath = "../../schemas/action.json"
	default:
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Schema not found",
		})
		return
	}

	// Read and return schema content
	schemaBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to read schema",
			"details": err.Error(),
		})
		return
	}

	var schema map[string]interface{}
	if err := json.Unmarshal(schemaBytes, &schema); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to parse schema",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, schema)
}

// handleIntrospect returns API information
func handleIntrospect(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"name":         "MOVA Automation Engine API",
		"version":      "1.0.0",
		"description":  "REST API for MOVA workflow execution",
		"mova_version": "3.1",
		"endpoints": []gin.H{
			{
				"method":       "POST",
				"path":         "/v1/execute",
				"description":  "Execute a MOVA workflow",
				"query_params": []string{"wait"},
			},
			{
				"method":      "GET",
				"path":        "/v1/runs/:id",
				"description": "Get execution status",
			},
			{
				"method":      "GET",
				"path":        "/v1/runs/:id/logs",
				"description": "Get execution logs (JSONL)",
			},
			{
				"method":      "DELETE",
				"path":        "/v1/runs/:id",
				"description": "Cancel execution",
			},
			{
				"method":      "POST",
				"path":        "/v1/validate",
				"description": "Validate MOVA envelope",
			},
			{
				"method":      "GET",
				"path":        "/v1/schemas",
				"description": "List available schemas",
			},
			{
				"method":      "GET",
				"path":        "/v1/schemas/:name",
				"description": "Get specific schema",
			},
			{
				"method":      "GET",
				"path":        "/v1/introspect",
				"description": "API information",
			},
		},
		"supported_actions": []string{
			"set", "if", "repeat", "call", "parallel", "sleep",
			"parse_json", "http_fetch", "mcp_call", "print",
		},
	})
}

// Middleware functions
func loggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

func errorHandlingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Handle errors after request processing
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal server error",
				"details": err.Error(),
			})
		}
	}
}

// Helper functions
func generateRunID() string {
	return fmt.Sprintf("run_%d", time.Now().UnixNano())
}

// updateSystemMetrics periodically updates system metrics
func updateSystemMetrics() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			var m runtime.MemStats
			runtime.ReadMemStats(&m)

			UpdateSystemMetrics(
				runtime.NumGoroutine(),
				m.Alloc,
			)
		}
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
