package tracing

import (
	"net/http"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
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

// OpenTracingHandler http.Handler.
func OpenTracingHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tracer := opentracing.GlobalTracer()
		spanCtx, _ := tracer.Extract(
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(r.Header))
		span := tracer.StartSpan(r.RequestURI, opentracing.ChildOf(spanCtx))
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
