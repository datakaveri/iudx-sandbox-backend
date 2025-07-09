package health

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"runtime"
	"time"

	"github.com/iudx-sandbox-backend/pkg/logger"
)

// HealthStatus represents the overall health status
type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusUnhealthy HealthStatus = "unhealthy"
	StatusDegraded  HealthStatus = "degraded"
)

// HealthCheck represents a single health check
type HealthCheck struct {
	Name      string                 `json:"name"`
	Status    HealthStatus           `json:"status"`
	Message   string                 `json:"message,omitempty"`
	Duration  string                 `json:"duration"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// HealthResponse represents the overall health response
type HealthResponse struct {
	Status    HealthStatus           `json:"status"`
	Version   string                 `json:"version"`
	Timestamp time.Time              `json:"timestamp"`
	Uptime    string                 `json:"uptime"`
	Checks    map[string]HealthCheck `json:"checks"`
	System    SystemInfo             `json:"system"`
}

// SystemInfo represents system information
type SystemInfo struct {
	GoVersion     string `json:"go_version"`
	NumGoroutines int    `json:"num_goroutines"`
	NumCPU        int    `json:"num_cpu"`
	MemoryAlloc   uint64 `json:"memory_alloc_bytes"`
	MemorySys     uint64 `json:"memory_sys_bytes"`
	MemoryGC      uint32 `json:"memory_gc_cycles"`
}

// Checker defines the interface for health checkers
type Checker interface {
	Check(ctx context.Context) HealthCheck
}

// TaskQueueHealthChecker interface for task queue health checking
type TaskQueueHealthChecker interface {
	Health() error
	GetStatsInterface() interface{}
	IsStarted() bool
}

// HealthManager manages health checks
type HealthManager struct {
	checkers  map[string]Checker
	startTime time.Time
	version   string
}

// NewHealthManager creates a new health manager
func NewHealthManager(version string) *HealthManager {
	return &HealthManager{
		checkers:  make(map[string]Checker),
		startTime: time.Now(),
		version:   version,
	}
}

// RegisterChecker registers a health checker
func (hm *HealthManager) RegisterChecker(name string, checker Checker) {
	hm.checkers[name] = checker
}

// RunChecks runs all registered health checks
func (hm *HealthManager) RunChecks(ctx context.Context) HealthResponse {
	checks := make(map[string]HealthCheck)
	overallStatus := StatusHealthy

	// Run all checks concurrently
	checkChan := make(chan struct {
		name  string
		check HealthCheck
	}, len(hm.checkers))

	for name, checker := range hm.checkers {
		go func(n string, c Checker) {
			checkCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			result := c.Check(checkCtx)
			checkChan <- struct {
				name  string
				check HealthCheck
			}{name: n, check: result}
		}(name, checker)
	}

	// Collect results
	for i := 0; i < len(hm.checkers); i++ {
		result := <-checkChan
		checks[result.name] = result.check

		// Determine overall status
		switch result.check.Status {
		case StatusUnhealthy:
			overallStatus = StatusUnhealthy
		case StatusDegraded:
			if overallStatus == StatusHealthy {
				overallStatus = StatusDegraded
			}
		}
	}

	return HealthResponse{
		Status:    overallStatus,
		Version:   hm.version,
		Timestamp: time.Now(),
		Uptime:    time.Since(hm.startTime).String(),
		Checks:    checks,
		System:    getSystemInfo(),
	}
}

// getSystemInfo collects system information
func getSystemInfo() SystemInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return SystemInfo{
		GoVersion:     runtime.Version(),
		NumGoroutines: runtime.NumGoroutine(),
		NumCPU:        runtime.NumCPU(),
		MemoryAlloc:   m.Alloc,
		MemorySys:     m.Sys,
		MemoryGC:      m.NumGC,
	}
}

// DatabaseChecker checks database connectivity
type DatabaseChecker struct {
	db *sql.DB
}

// NewDatabaseChecker creates a new database checker
func NewDatabaseChecker(db *sql.DB) *DatabaseChecker {
	return &DatabaseChecker{db: db}
}

// Check performs database health check
func (dc *DatabaseChecker) Check(ctx context.Context) HealthCheck {
	start := time.Now()

	check := HealthCheck{
		Name:      "database",
		Timestamp: start,
		Metadata:  make(map[string]interface{}),
	}

	// Test database connection
	if err := dc.db.PingContext(ctx); err != nil {
		check.Status = StatusUnhealthy
		check.Message = "Database connection failed: " + err.Error()
		check.Duration = time.Since(start).String()
		return check
	}

	// Test simple query
	var count int
	if err := dc.db.QueryRowContext(ctx, "SELECT 1").Scan(&count); err != nil {
		check.Status = StatusDegraded
		check.Message = "Database query test failed: " + err.Error()
		check.Duration = time.Since(start).String()
		return check
	}

	// Check connection stats
	stats := dc.db.Stats()
	check.Metadata["open_connections"] = stats.OpenConnections
	check.Metadata["in_use"] = stats.InUse
	check.Metadata["idle"] = stats.Idle
	check.Metadata["wait_count"] = stats.WaitCount
	check.Metadata["wait_duration"] = stats.WaitDuration.String()

	check.Status = StatusHealthy
	check.Message = "Database is healthy"
	check.Duration = time.Since(start).String()

	return check
}

// TaskQueueChecker checks task queue health
type TaskQueueChecker struct {
	taskQueue TaskQueueHealthChecker
}

// NewTaskQueueChecker creates a new task queue checker
func NewTaskQueueChecker(tq TaskQueueHealthChecker) *TaskQueueChecker {
	return &TaskQueueChecker{taskQueue: tq}
}

// Check performs task queue health check
func (tqc *TaskQueueChecker) Check(ctx context.Context) HealthCheck {
	start := time.Now()

	check := HealthCheck{
		Name:      "task_queue",
		Timestamp: start,
		Metadata:  make(map[string]interface{}),
	}

	// Check if task queue is started
	if !tqc.taskQueue.IsStarted() {
		check.Status = StatusUnhealthy
		check.Message = "Task queue is not started"
		check.Duration = time.Since(start).String()
		return check
	}

	// Check task queue health
	if err := tqc.taskQueue.Health(); err != nil {
		check.Status = StatusDegraded
		check.Message = "Task queue health check failed: " + err.Error()
		check.Duration = time.Since(start).String()
		return check
	}

	// Add task queue statistics
	stats := tqc.taskQueue.GetStatsInterface()
	check.Metadata["stats"] = stats

	check.Status = StatusHealthy
	check.Message = "Task queue is healthy"
	check.Duration = time.Since(start).String()

	return check
}

// ReadinessChecker checks if application is ready to serve traffic
type ReadinessChecker struct {
	dependencies []Checker
}

// NewReadinessChecker creates a new readiness checker
func NewReadinessChecker(dependencies ...Checker) *ReadinessChecker {
	return &ReadinessChecker{dependencies: dependencies}
}

// Check performs readiness check
func (rc *ReadinessChecker) Check(ctx context.Context) HealthCheck {
	start := time.Now()

	check := HealthCheck{
		Name:      "readiness",
		Timestamp: start,
	}

	// Check all dependencies
	for _, dep := range rc.dependencies {
		depCheck := dep.Check(ctx)
		if depCheck.Status == StatusUnhealthy {
			check.Status = StatusUnhealthy
			check.Message = "Dependency " + depCheck.Name + " is unhealthy"
			check.Duration = time.Since(start).String()
			return check
		}
	}

	check.Status = StatusHealthy
	check.Message = "Application is ready"
	check.Duration = time.Since(start).String()

	return check
}

// LivenessChecker checks if application is alive
type LivenessChecker struct{}

// NewLivenessChecker creates a new liveness checker
func NewLivenessChecker() *LivenessChecker {
	return &LivenessChecker{}
}

// Check performs liveness check
func (lc *LivenessChecker) Check(ctx context.Context) HealthCheck {
	start := time.Now()

	return HealthCheck{
		Name:      "liveness",
		Status:    StatusHealthy,
		Message:   "Application is alive",
		Duration:  time.Since(start).String(),
		Timestamp: start,
	}
}

// HTTPHandler creates HTTP handlers for health endpoints
func (hm *HealthManager) HTTPHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		health := hm.RunChecks(ctx)

		w.Header().Set("Content-Type", "application/json")

		// Set appropriate HTTP status code
		switch health.Status {
		case StatusHealthy:
			w.WriteHeader(http.StatusOK)
		case StatusDegraded:
			w.WriteHeader(http.StatusOK) // Still OK but with warnings
		case StatusUnhealthy:
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		if err := json.NewEncoder(w).Encode(health); err != nil {
			logger.Error.Printf("Failed to encode health response: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}

// ReadinessHandler creates a readiness endpoint
func (hm *HealthManager) ReadinessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Only check readiness-related checkers
		readinessChecker := NewReadinessChecker()
		for _, checker := range hm.checkers {
			if _, ok := checker.(*DatabaseChecker); ok {
				readinessChecker.dependencies = append(readinessChecker.dependencies, checker)
			}
			if _, ok := checker.(*TaskQueueChecker); ok {
				readinessChecker.dependencies = append(readinessChecker.dependencies, checker)
			}
		}

		check := readinessChecker.Check(ctx)

		w.Header().Set("Content-Type", "application/json")

		if check.Status == StatusHealthy {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		if err := json.NewEncoder(w).Encode(check); err != nil {
			logger.Error.Printf("Failed to encode readiness response: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}

// LivenessHandler creates a liveness endpoint
func (hm *HealthManager) LivenessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		livenessChecker := NewLivenessChecker()
		check := livenessChecker.Check(ctx)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(check); err != nil {
			logger.Error.Printf("Failed to encode liveness response: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}
