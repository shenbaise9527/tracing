// +build !jaeger

package tracing

import (
	"io"

	"github.com/opentracing/opentracing-go"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
)

type noopZkCloser struct{}

func (noopZkCloser) Close() error {
	return nil
}

// newTracer 创建基于zipkin的tracer对象.
func newTracer(tracingURL, serverName, localEndpoint string) (opentracing.Tracer, io.Closer, error) {
	zipkinReporter := zipkinhttp.NewReporter(tracingURL)
	endpoint, err := zipkin.NewEndpoint(serverName, localEndpoint)
	if err != nil {
		return nil, nil, err
	}

	nativeTracer, err := zipkin.NewTracer(zipkinReporter, zipkin.WithLocalEndpoint(endpoint))
	if err != nil {
		return nil, nil, err
	}

	return zipkinot.Wrap(nativeTracer), noopZkCloser{}, nil
}
