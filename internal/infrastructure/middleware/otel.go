package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// InitOTEL initializes OpenTelemetry
func InitOTEL() (func(context.Context) error, error) {
	// Check if OTEL is enabled
	if os.Getenv("OTEL_ENABLED") != "true" {
		return func(context.Context) error { return nil }, nil
	}

	// Get OTEL configuration
	serviceName := os.Getenv("OTEL_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "e-commerce-api"
	}

	// Create exporter
	ctx := context.Background()
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:4317"
	}

	exporter, err := otlptrace.New(
		ctx,
		otlptracegrpc.NewClient(
			otlptracegrpc.WithEndpoint(endpoint),
			otlptracegrpc.WithInsecure(),
		),
	)
	if err != nil {
		return nil, err
	}

	// Create resource
	res, err := resource.New(ctx,
		resource.WithAttributes(
			attribute.String("service.name", serviceName),
		),
	)
	if err != nil {
		return nil, err
	}

	// Create trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	// Set the global tracer
	return tp.Shutdown, nil
}

// TraceMiddleware adds OpenTelemetry tracing to requests
func TraceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip if OTEL is disabled
		if os.Getenv("OTEL_ENABLED") != "true" {
			next.ServeHTTP(w, r)
			return
		}

		// Get the current tracer
		tracer := otel.Tracer("api-server")

		// Get request details for span
		requestID := r.Header.Get(RequestIDHeader)
		method := r.Method
		path := r.URL.Path

		// Create a span for this request
		ctx, span := tracer.Start(
			r.Context(),
			"http.request",
			trace.WithAttributes(
				attribute.String("http.method", method),
				attribute.String("http.path", path),
				attribute.String("request.id", requestID),
			),
		)
		defer span.End()

		// Add the span context to the request
		r = r.WithContext(ctx)

		// Create custom response writer to capture status code
		customWriter := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Process the request
		next.ServeHTTP(customWriter, r)

		// Record response status
		span.SetAttributes(attribute.Int("http.status_code", customWriter.statusCode))

		// Mark span as error if status code is 4xx or 5xx
		if customWriter.statusCode >= 400 {
			span.SetAttributes(attribute.Bool("error", true))
		}
	})
}

// LogWithContext adds trace context to slog events
func LogWithContext(ctx context.Context, level slog.Level, msg string, args ...any) {
	// Get span from context
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		// Add trace and span IDs to the log
		args = append(args,
			slog.String("trace_id", span.SpanContext().TraceID().String()),
			slog.String("span_id", span.SpanContext().SpanID().String()),
		)
	}

	// Log with added context
	Logger.Log(ctx, level, msg, args...)
}
