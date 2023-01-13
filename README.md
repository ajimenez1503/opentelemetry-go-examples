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
![alt text](img/zipkin-example.png "Zipkin")


## Example 2 - Jaeger

Send an example span to a Jaeger service.
- Ref: https://github.com/open-telemetry/opentelemetry-go/tree/main/example/jaeger

- Running:
```
docker run --name jaeger \
  -e COLLECTOR_ZIPKIN_HTTP_PORT=9411 \
  -p 5775:5775/udp \
  -p 6831:6831/udp \
  -p 6832:6832/udp \
  -p 5778:5778 \
  -p 16686:16686 \
  -p 14268:14268 \
  -p 9411:9411 \
  jaegertracing/all-in-one:1.6
go run main.go
```
- Open Jaeger with the trace writen http://localhost:16686/trace/ec050c45678c972950e818fa88b36fad

![alt text](img/jaeger.png "Zipkin")
