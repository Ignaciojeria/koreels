package observability

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"koreels/internal/shared/configuration"

	"github.com/Ignaciojeria/ioc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
	"go.opentelemetry.io/otel/trace"
)

var _ = ioc.Register(NewObservability)

type Observability struct {
	Tracer trace.Tracer
	Logger *slog.Logger
}

func NewObservability(conf configuration.Conf) (Observability, error) {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})).With(
		slog.String("service", conf.PROJECT_NAME),
		slog.String("version", conf.VERSION),
	)
	slog.SetDefault(logger)

	var tp *sdktrace.TracerProvider
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")

	if endpoint != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		exporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(endpoint), otlptracehttp.WithInsecure())
		if err != nil {
			logger.Error("failed to create otlp exporter", "error", err)
			return Observability{}, fmt.Errorf("failed to create otlp exporter: %w", err)
		}
		res, err := resource.Merge(resource.Default(),
			resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceName(conf.PROJECT_NAME), semconv.ServiceVersion(conf.VERSION)))
		if err != nil {
			return Observability{}, fmt.Errorf("failed to create resource: %w", err)
		}
		tp = sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter), sdktrace.WithResource(res))
		otel.SetTracerProvider(tp)
	} else {
		tp = sdktrace.NewTracerProvider()
		otel.SetTracerProvider(tp)
	}

	return Observability{Tracer: tp.Tracer(conf.PROJECT_NAME), Logger: logger}, nil
}
