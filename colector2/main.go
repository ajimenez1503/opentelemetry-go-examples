package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

// Initialize the OTLP exporter
func initProvider() func() {
	ctx := context.Background()

	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String("demo-server"),
		),
	)
	if err != nil {
		log.Fatalf("Failed to create resource: %w", err)
	}

	otelAgentAddr := "0.0.0.0:4317"
	metricExp, err := otlpmetricgrpc.New(
		ctx,
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint(otelAgentAddr))
	if err != nil {
		log.Fatalf("Failed to create collector metric exporter: %w", err)
	}
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(
				metricExp,
				sdkmetric.WithInterval(2*time.Second),
			),
		),
	)

	global.SetMeterProvider(meterProvider)

	traceClient := otlptracegrpc.NewClient(
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(otelAgentAddr),
		otlptracegrpc.WithDialOption(grpc.WithBlock()))

	// Set up a trace exporter
	traceExp, err := otlptrace.New(ctx, traceClient)
	if err != nil {
		log.Fatalf("Failed to create trace exporter: %w", err)
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(traceExp)),
	)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	otel.SetTracerProvider(tracerProvider)

	// shutdown will flush any remainig spans and shut down the exporter
	return func() {
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		if err := traceExp.Shutdown(ctx); err != nil {
			otel.Handle(err)
		}
		if err := meterProvider.Shutdown(ctx); err != nil {
			otel.Handle(err)
		}
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdown := initProvider()
	defer shutdown()

	tracer := otel.Tracer("demo-server-tracer")
	meter := global.Meter("demo-server-meter")

	commonAttrs := []attribute.KeyValue{
		attribute.String("attrA", "chocolate"),
		attribute.String("attrB", "raspberry"),
		attribute.String("attrC", "vanilla"),
	}

	counter, _ := meter.SyncInt64().Counter(
		"demo_server/counter",
		instrument.WithDescription("A conter"),
	)

	// create span
	ctx, span := tracer.Start(ctx, "collector-example", trace.WithAttributes(commonAttrs...))
	counter.Add(ctx, 0, commonAttrs...)
	log.Printf("Start iteration")
	defer span.End()
	for i := 0; i < 10000; i++ {
		iCtx, iSpan := tracer.Start(ctx, fmt.Sprintf("sample-%d", i))
		counter.Add(iCtx, int64(i), commonAttrs...)
		<-time.After(time.Second)
		log.Printf("Iteration %d", i)
		iSpan.End()
	}

	log.Printf("Done!")
}
