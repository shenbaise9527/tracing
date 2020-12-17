package tracing

import (
	"context"

	"github.com/opentracing/opentracing-go"
)

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
