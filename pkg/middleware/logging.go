package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/iudx-sandbox-backend/pkg/logger"
	"github.com/julienschmidt/httprouter"
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    int64
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = 200
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.written += int64(n)
	return n, err
}

func LogRequest(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		start := time.Now()

		// Generate request ID if not present
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Add request ID to context
		ctx := context.WithValue(r.Context(), "request_id", requestID)
		r = r.WithContext(ctx)

		// Add request ID to response headers
		w.Header().Set("X-Request-ID", requestID)

		// Wrap response writer to capture status and size
		wrapped := &responseWriter{ResponseWriter: w, statusCode: 0}

		// Log request start
		logger.InfoWithMetadata("HTTP request started", map[string]interface{}{
			"method":         r.Method,
			"path":           r.URL.Path,
			"query":          r.URL.RawQuery,
			"remote_addr":    r.RemoteAddr,
			"user_agent":     r.UserAgent(),
			"request_id":     requestID,
			"content_length": r.ContentLength,
			"host":           r.Host,
		})

		// Log debug information about headers (excluding sensitive ones)
		debugHeaders := make(map[string]string)
		for name, values := range r.Header {
			// Skip sensitive headers
			if name == "Authorization" || name == "Cookie" || name == "X-API-Key" {
				debugHeaders[name] = "[REDACTED]"
			} else if len(values) > 0 {
				debugHeaders[name] = values[0]
			}
		}

		logger.DebugWithMetadata("Request headers", map[string]interface{}{
			"request_id": requestID,
			"headers":    debugHeaders,
		})

		// Call next handler
		next(wrapped, r, p)

		// Calculate duration
		duration := time.Since(start)

		// Determine log level based on status code
		metadata := map[string]interface{}{
			"method":        r.Method,
			"path":          r.URL.Path,
			"status_code":   wrapped.statusCode,
			"duration_ms":   duration.Milliseconds(),
			"duration":      duration.String(),
			"response_size": wrapped.written,
			"request_id":    requestID,
			"remote_addr":   r.RemoteAddr,
		}

		// Log based on status code and performance
		switch {
		case wrapped.statusCode >= 500:
			logger.ErrorWithMetadata("HTTP request completed with server error", metadata, nil)
		case wrapped.statusCode >= 400:
			logger.WarnWithMetadata("HTTP request completed with client error", metadata)
		case duration > 5*time.Second:
			logger.WarnWithMetadata("HTTP request completed (slow response)", metadata)
		case duration > 1*time.Second:
			logger.InfoWithMetadata("HTTP request completed", metadata)
		default:
			logger.DebugWithMetadata("HTTP request completed", metadata)
		}

		// Log performance warnings
		if duration > 10*time.Second {
			logger.WarnWithMetadata("Very slow HTTP request detected", map[string]interface{}{
				"request_id": requestID,
				"path":       r.URL.Path,
				"duration":   duration.String(),
				"threshold":  "10s",
			})
		}
	}
}
