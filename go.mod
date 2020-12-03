module github.com/shenbaise9527/tracing

go 1.14

require (
	github.com/HdrHistogram/hdrhistogram-go v1.0.1 // indirect
	github.com/opentracing/opentracing-go v1.2.0
	github.com/openzipkin-contrib/zipkin-go-opentracing v0.4.5
	github.com/openzipkin/zipkin-go v0.2.5
	github.com/pkg/errors v0.9.1 // indirect
	github.com/uber/jaeger-client-go v2.25.0+incompatible
	github.com/uber/jaeger-lib v2.4.0+incompatible // indirect
	go.uber.org/atomic v1.7.0 // indirect
	google.golang.org/grpc v1.34.0
)

replace google.golang.org/grpc v1.34.0 => google.golang.org/grpc v1.29.1
