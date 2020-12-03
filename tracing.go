package tracing

import (
	"context"
	"io"
	"net/http"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// NewOpenTracer 初始化opentracing,第一个返回值在外部应该调用close方法.
func NewOpenTracer(tracingURL, serverName, localEndpoint string) (io.Closer, error) {
	tracer, closer, err := newTracer(tracingURL, serverName, localEndpoint)
	if err != nil {
		return nil, err
	}

	opentracing.InitGlobalTracer(tracer)

	return closer, nil
}

// HttpTracing http.Handler.
func HttpTracing(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tracer := opentracing.GlobalTracer()
		carrier := opentracing.HTTPHeadersCarrier(r.Header)
		clientspan, err := tracer.Extract(opentracing.HTTPHeaders, carrier)
		var span opentracing.Span
		if err == nil {
			span = tracer.StartSpan(r.RequestURI, opentracing.ChildOf(clientspan))
		} else {
			span = tracer.StartSpan(r.RequestURI)
		}

		defer span.Finish()
		rc := opentracing.ContextWithSpan(r.Context(), span)
		r = r.WithContext(rc)
		next(w, r)
	}
}

// OpenTracingServerInterceptor grpc unary serverinterceptor.
func OpenTracingServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		md, ok := metadata.FromIncomingContext(ctx)
		var span opentracing.Span
		tracer := opentracing.GlobalTracer()
		if !ok {
			span = tracer.StartSpan(info.FullMethod)
		} else {
			carrier := make(opentracing.TextMapCarrier)
			for k, v := range md {
				carrier.Set(k, v[0])
			}

			spanctx, err := tracer.Extract(opentracing.TextMap, carrier)
			if err != nil {
				span = tracer.StartSpan(info.FullMethod)
			} else {
				span = tracer.StartSpan(info.FullMethod, opentracing.ChildOf(spanctx))
			}
		}

		defer span.Finish()
		ctx = opentracing.ContextWithSpan(ctx, span)
		return handler(ctx, req)
	}
}

// OpenTracingClientInterceptor grpc unary clientinterceptor.
func OpenTracingClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		span:= opentracing.SpanFromContext(ctx)
		if span != nil {
			tracer := span.Tracer()
			clientspan := tracer.StartSpan(method, opentracing.ChildOf(span.Context()))
			defer clientspan.Finish()
			carrier := make(opentracing.TextMapCarrier)
			err := tracer.Inject(clientspan.Context(), opentracing.TextMap, carrier)
			if err == nil {
				var pairs []string
				_ = carrier.ForeachKey(func(key, val string) error {
					pairs = append(pairs, key, val)
					return nil
				})

				ctx = metadata.AppendToOutgoingContext(ctx, pairs...)
			}
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// GetSpanFromContext 从context中获取span,只适用于http和serverInterceptor的context.
func GetSpanFromContext(ctx context.Context) opentracing.Span {
	return opentracing.SpanFromContext(ctx)
}

// ChildOfSpanFromContext 根据context中的span生成ChildOf的span.
func ChildOfSpanFromContext(ctx context.Context, operationName string) opentracing.Span {
	return newSubSpanFromContext(ctx, operationName, opentracing.ChildOf)
}

// FollowsSpanFromContext 根据context中的span生成FollowsFrom的span.
func FollowsSpanFromContext(ctx context.Context, operationName string) opentracing.Span {
	return newSubSpanFromContext(ctx, operationName, opentracing.FollowsFrom)
}

func newSubSpanFromContext(
	ctx context.Context,
	operationName string,
	op func(opentracing.SpanContext) opentracing.SpanReference) opentracing.Span {
	tracer := opentracing.GlobalTracer()
	span := GetSpanFromContext(ctx)
	if span == nil {
		span = tracer.StartSpan(operationName)
	} else {
		span = tracer.StartSpan(operationName, op(span.Context()))
	}

	return span
}
