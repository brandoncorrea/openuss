package tracing_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"bwawan.com/openuss/internal/logging/logtest"
	"bwawan.com/openuss/internal/tracing"
)

func sampledContext(t *testing.T) context.Context {
	t.Helper()

	traceID, err := trace.TraceIDFromHex("0102030405060708090a0b0c0d0e0f10")
	require.NoError(t, err)

	spanID, err := trace.SpanIDFromHex("1112131415161718")
	require.NoError(t, err)

	return trace.ContextWithSpanContext(context.Background(), trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
	}))
}

func TestInitInstallsW3CPropagation(t *testing.T) {
	logger, _ := logtest.New()

	_, err := tracing.Init(context.Background(), logger)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, "http://dss.example/dss/v1/operational_intent_references", nil)
	require.NoError(t, err)

	otel.GetTextMapPropagator().Inject(sampledContext(t), propagation.HeaderCarrier(req.Header))

	require.Equal(t,
		"00-0102030405060708090a0b0c0d0e0f10-1112131415161718-01",
		req.Header.Get("traceparent"))
}

func TestInitWithoutEndpointDisablesExport(t *testing.T) {
	t.Setenv(tracing.EndpointVar, "")

	logger, rec := logtest.New()

	shutdown, err := tracing.Init(context.Background(), logger)
	require.NoError(t, err)
	require.NoError(t, shutdown(context.Background()))

	require.NotNil(t, rec.Find("tracing disabled"), "expected a disabled-tracing log line")
	require.Nil(t, rec.Find("tracing enabled"))

	// A no-op provider yields unsampled spans, so nothing reaches the logs.
	_, span := otel.Tracer("test").Start(context.Background(), "op")
	defer span.End()
	require.False(t, span.SpanContext().IsSampled())
}

func TestInitWithEndpointEnablesExport(t *testing.T) {
	t.Setenv(tracing.EndpointVar, "http://127.0.0.1:4318")

	logger, rec := logtest.New()

	shutdown, err := tracing.Init(context.Background(), logger)
	require.NoError(t, err)
	t.Cleanup(func() { _ = shutdown(context.Background()) })

	entry := rec.Find("tracing enabled")
	require.NotNil(t, entry, "expected an enabled-tracing log line")
	require.Equal(t, "http://127.0.0.1:4318", entry["endpoint"])

	// Spans are now real, which is what puts trace_id into the logs.
	_, span := otel.Tracer("test").Start(context.Background(), "op")
	defer span.End()
	require.True(t, span.SpanContext().IsValid())
}

func TestInitWithoutEndpointResetsAnExistingProvider(t *testing.T) {
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample())))
	t.Setenv(tracing.EndpointVar, "")

	logger, _ := logtest.New()
	_, err := tracing.Init(context.Background(), logger)
	require.NoError(t, err)

	_, span := otel.Tracer("test").Start(context.Background(), "op")
	defer span.End()
	require.False(t, span.SpanContext().IsSampled(), "Init must disable tracing, not merely decline to enable it")
}

func TestInitRoutesOTelErrorsToTheLogger(t *testing.T) {
	logger, rec := logtest.New()

	_, err := tracing.Init(context.Background(), logger)
	require.NoError(t, err)

	otel.Handle(errors.New("export failed"))

	entry := rec.Find("opentelemetry error")
	require.NotNil(t, entry, "SDK errors must reach the JSON log stream")
	require.Equal(t, "export failed", entry["error"])
}

func TestInitRoutesSDKDiagnosticsToTheLogger(t *testing.T) {
	t.Setenv(tracing.EndpointVar, "https://collector.invalid:4318")
	t.Setenv("OTEL_EXPORTER_OTLP_CERTIFICATE", "/nonexistent/ca.pem")

	logger, rec := logtest.New()

	shutdown, err := tracing.Init(context.Background(), logger)
	require.NoError(t, err)
	t.Cleanup(func() { _ = shutdown(context.Background()) })

	require.NotNil(t, rec.Find("read tls ca cert file"),
		"a misconfigured collector must not fail silently")
}
