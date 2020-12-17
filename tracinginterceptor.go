package tracing

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// OpenTracingClientInterceptor grpc unary clientinterceptor.
func OpenTracingClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		span := ChildOfSpanFromContext(ctx, method)
		ext.Component.Set(span, "grpc")
		ext.SpanKindRPCClient.Set(span)
		defer span.Finish()
		carrier := make(opentracing.TextMapCarrier)
		tracer := opentracing.GlobalTracer()
		err := tracer.Inject(span.Context(), opentracing.TextMap, carrier)
		if err == nil {
			var pairs []string
			_ = carrier.ForeachKey(func(key, val string) error {
				pairs = append(pairs, key, val)
				return nil
			})

			ctx = metadata.AppendToOutgoingContext(ctx, pairs...)
		}

		err = invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			ext.LogError(span, err)
		}

		return err
	}
}

// OpenTracingServerInterceptor grpc unary serverinterceptor.
func OpenTracingServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		md, ok := metadata.FromIncomingContext(ctx)
		var spanCtx opentracing.SpanContext
		tracer := opentracing.GlobalTracer()
		if ok {
			carrier := make(opentracing.TextMapCarrier)
			for k, v := range md {
				carrier.Set(k, v[0])
			}

			spanCtx, _ = tracer.Extract(opentracing.TextMap, carrier)
		}

		span := tracer.StartSpan(info.FullMethod, ext.RPCServerOption(spanCtx))
		ext.Component.Set(span, "grpc")
		defer span.Finish()
		ctx = opentracing.ContextWithSpan(ctx, span)
		resp, err = handler(ctx, req)
		if err != nil {
			ext.LogError(span, err)
		}

		return
	}
}
