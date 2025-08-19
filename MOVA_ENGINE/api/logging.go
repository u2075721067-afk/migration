package main

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

func init() {
	logger = logrus.New()

	// Set JSON formatter for structured logging
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
		},
	})

	// Set log level from environment
	level := os.Getenv("MOVA_LOG_LEVEL")
	switch level {
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "info":
		logger.SetLevel(logrus.InfoLevel)
	case "warn":
		logger.SetLevel(logrus.WarnLevel)
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

	// Output to stdout for container environments
	logger.SetOutput(os.Stdout)
}

// StructuredLoggingMiddleware creates a structured logging middleware
func StructuredLoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Log to structured logger
		fields := logrus.Fields{
			"method":       param.Method,
			"path":         param.Path,
			"status":       param.StatusCode,
			"latency":      param.Latency.Nanoseconds(),
			"latency_ms":   param.Latency.Milliseconds(),
			"client_ip":    param.ClientIP,
			"user_agent":   param.Request.UserAgent(),
			"request_id":   param.Request.Header.Get("X-Request-ID"),
			"content_type": param.Request.Header.Get("Content-Type"),
		}

		if param.ErrorMessage != "" {
			fields["error"] = param.ErrorMessage
			logger.WithFields(fields).Error("HTTP request completed with error")
		} else {
			logger.WithFields(fields).Info("HTTP request completed")
		}

		// Return empty string since we're using structured logger
		return ""
	})
}

// LogWorkflowStart logs workflow execution start
func LogWorkflowStart(runID, intent string, actionCount int) {
	logger.WithFields(logrus.Fields{
		"event":        "workflow_start",
		"run_id":       runID,
		"intent":       intent,
		"action_count": actionCount,
	}).Info("Workflow execution started")
}

// LogWorkflowComplete logs workflow execution completion
func LogWorkflowComplete(runID string, status string, duration time.Duration, actionCount int) {
	logger.WithFields(logrus.Fields{
		"event":        "workflow_complete",
		"run_id":       runID,
		"status":       status,
		"duration_ms":  duration.Milliseconds(),
		"duration_ns":  duration.Nanoseconds(),
		"action_count": actionCount,
	}).Info("Workflow execution completed")
}

// LogActionStart logs action execution start
func LogActionStart(runID, actionID, actionType string, params map[string]interface{}) {
	logger.WithFields(logrus.Fields{
		"event":       "action_start",
		"run_id":      runID,
		"action_id":   actionID,
		"action_type": actionType,
		"params":      params,
	}).Debug("Action execution started")
}

// LogActionComplete logs action execution completion
func LogActionComplete(runID, actionID, actionType, status string, duration time.Duration, result interface{}) {
	fields := logrus.Fields{
		"event":       "action_complete",
		"run_id":      runID,
		"action_id":   actionID,
		"action_type": actionType,
		"status":      status,
		"duration_ms": duration.Milliseconds(),
		"duration_ns": duration.Nanoseconds(),
	}

	if result != nil {
		fields["result"] = result
	}

	if status == "error" {
		logger.WithFields(fields).Error("Action execution failed")
	} else {
		logger.WithFields(fields).Debug("Action execution completed")
	}
}

// LogError logs application errors
func LogError(component, operation string, err error, fields logrus.Fields) {
	if fields == nil {
		fields = logrus.Fields{}
	}

	fields["component"] = component
	fields["operation"] = operation
	fields["error"] = err.Error()

	logger.WithFields(fields).Error("Application error occurred")
}

// LogValidationError logs validation errors
func LogValidationError(runID, validationType string, errors []string) {
	logger.WithFields(logrus.Fields{
		"event":           "validation_error",
		"run_id":          runID,
		"validation_type": validationType,
		"errors":          errors,
		"error_count":     len(errors),
	}).Warn("Validation failed")
}

// LogSecurityEvent logs security-related events
func LogSecurityEvent(event, clientIP, userAgent string, fields logrus.Fields) {
	if fields == nil {
		fields = logrus.Fields{}
	}

	fields["event"] = "security_" + event
	fields["client_ip"] = clientIP
	fields["user_agent"] = userAgent

	logger.WithFields(fields).Warn("Security event detected")
}

// LogSystemEvent logs system-level events
func LogSystemEvent(event string, fields logrus.Fields) {
	if fields == nil {
		fields = logrus.Fields{}
	}

	fields["event"] = "system_" + event

	logger.WithFields(fields).Info("System event")
}
