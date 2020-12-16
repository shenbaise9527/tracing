package tracing

import (
	"context"
	"net/http"

	"github.com/opentracing/opentracing-go/ext"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type withHTTPCodeResponse struct {
	writer http.ResponseWriter
	code   int
}

func (w *withHTTPCodeResponse) Header() http.Header {
	return w.writer.Header()
}

func (w *withHTTPCodeResponse) Write(bytes []byte) (int, error) {
	return w.writer.Write(bytes)
}

func (w *withHTTPCodeResponse) WriteHeader(code int) {
	w.writer.WriteHeader(code)
	w.code = code
}

// HttpTracing http.Handler.
func HttpTracing(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tracer := opentracing.GlobalTracer()
		carrier := opentracing.HTTPHeadersCarrier(r.Header)
		spanCtx, err := tracer.Extract(opentracing.HTTPHeaders, carrier)
		var span opentracing.Span
		if err == nil {
			span = tracer.StartSpan(r.RequestURI, opentracing.ChildOf(spanCtx))
		} else {
			span = tracer.StartSpan(r.RequestURI)
		}

		ext.HTTPMethod.Set(span, r.Method)
		defer span.Finish()
		cw := &withHTTPCodeResponse{writer: w}
		rc := opentracing.ContextWithSpan(r.Context(), span)
		r = r.WithContext(rc)
		defer func() {
			ext.HTTPStatusCode.Set(span, uint16(cw.code))
			if cw.code >= http.StatusBadRequest {
				ext.Error.Set(span, true)
			}
		}()

		next(cw, r)
	}
}

// OpenTracingServerInterceptor grpc unary serverinterceptor.
func OpenTracingServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		md, ok := metadata.FromIncomingContext(ctx)
		var span opentracing.Span
		tracer := opentracing.GlobalTracer()
		if ok {
			carrier := make(opentracing.TextMapCarrier)
			for k, v := range md {
				carrier.Set(k, v[0])
			}

			if spanCtx, err := tracer.Extract(opentracing.TextMap, carrier); err == nil {
				span = tracer.StartSpan(info.FullMethod, opentracing.ChildOf(spanCtx))
			}
		}

		if span == nil {
			span = tracer.StartSpan(info.FullMethod)
		}

		ext.Component.Set(span, "grpc")
		ext.SpanKind.Set(span, ext.SpanKindRPCServerEnum)
		defer span.Finish()
		ctx = opentracing.ContextWithSpan(ctx, span)
		resp, err = handler(ctx, req)
		if err != nil {
			ext.LogError(span, err)
		}

		return
	}
}

// OpenTracingClientInterceptor grpc unary clientinterceptor.
func OpenTracingClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		span := ChildOfSpanFromContext(ctx, method)
		ext.Component.Set(span, "grpc")
		ext.SpanKind.Set(span, ext.SpanKindRPCClientEnum)
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
	span := opentracing.SpanFromContext(ctx)
	if span == nil {
		span = tracer.StartSpan(operationName)
	} else {
		span = tracer.StartSpan(operationName, op(span.Context()))
	}

	return span
}
