// Package metrics provides Prometheus metrics for OweHost
package metrics

import (
	"fmt"
	"net/http"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// Metrics provides application metrics
type Metrics struct {
	// HTTP metrics
	httpRequestsTotal   map[string]*uint64
	httpRequestDuration map[string]*durationMetric
	httpRequestsActive  int64

	// System metrics
	startTime time.Time

	// Business metrics
	usersTotal      int64
	domainsTotal    int64
	databasesTotal  int64
	backupsTotal    int64
	
	// Error metrics
	errorsTotal map[string]*uint64

	mu sync.RWMutex
}

type durationMetric struct {
	count uint64
	sum   float64
}

// NewMetrics creates a new metrics instance
func NewMetrics() *Metrics {
	return &Metrics{
		httpRequestsTotal:   make(map[string]*uint64),
		httpRequestDuration: make(map[string]*durationMetric),
		errorsTotal:         make(map[string]*uint64),
		startTime:           time.Now(),
	}
}

// RecordRequest records an HTTP request
func (m *Metrics) RecordRequest(method, path string, statusCode int, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s_%s_%d", method, normalizePath(path), statusCode)
	
	if m.httpRequestsTotal[key] == nil {
		var zero uint64
		m.httpRequestsTotal[key] = &zero
	}
	atomic.AddUint64(m.httpRequestsTotal[key], 1)

	if m.httpRequestDuration[key] == nil {
		m.httpRequestDuration[key] = &durationMetric{}
	}
	m.httpRequestDuration[key].count++
	m.httpRequestDuration[key].sum += duration.Seconds()
}

// RecordError records an error
func (m *Metrics) RecordError(errorType string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.errorsTotal[errorType] == nil {
		var zero uint64
		m.errorsTotal[errorType] = &zero
	}
	atomic.AddUint64(m.errorsTotal[errorType], 1)
}

// IncrementActive increments active requests
func (m *Metrics) IncrementActive() {
	atomic.AddInt64(&m.httpRequestsActive, 1)
}

// DecrementActive decrements active requests
func (m *Metrics) DecrementActive() {
	atomic.AddInt64(&m.httpRequestsActive, -1)
}

// SetUsersTotal sets the total users count
func (m *Metrics) SetUsersTotal(count int64) {
	atomic.StoreInt64(&m.usersTotal, count)
}

// SetDomainsTotal sets the total domains count
func (m *Metrics) SetDomainsTotal(count int64) {
	atomic.StoreInt64(&m.domainsTotal, count)
}

// SetDatabasesTotal sets the total databases count
func (m *Metrics) SetDatabasesTotal(count int64) {
	atomic.StoreInt64(&m.databasesTotal, count)
}

// SetBackupsTotal sets the total backups count
func (m *Metrics) SetBackupsTotal(count int64) {
	atomic.StoreInt64(&m.backupsTotal, count)
}

// Handler returns an HTTP handler for Prometheus metrics
func (m *Metrics) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		
		m.mu.RLock()
		defer m.mu.RUnlock()

		// Memory stats
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		// HTTP request metrics
		fmt.Fprintf(w, "# HELP owehost_http_requests_total Total number of HTTP requests\n")
		fmt.Fprintf(w, "# TYPE owehost_http_requests_total counter\n")
		
		keys := make([]string, 0, len(m.httpRequestsTotal))
		for k := range m.httpRequestsTotal {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		
		for _, key := range keys {
			count := atomic.LoadUint64(m.httpRequestsTotal[key])
			fmt.Fprintf(w, "owehost_http_requests_total{key=\"%s\"} %d\n", key, count)
		}

		// Active requests
		fmt.Fprintf(w, "\n# HELP owehost_http_requests_active Current number of active HTTP requests\n")
		fmt.Fprintf(w, "# TYPE owehost_http_requests_active gauge\n")
		fmt.Fprintf(w, "owehost_http_requests_active %d\n", atomic.LoadInt64(&m.httpRequestsActive))

		// Request duration
		fmt.Fprintf(w, "\n# HELP owehost_http_request_duration_seconds HTTP request duration in seconds\n")
		fmt.Fprintf(w, "# TYPE owehost_http_request_duration_seconds summary\n")
		for key, dm := range m.httpRequestDuration {
			if dm.count > 0 {
				avg := dm.sum / float64(dm.count)
				fmt.Fprintf(w, "owehost_http_request_duration_seconds{key=\"%s\",quantile=\"avg\"} %f\n", key, avg)
			}
		}

		// Business metrics
		fmt.Fprintf(w, "\n# HELP owehost_users_total Total number of users\n")
		fmt.Fprintf(w, "# TYPE owehost_users_total gauge\n")
		fmt.Fprintf(w, "owehost_users_total %d\n", atomic.LoadInt64(&m.usersTotal))

		fmt.Fprintf(w, "\n# HELP owehost_domains_total Total number of domains\n")
		fmt.Fprintf(w, "# TYPE owehost_domains_total gauge\n")
		fmt.Fprintf(w, "owehost_domains_total %d\n", atomic.LoadInt64(&m.domainsTotal))

		fmt.Fprintf(w, "\n# HELP owehost_databases_total Total number of databases\n")
		fmt.Fprintf(w, "# TYPE owehost_databases_total gauge\n")
		fmt.Fprintf(w, "owehost_databases_total %d\n", atomic.LoadInt64(&m.databasesTotal))

		fmt.Fprintf(w, "\n# HELP owehost_backups_total Total number of backups\n")
		fmt.Fprintf(w, "# TYPE owehost_backups_total gauge\n")
		fmt.Fprintf(w, "owehost_backups_total %d\n", atomic.LoadInt64(&m.backupsTotal))

		// Error metrics
		fmt.Fprintf(w, "\n# HELP owehost_errors_total Total number of errors\n")
		fmt.Fprintf(w, "# TYPE owehost_errors_total counter\n")
		for errorType, count := range m.errorsTotal {
			fmt.Fprintf(w, "owehost_errors_total{type=\"%s\"} %d\n", errorType, atomic.LoadUint64(count))
		}

		// System metrics
		fmt.Fprintf(w, "\n# HELP owehost_uptime_seconds Time since server start in seconds\n")
		fmt.Fprintf(w, "# TYPE owehost_uptime_seconds gauge\n")
		fmt.Fprintf(w, "owehost_uptime_seconds %f\n", time.Since(m.startTime).Seconds())

		fmt.Fprintf(w, "\n# HELP owehost_goroutines Number of goroutines\n")
		fmt.Fprintf(w, "# TYPE owehost_goroutines gauge\n")
		fmt.Fprintf(w, "owehost_goroutines %d\n", runtime.NumGoroutine())

		fmt.Fprintf(w, "\n# HELP owehost_memory_alloc_bytes Allocated memory in bytes\n")
		fmt.Fprintf(w, "# TYPE owehost_memory_alloc_bytes gauge\n")
		fmt.Fprintf(w, "owehost_memory_alloc_bytes %d\n", memStats.Alloc)

		fmt.Fprintf(w, "\n# HELP owehost_memory_sys_bytes Total memory from OS in bytes\n")
		fmt.Fprintf(w, "# TYPE owehost_memory_sys_bytes gauge\n")
		fmt.Fprintf(w, "owehost_memory_sys_bytes %d\n", memStats.Sys)

		fmt.Fprintf(w, "\n# HELP owehost_gc_runs_total Total number of GC runs\n")
		fmt.Fprintf(w, "# TYPE owehost_gc_runs_total counter\n")
		fmt.Fprintf(w, "owehost_gc_runs_total %d\n", memStats.NumGC)
	}
}

// MetricsMiddleware creates middleware that records request metrics
func MetricsMiddleware(m *Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			m.IncrementActive()
			defer m.DecrementActive()

			// Wrap response writer to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			
			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)
			m.RecordRequest(r.Method, r.URL.Path, wrapped.statusCode, duration)
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// normalizePath normalizes API paths for metrics aggregation
func normalizePath(path string) string {
	// Remove IDs from paths for aggregation
	// e.g., /api/v1/users/usr_123 -> /api/v1/users/:id
	parts := []byte(path)
	result := make([]byte, 0, len(parts))
	inID := false
	
	for i := 0; i < len(parts); i++ {
		if parts[i] == '/' {
			if inID {
				result = append(result, ":id"...)
				inID = false
			}
			result = append(result, parts[i])
		} else if !inID && len(result) > 0 && result[len(result)-1] == '/' {
			// Check if this segment looks like an ID
			if isIDChar(parts[i]) {
				inID = true
			} else {
				result = append(result, parts[i])
			}
		} else if !inID {
			result = append(result, parts[i])
		}
	}
	
	if inID {
		result = append(result, ":id"...)
	}
	
	return string(result)
}

func isIDChar(c byte) bool {
	return (c >= '0' && c <= '9') || c == '_'
}
