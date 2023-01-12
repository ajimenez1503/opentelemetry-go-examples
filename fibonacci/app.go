package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"strconv"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// Trace name used to identiy the instrumentation library
const name = "github.com/softwarejimenez/opentelemetry-go-examples/fibonacci"

// App is a fibonacci computation application
type App struct {
	r io.Reader
	l *log.Logger
}

// Return a new App
func NewApp(r io.Reader, l *log.Logger) *App {
	return &App{r, l}
}

// Poll ask the user for input and returns the request
func (a *App) Poll(ctx context.Context) (uint, error) {
	_, span := otel.Tracer(name).Start(ctx, "Poll")
	defer span.End()

	a.l.Print("What Fibonacci number would you like to know:")
	var n uint
	_, err := fmt.Fscanf(a.r, "%d\n", &n)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return 0, err
	}

	// Save the request n into nStr
	nStr := strconv.FormatUint(uint64(n), 10)
	span.SetAttributes(attribute.String("request.n", nStr))

	return n, nil
}

// Write writes the n-th Fibonacci number back to the user.
func (a *App) Write(ctx context.Context, n uint) {
	ctx, span := otel.Tracer(name).Start(ctx, "write")
	defer span.End()

	f, err := func(ctx context.Context) (uint64, error) {
		_, span := otel.Tracer(name).Start(ctx, "Fibonacci")
		defer span.End()
		result, err := Fibonacci(n)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		return result, err
	}(ctx)

	if err != nil {
		a.l.Printf("Fibonacci(%d): %v\n", n, err)
	} else {
		a.l.Printf("Fibonacci(%d): %d\n", n, f)
	}
}

// Run starts polling euser for Fibonacci number requests and writes results
func (a *App) Run(ctx context.Context) error {
	for {
		ctx, span := otel.Tracer(name).Start(ctx, "Run")
		defer span.End()

		n, err := a.Poll(ctx)
		if err != nil {
			return err
		}
		a.Write(ctx, n)
	}
}
