# tracing

## 功能
1. 集成zipkin和jaeger,根据`build -tags`来区分是连接zipkin还是连接jaeger,默认是连接到zipikin的,通过`go build -tags=jaeger`来连接到jaeger.
2. 提供http.Handler函数,`HttpTracing`可以很方便集成到http中间件.
3. 提供grpc.UnaryClientInterceptor函数,一元RPC客户端拦截器`OpenTracingClientInterceptor`
4. 提供grpc.UnaryServerInterceptor函数,一元RPC服务端拦截器`OpenTracingServerInterceptor`.
5. 函数`GetSpanFromContext`从`context`中获取`opentracing.Span`对象,主要适用于`http.Handler`和`grpc.UnaryServerInterceptor`的context,然后可以根据需要调用`SetTag`和`Log`来设置相关的信息.
6. 函数`ChildOfSpanFromContext`从`context`中生成`ClildOf`的`opentracing.Span`对象,用于跟踪某些子操作过程.
7. 函数`FollowsSpanFromContext`从`context`中生成`FollewsFrom`的`opentracing.Span`对象,用于跟踪某些操作过程.

## 例子
``` go
func main() {
    // 要在main函数中初始化opentracing,并要调用Close().
    var tracingURL, serverName, localEndpoint string
	closer, err := tracing.NewOpenTracer(tracingURL, serverName, localEndpoint)
	if err != nil {
		fmt.Printf("inin opentracing failed: %s", err)

		return
	}

    defer closer.Close()
}
```