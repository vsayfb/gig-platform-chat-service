package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/vsayfb/gig-platform-chat-service/pkg/metrics"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Use chi's WrapResponseWriter to prevent hijacking error
		ww := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(ww, r)

		route := chi.RouteContext(r.Context()).RoutePattern()
		if route == "" {
			route = "unknown"
		}

		if route == "/metrics" {
			return
		}

		status := strconv.Itoa(ww.Status())
		duration := time.Since(start).Seconds()

		metrics.HttpRequestsTotal.
			WithLabelValues(route, r.Method, status).
			Inc()

		metrics.HttpRequestDuration.
			WithLabelValues(route, r.Method, status).
			Observe(duration)
	})
}
