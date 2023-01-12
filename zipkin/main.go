package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
)

var logger = log.New(os.Stderr, "zipkin-example", log.Ldate|log.Ltime|log.Llongfile)

// Create a trace and register ir as trace provider
func initTracer(url string) (func(context.Context) error, error) {
	// Create zipkin exporter
	exporter, err := zipkin.New(url, zipkin.WithLogger(logger))
	if err != nil {
		return nil, err
	}

	batcher := sdktrace.NewBatchSpanProcessor(exporter)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(batcher),
		sdktrace.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String("zipkin-test"),
			),
		),
	)
	otel.SetTracerProvider(tp)

	return tp.Shutdown, nil
}

func bar(ctx context.Context) {
	tr := otel.GetTracerProvider().Tracer("component-bar")
	_, span := tr.Start(ctx, "bar")
	defer span.End()
	<-time.After(6 * time.Millisecond)
}

func main() {
	url := "http://localhost:9411/api/v2/spans"

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	shutdown, err := initTracer(url)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			log.Fatal("Failed to shutdown TracerProvider: %w", err)
		}
	}()

	tr := otel.GetTracerProvider().Tracer("component-main")
	ctx, span := tr.Start(ctx, "foo", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()
	<-time.After(6 * time.Microsecond)
	bar(ctx)
	<-time.After(6 * time.Millisecond)
}
