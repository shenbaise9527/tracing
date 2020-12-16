# tracing

## 功能
* 提供http.Handler函数,`HttpTracing`可以很方便集成到http中间件.
* 提供grpc.UnaryClientInterceptor函数,一元RPC客户端拦截器`OpenTracingClientInterceptor`
* 提供grpc.UnaryServerInterceptor函数,一元RPC服务端拦截器`OpenTracingServerInterceptor`.
* 函数`GetSpanFromContext`从`context`中获取`opentracing.Span`对象,主要适用于`http.Handler`和`grpc.UnaryServerInterceptor`的context,然后可以根据需要调用`SetTag`和`Log`来设置相关的信息.
* 函数`ChildOfSpanFromContext`从`context`中生成`ClildOf`的`opentracing.Span`对象,用于跟踪某些子操作过程.
* 函数`FollowsSpanFromContext`从`context`中生成`FollewsFrom`的`opentracing.Span`对象,用于跟踪某些操作过程.

## 例子
创建好Tracer对象之后,一定要调用`opentracing.SetGlobalTracer(tracer)`,把tracer对象注册到opentracing的GlobalTracer中.
### 连接zipkin
``` go
func main() {
	// 采用HTTP方式来传输.
	zipkinReporter := zipkinhttp.NewReporter("http://127.0.0.1:9411/api/v2/spans")
	// 本地端点.
	endpoint, err := zipkin.NewEndpoint("goods", "127.0.0.1:8888")
	if err != nil {
		fmt.Println(err)

		return
	}

	// 创建zipkin的tracer对象.
	nativeTracer, err := zipkin.NewTracer(zipkinReporter, zipkin.WithLocalEndpoint(endpoint))
	if err != nil {
		fmt.Println(err)

		return
	}

	// 包装为opentracing规范的tracer对象.
	tracer := zipkinot.Wrap(nativeTracer)

	// 注册为opentracing里的GlobalTracer对象.
	opentracing.SetGlobalTracer(tracer)

	// 业务代码.
}
```
### 连接jaeger
``` go
func main() {
	cfg := jaegercfg.Configuration{
		ServiceName: "goods",
		// 采样策略,这里使用Const,全部采样.
		Sampler: &jaegercfg.SamplerConfig{
			Type:  "const",
			Param: 1.0,
		},
		Reporter: &jaegercfg.ReporterConfig{
			BufferFlushInterval: time.Second,
			LocalAgentHostPort:  "192.168.20.153:6831", // 采用UDP协议连接Agent.
			//CollectorEndpoint:   "http://192.168.20.153:14268/api/traces", // 采用HTTP协议直连Collector
		},
	}

	// 根据配置生成Tracer对象,启用Span的内存池.
	tracer, closer, err := cfg.NewTracer(jaegercfg.PoolSpans(true))
	if err != nil {
		fmt.Println(err)

		return
	}

	// 调用Close,释放资源.
	defer closer.Close()

	// 注册为opentracing里的GlobalTracer对象.
	opentracing.SetGlobalTracer(tracer)

	// 业务代码.
}
```