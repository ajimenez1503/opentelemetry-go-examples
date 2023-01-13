package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/sdk/metric"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func serverMetrics() {
	log.Println("Serving metrics at localhost:2223/metrics")
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(":2223", nil)
	if err != nil {
		fmt.Printf("error serving htpp: %v", err)
		return
	}
}

func main() {
	ctx := context.Background()

	exporter, err := prometheus.New()
	if err != nil {
		log.Fatal(err)
	}

	provider := metric.NewMeterProvider(metric.WithReader(exporter))
	meter := provider.Meter("github.com/softwarejimenez/opentelemetry-go-examples/prometheus")

	// start the prometheus HTTP server
	go serverMetrics()

	attrs := []attribute.KeyValue{
		attribute.Key("A").String("B"),
		attribute.Key("C").String("D"),
	}

	// counter
	counter, err := meter.SyncFloat64().Counter("foo", instrument.WithDescription("a simple counter"))
	if err != nil {
		log.Fatal(err)
	}
	counter.Add(ctx, 5, attrs...)

	gauge, err := meter.AsyncFloat64().Gauge("bar", instrument.WithDescription("a fun little gauge"))
	if err != nil {
		log.Fatal(err)
	}

	err = meter.RegisterCallback([]instrument.Asynchronous{gauge}, func(ctx context.Context) {
		n := -10. + rand.Float64()*(90.) // [-10, 100)
		gauge.Observe(ctx, n, attrs...)
	})
	if err != nil {
		log.Fatal(err)
	}

	// this is the equivalent of prometheus.NewHistogramVec
	histogram, err := meter.SyncFloat64().Histogram("baz", instrument.WithDescription("a very nice histogram"))
	if err != nil {
		log.Fatal(err)
	}
	histogram.Record(ctx, 23, attrs...)
	histogram.Record(ctx, 7, attrs...)
	histogram.Record(ctx, 101, attrs...)
	histogram.Record(ctx, 105, attrs...)

	ctx, _ = signal.NotifyContext(ctx, os.Interrupt)
	<-ctx.Done()
}
