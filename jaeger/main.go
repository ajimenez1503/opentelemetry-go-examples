package main

import (
	"context"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

const (
	service    = "trace-demo"
	enviroment = "dev"
	id         = 1
)

// Get TraceProvider configured to use jaeger
func traceProvider(url string) (*sdktrace.TracerProvider, error) {
	// Create zipkin exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(service),
				attribute.String("enviroment", enviroment),
				attribute.Int64("ID", id),
			),
		),
	)
	return tp, nil
}

func bar(ctx context.Context) {
	tr := otel.Tracer("component-bar")
	_, span := tr.Start(ctx, "bar")
	span.SetAttributes(attribute.Key("testset").String("value"))
	defer span.End()
}

func main() {
	url := "http://localhost:14268/api/traces"

	tp, err := traceProvider(url)
	if err != nil {
		log.Fatal(err)
	}

	// Register the traceProvider
	otel.SetTracerProvider(tp)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// shutdown traceProvider
	defer func(ctx context.Context) {
		ctx, cancel = context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}(ctx)

	tr := otel.Tracer("component-main")
	ctx, span := tr.Start(ctx, "foo")
	defer span.End()
	bar(ctx)
}
