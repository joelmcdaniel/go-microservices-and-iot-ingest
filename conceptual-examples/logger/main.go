package main

import (
	"context"
	"log/slog"
	"os"
)

// Define a custom type for context keys to prevent collisions
type contextKey string

const traceIDKey contextKey = "trace_id"

// Helper to get TraceID from context
func GetTraceID(ctx context.Context) string {
	if id, ok := ctx.Value(traceIDKey).(string); ok {
		return id
	}
	return "0000-0000" // Fallback if no trace ID exists
}

// Simulate an HTTP middleware or MQTT interceptor
func WithTraceID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, traceIDKey, id)
}

type ContextHandler struct {
	slog.Handler
}

// Handle is called for every log statement (Info, Error, etc.)
func (h ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	// 1. Extract Trace ID from the context
	if id := GetTraceID(ctx); id != "" {
		// 2. Add it to the log record
		r.AddAttrs(slog.String("trace_id", id))
	}
	// 3. Pass the enriched record to the underlying JSON handler
	return h.Handler.Handle(ctx, r)
}

func main() {
	// 1. Setup the Logger
	// We want JSON output to Stdout
	baseHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo, // Only log INFO and above
	})

	// Wrap it with our automatic Context Handler
	logger := slog.New(ContextHandler{baseHandler})
	// Set it as the global default
	slog.SetDefault(logger)

	// --- SIMULATING A REQUEST ---
	// 2. A request arrives at the Gateway
	// We generate a Trace ID: "req-12345"
	ctx := context.Background()
	ctx = WithTraceID(ctx, "req-12345")

	slog.InfoContext(ctx, "Request received at Gateway")

	// 3. Pass control to business logic
	ProcessSensorData(ctx, "sensor-99")
}

func ProcessSensorData(ctx context.Context, sensorID string) {
	// Notice: We don't manually pass "req-12345" here.
	// We just pass the context.
	slog.InfoContext(ctx, "Processing sensor data",
		slog.String("sensor_id", sensorID),
		slog.Int("temperature", 42),
	)

	// Simulate a DB call that fails
	// The log will still have the Trace ID!
	slog.ErrorContext(ctx, "Database connection failed",
		slog.String("db_host", "192.168.1.5"),
	)
}
