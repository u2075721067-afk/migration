package main

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP metrics
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"endpoint", "method", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"endpoint", "method"},
	)

	// Workflow metrics
	workflowRunsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "workflow_runs_total",
			Help: "Total number of workflow runs",
		},
		[]string{"status"},
	)

	workflowDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "workflow_duration_seconds",
			Help:    "Workflow execution duration in seconds",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60, 120, 300},
		},
		[]string{"status"},
	)

	// Executor metrics
	executorActionTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "executor_action_total",
			Help: "Total number of executor actions",
		},
		[]string{"type", "status"},
	)

	executorActionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "executor_action_duration_seconds",
			Help:    "Executor action duration in seconds",
			Buckets: []float64{0.001, 0.01, 0.1, 0.5, 1, 2, 5, 10},
		},
		[]string{"type"},
	)

	// System metrics
	activeGoroutines = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_goroutines",
			Help: "Number of active goroutines",
		},
	)

	memoryUsage = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "memory_usage_bytes",
			Help: "Memory usage in bytes",
		},
	)
)

// PrometheusMiddleware creates a middleware for Prometheus metrics
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		// Skip metrics endpoint itself
		if path == "/metrics" {
			c.Next()
			return
		}

		c.Next()

		// Record metrics
		status := strconv.Itoa(c.Writer.Status())
		duration := time.Since(start).Seconds()

		httpRequestsTotal.WithLabelValues(path, c.Request.Method, status).Inc()
		httpRequestDuration.WithLabelValues(path, c.Request.Method).Observe(duration)
	}
}

// RecordWorkflowMetrics records workflow execution metrics
func RecordWorkflowMetrics(status string, duration time.Duration) {
	workflowRunsTotal.WithLabelValues(status).Inc()
	workflowDuration.WithLabelValues(status).Observe(duration.Seconds())
}

// RecordExecutorMetrics records executor action metrics
func RecordExecutorMetrics(actionType, status string, duration time.Duration) {
	executorActionTotal.WithLabelValues(actionType, status).Inc()
	executorActionDuration.WithLabelValues(actionType).Observe(duration.Seconds())
}

// UpdateSystemMetrics updates system-level metrics
func UpdateSystemMetrics(goroutines int, memBytes uint64) {
	activeGoroutines.Set(float64(goroutines))
	memoryUsage.Set(float64(memBytes))
}

