package router

import (
	"bytes"
	"io"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// responseBodyWriter is a wrapper around gin.ResponseWriter to capture the response body
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w responseBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// TraceLogger is a middleware that logs request and response details to a separate span
func TraceLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		parentSpan := trace.SpanFromContext(c.Request.Context())
		if !parentSpan.IsRecording() {
			c.Next()
			return
		}

		// Create a child span for Request/Response logging
		tracer := parentSpan.TracerProvider().Tracer("auth-server-middleware")
		ctx, childSpan := tracer.Start(c.Request.Context(), "HTTP I/O", trace.WithSpanKind(trace.SpanKindInternal))
		defer childSpan.End()

		// --- Log Request ---
		var reqBody []byte
		if c.Request.Body != nil {
			reqBody, _ = io.ReadAll(c.Request.Body)
			// Restore the request body for downstream handlers
			c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBody))
		}

		childSpan.SetAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.String()),
			attribute.String("http.request.body", string(reqBody)),
			attribute.String("http.user_agent", c.Request.UserAgent()),
			attribute.String("http.client_ip", c.ClientIP()),
		)

		// --- Wrap Response Writer ---
		w := &responseBodyWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = w

		// Update context with the child span (optional, but good for consistency)
		c.Request = c.Request.WithContext(ctx)

		// Process request
		c.Next()

		// --- Log Response ---
		childSpan.SetAttributes(
			attribute.Int("http.status_code", c.Writer.Status()),
			attribute.String("http.response.body", w.body.String()),
		)
	}
}
