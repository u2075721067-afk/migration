package main

import (
	"context"
	"log"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

var tracer oteltrace.Tracer

// InitTracing initializes OpenTelemetry tracing
func InitTracing() func() {
	// Create Jaeger exporter
	jaegerEndpoint := os.Getenv("JAEGER_ENDPOINT")
	if jaegerEndpoint == "" {
		jaegerEndpoint = "http://localhost:14268/api/traces"
	}

	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerEndpoint)))
	if err != nil {
		log.Printf("Failed to create Jaeger exporter: %v", err)
		// Return no-op cleanup function
		return func() {}
	}

	// Create resource
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String("mova-engine"),
			semconv.ServiceVersionKey.String("1.0.0"),
			semconv.DeploymentEnvironmentKey.String(getEnv("ENVIRONMENT", "development")),
		),
	)
	if err != nil {
		log.Printf("Failed to create resource: %v", err)
		return func() {}
	}

	// Create trace provider
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exp),
		trace.WithResource(res),
		trace.WithSampler(trace.TraceIDRatioBased(0.1)), // Sample 10% of traces
	)

	otel.SetTracerProvider(tp)
	tracer = tp.Tracer("mova-engine")

	return func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}
}

// StartWorkflowSpan starts a span for workflow execution
func StartWorkflowSpan(ctx context.Context, runID, intent string) (context.Context, oteltrace.Span) {
	return tracer.Start(ctx, "workflow.execute",
		oteltrace.WithAttributes(
			attribute.String("workflow.run_id", runID),
			attribute.String("workflow.intent", intent),
		),
	)
}

// StartActionSpan starts a span for action execution
func StartActionSpan(ctx context.Context, actionID, actionType string) (context.Context, oteltrace.Span) {
	return tracer.Start(ctx, "action.execute",
		oteltrace.WithAttributes(
			attribute.String("action.id", actionID),
			attribute.String("action.type", actionType),
		),
	)
}

// AddSpanEvent adds an event to the current span
func AddSpanEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	span := oteltrace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.AddEvent(name, oteltrace.WithAttributes(attrs...))
	}
}

// SetSpanError sets error information on the span
func SetSpanError(ctx context.Context, err error) {
	span := oteltrace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.RecordError(err)
		span.SetStatus(oteltrace.StatusCodeError, err.Error())
	}
}

// SetSpanAttributes sets attributes on the current span
func SetSpanAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	span := oteltrace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(attrs...)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
