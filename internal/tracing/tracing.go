// Package tracing wires OpenTelemetry for OpenUSS.
package tracing

import (
	"context"
	"log/slog"
	"os"

	"github.com/go-logr/logr"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
	"go.opentelemetry.io/otel/trace/noop"
)

const ServiceName = "openuss"
const EndpointVar = "OTEL_EXPORTER_OTLP_ENDPOINT"

// Init installs the propagator and tracer provider, returning a function that
// flushes pending spans.
func Init(ctx context.Context, logger *slog.Logger) (func(context.Context) error, error) {
	otel.SetLogger(logr.FromSlogHandler(logger.Handler()))
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		logger.ErrorContext(context.Background(), "opentelemetry error", slog.Any("error", err))
	}))

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	endpoint := os.Getenv(EndpointVar)
	if endpoint == "" {
		otel.SetTracerProvider(noop.NewTracerProvider())
		logger.InfoContext(ctx, "tracing disabled", slog.String("reason", EndpointVar+" unset"))
		return func(context.Context) error { return nil }, nil
	}

	exporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return nil, err
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(ServiceName),
		)),
	)
	otel.SetTracerProvider(provider)

	logger.InfoContext(ctx, "tracing enabled", slog.String("endpoint", endpoint))

	return provider.Shutdown, nil
}
