// +build jaeger

package tracing

import (
	"io"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/transport"
)

func newTracer(tracingURL, serverName, localEndpoint string) (opentracing.Tracer, io.Closer, error) {
	cfg := jaegercfg.Configuration{
		ServiceName: serverName,
		Sampler: &jaegercfg.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
	}

	sender := transport.NewHTTPTransport(tracingURL)
	jaegerReporter := jaeger.NewRemoteReporter(sender)
	tracer, closer, err := cfg.NewTracer(
		jaegercfg.Reporter(jaegerReporter),
		jaegercfg.Logger(jaeger.NullLogger),
		jaegercfg.PoolSpans(true))
	if err != nil {
		return nil, nil, err
	}

	return tracer, closer, nil
}
