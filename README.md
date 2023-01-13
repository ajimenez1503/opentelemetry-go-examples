# opentelemetry-go-examples
Example of open telemetry in Go

## Example 1 - Fibonacci
- Ref:
    - https://opentelemetry.io/docs/instrumentation/go/getting-started/
    - https://github.com/open-telemetry/opentelemetry-go/tree/main/example/fib

- Running:
```
cd fibonacci
go run .
cat traces.txt
```

## Example 2 - Zipkin

Send an example span to a Zipkin service.
- Ref: https://github.com/open-telemetry/opentelemetry-go/tree/main/example/zipkin

- Running:
```
docker run -p 9411:9411 openzipkin/zipkin
go run main.go
```
- Open zipkin with the trace writen http://localhost:9411/zipkin/traces/62da195c8008d818e14d5a70fed2fc54
