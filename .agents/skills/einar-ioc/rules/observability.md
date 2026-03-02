# observability

> OpenTelemetry observability setup

## app/shared/infrastructure/observability/observability.go

```go
package observability

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"archetype/app/shared/configuration"

	"github.com/Ignaciojeria/ioc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

var _ = ioc.Register(NewObservability)

type Observability struct {
	Tracer trace.Tracer
	Logger *slog.Logger
}

// NewObservability sets up the OpenTelemetry Tracer and Slog Logger.
// It uses OTEL_EXPORTER_OTLP_ENDPOINT from the environment if present.
func NewObservability(conf configuration.Conf) (Observability, error) {
	// Configure global OpenTelemetry text map propagators to support trace continuity
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Setup Slog with JSON structure
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})).With(
		slog.String("service", conf.PROJECT_NAME),
		slog.String("version", conf.VERSION),
	)

	// Set as global default logger
	slog.SetDefault(logger)

	var tp *sdktrace.TracerProvider
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")

	if endpoint != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		exporter, err := otlptracehttp.New(ctx,
			otlptracehttp.WithEndpointURL(endpoint),
			otlptracehttp.WithInsecure(),
		)
		if err != nil {
			logger.Error("failed to create otlp exporter", "error", err)
			return Observability{}, fmt.Errorf("failed to create otlp exporter: %w", err)
		}

		res, err := resource.Merge(
			resource.Default(),
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceName(conf.PROJECT_NAME),
				semconv.ServiceVersion(conf.VERSION),
			),
		)
		if err != nil {
			return Observability{}, fmt.Errorf("failed to create resource: %w", err)
		}

		tp = sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(res),
		)
		otel.SetTracerProvider(tp)
		logger.Info("opentelemetry tracer initialized", "endpoint", endpoint)
	} else {
		// Fallback to No-op tracer if endpoint is not set to avoid breaking local dev
		tp = sdktrace.NewTracerProvider()
		otel.SetTracerProvider(tp)
		logger.Info("opentelemetry disabled, no OTEL_EXPORTER_OTLP_ENDPOINT environment variable set")
	}

	tracer := tp.Tracer(conf.PROJECT_NAME)

	return Observability{
		Tracer: tracer,
		Logger: logger,
	}, nil
}
```

---

## Unit tests

When creating a new component, generate tests following this pattern:

### app/shared/infrastructure/observability/observability_test.go

```go
package observability

import (
	"testing"

	"archetype/app/shared/configuration"

	"github.com/stretchr/testify/assert"
)

func TestNewObservability(t *testing.T) {
	conf := configuration.Conf{
		PROJECT_NAME: "test-svc",
		VERSION:      "1.0",
	}

	obs, err := NewObservability(conf)
	assert.NoError(t, err)
	assert.NotNil(t, obs.Tracer)
	assert.NotNil(t, obs.Logger)
}
```
