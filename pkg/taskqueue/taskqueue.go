package taskqueue

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/iudx-sandbox-backend/pkg/logger"
)

// TaskQueue implements a simple in-memory task queue
type TaskQueue struct {
	tasks   chan *Task
	workers int
	stats   *Stats
	mu      sync.RWMutex
	wg      sync.WaitGroup
	stop    chan struct{}
	started bool
}

// Stats represents task queue statistics
type Stats struct {
	Processed   int64     `json:"processed"`
	Failed      int64     `json:"failed"`
	Active      int64     `json:"active"`
	Queued      int64     `json:"queued"`
	LastUpdated time.Time `json:"last_updated"`
}

// Task represents a task to be executed
type Task struct {
	ID      string
	Type    string
	Payload interface{}
	Handler TaskHandler
	Created time.Time
}

// TaskHandler defines the function signature for task handlers
type TaskHandler func(context.Context, interface{}) error

// InitializeTaskQueue creates a new in-memory task queue with default settings
func InitializeTaskQueue() *TaskQueue {
	logger.InfoWithMetadata("Initializing in-memory task queue with default settings", map[string]interface{}{
		"capacity": 1000,
		"workers":  10,
	})
	return NewTaskQueue(1000, 10) // 1000 task capacity, 10 workers
}

// NewTaskQueue creates a new task queue with specified capacity and workers
func NewTaskQueue(capacity, workers int) *TaskQueue {
	logger.InfoWithMetadata("Creating new task queue", map[string]interface{}{
		"capacity": capacity,
		"workers":  workers,
	})

	return &TaskQueue{
		tasks:   make(chan *Task, capacity),
		workers: workers,
		stats: &Stats{
			LastUpdated: time.Now(),
		},
		stop: make(chan struct{}),
	}
}

// Start starts the task queue workers
func (tq *TaskQueue) Start(ctx context.Context) error {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	if tq.started {
		logger.WarnWithMetadata("Attempted to start already running task queue", map[string]interface{}{
			"workers": tq.workers,
		})
		return fmt.Errorf("task queue already started")
	}

	tq.started = true

	logger.InfoWithMetadata("Starting task queue workers", map[string]interface{}{
		"worker_count":   tq.workers,
		"queue_capacity": cap(tq.tasks),
		"startup_time":   time.Now().Format(time.RFC3339),
	})

	// Start worker goroutines
	for i := 0; i < tq.workers; i++ {
		tq.wg.Add(1)
		go tq.worker(ctx, i)

		logger.DebugWithMetadata("Started task queue worker", map[string]interface{}{
			"worker_id": i,
		})
	}

	logger.InfoWithMetadata("Task queue started successfully", map[string]interface{}{
		"workers_started": tq.workers,
		"queue_ready":     true,
	})

	return nil
}

// Stop gracefully stops the task queue
func (tq *TaskQueue) Stop() error {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	if !tq.started {
		logger.DebugWithMetadata("Task queue stop requested but not running", map[string]interface{}{
			"current_state": "stopped",
		})
		return nil
	}

	logger.InfoWithMetadata("Stopping task queue gracefully", map[string]interface{}{
		"active_workers": tq.workers,
		"queued_tasks":   len(tq.tasks),
		"shutdown_time":  time.Now().Format(time.RFC3339),
	})

	close(tq.stop)
	close(tq.tasks)

	// Wait for all workers to finish with timeout
	done := make(chan struct{})
	go func() {
		tq.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.InfoWithMetadata("All task queue workers stopped gracefully", map[string]interface{}{
			"shutdown_duration": "< 30s",
		})
	case <-time.After(30 * time.Second):
		logger.WarnWithMetadata("Task queue shutdown timeout reached", map[string]interface{}{
			"timeout": "30s",
			"status":  "forced_shutdown",
		})
	}

	tq.started = false
	logger.InfoWithMetadata("Task queue stopped successfully", map[string]interface{}{
		"final_stats": tq.stats,
	})

	return nil
}

// AddTask adds a task to the queue
func (tq *TaskQueue) AddTask(ctx context.Context, taskType string, payload interface{}, handler TaskHandler) error {
	if handler == nil {
		logger.ErrorWithMetadata("Task handler cannot be nil", map[string]interface{}{
			"task_type": taskType,
		}, nil)
		return fmt.Errorf("task handler cannot be nil")
	}

	task := &Task{
		ID:      fmt.Sprintf("%s-%d", taskType, time.Now().UnixNano()),
		Type:    taskType,
		Payload: payload,
		Handler: handler,
		Created: time.Now(),
	}

	logger.DebugWithMetadata("Attempting to add task to queue", map[string]interface{}{
		"task_id":        task.ID,
		"task_type":      taskType,
		"queue_size":     len(tq.tasks),
		"queue_capacity": cap(tq.tasks),
	})

	select {
	case tq.tasks <- task:
		tq.updateStats(func(s *Stats) {
			s.Queued++
			s.LastUpdated = time.Now()
		})

		logger.InfoWithMetadata("Task added to queue successfully", map[string]interface{}{
			"task_id":      task.ID,
			"task_type":    taskType,
			"queue_size":   len(tq.tasks),
			"queued_count": tq.stats.Queued,
		})

		return nil

	case <-ctx.Done():
		logger.WarnWithMetadata("Task addition cancelled due to context", map[string]interface{}{
			"task_id":   task.ID,
			"task_type": taskType,
			"reason":    ctx.Err().Error(),
		})
		return ctx.Err()

	default:
		tq.updateStats(func(s *Stats) {
			s.Failed++
			s.LastUpdated = time.Now()
		})

		logger.ErrorWithMetadata("Task queue is full, task rejected", map[string]interface{}{
			"task_id":        task.ID,
			"task_type":      taskType,
			"queue_capacity": cap(tq.tasks),
			"queue_size":     len(tq.tasks),
		}, nil)

		return fmt.Errorf("task queue is full, task rejected")
	}
}

// worker processes tasks from the queue
func (tq *TaskQueue) worker(ctx context.Context, workerID int) {
	defer tq.wg.Done()

	logger.InfoWithMetadata("Task queue worker started", map[string]interface{}{
		"worker_id":  workerID,
		"start_time": time.Now().Format(time.RFC3339),
	})

	taskCount := 0
	for {
		select {
		case task, ok := <-tq.tasks:
			if !ok {
				logger.InfoWithMetadata("Task queue worker stopping - channel closed", map[string]interface{}{
					"worker_id":       workerID,
					"tasks_processed": taskCount,
				})
				return
			}

			taskCount++
			logger.DebugWithMetadata("Worker received task", map[string]interface{}{
				"worker_id": workerID,
				"task_id":   task.ID,
				"task_type": task.Type,
				"wait_time": time.Since(task.Created).String(),
			})

			tq.updateStats(func(s *Stats) {
				s.Active++
				s.Queued--
				s.LastUpdated = time.Now()
			})

			tq.executeTask(ctx, task, workerID)

		case <-tq.stop:
			logger.InfoWithMetadata("Task queue worker stopping - stop signal received", map[string]interface{}{
				"worker_id":       workerID,
				"tasks_processed": taskCount,
			})
			return

		case <-ctx.Done():
			logger.InfoWithMetadata("Task queue worker stopping - context cancelled", map[string]interface{}{
				"worker_id":       workerID,
				"tasks_processed": taskCount,
				"reason":          ctx.Err().Error(),
			})
			return
		}
	}
}

// executeTask executes a single task
func (tq *TaskQueue) executeTask(ctx context.Context, task *Task, workerID int) {
	start := time.Now()

	defer func() {
		if r := recover(); r != nil {
			duration := time.Since(start)

			logger.ErrorWithMetadata("Task execution panicked", map[string]interface{}{
				"worker_id":   workerID,
				"task_id":     task.ID,
				"task_type":   task.Type,
				"panic":       r,
				"duration":    duration.String(),
				"stack_trace": true,
			}, fmt.Errorf("panic: %v", r))

			tq.updateStats(func(s *Stats) {
				s.Failed++
				s.Active--
				s.LastUpdated = time.Now()
			})
		}
	}()

	logger.InfoWithMetadata("Starting task execution", map[string]interface{}{
		"worker_id": workerID,
		"task_id":   task.ID,
		"task_type": task.Type,
		"age":       time.Since(task.Created).String(),
	})

	// Create a timeout context for the task
	taskCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	err := task.Handler(taskCtx, task.Payload)
	duration := time.Since(start)

	tq.updateStats(func(s *Stats) {
		if err != nil {
			s.Failed++
			logger.ErrorWithMetadata("Task execution failed", map[string]interface{}{
				"worker_id": workerID,
				"task_id":   task.ID,
				"task_type": task.Type,
				"duration":  duration.String(),
				"error":     err.Error(),
			}, err)
		} else {
			s.Processed++

			// Log based on execution time
			metadata := map[string]interface{}{
				"worker_id":   workerID,
				"task_id":     task.ID,
				"task_type":   task.Type,
				"duration":    duration.String(),
				"duration_ms": duration.Milliseconds(),
			}

			if duration > 2*time.Minute {
				logger.WarnWithMetadata("Task completed but was very slow", metadata)
			} else if duration > 30*time.Second {
				logger.InfoWithMetadata("Task completed (slow)", metadata)
			} else {
				logger.DebugWithMetadata("Task completed successfully", metadata)
			}
		}
		s.Active--
		s.LastUpdated = time.Now()
	})
}

// updateStats safely updates the statistics
func (tq *TaskQueue) updateStats(updateFunc func(*Stats)) {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	updateFunc(tq.stats)
}

// GetStats returns current queue statistics
func (tq *TaskQueue) GetStats() Stats {
	tq.mu.RLock()
	defer tq.mu.RUnlock()
	return *tq.stats
}

// GetStatsInterface returns current queue statistics as interface{} for health checker
func (tq *TaskQueue) GetStatsInterface() interface{} {
	return tq.GetStats()
}

// Health checks if the task queue is healthy
func (tq *TaskQueue) Health() error {
	stats := tq.GetStats()

	// Check if queue has been stalled
	if time.Since(stats.LastUpdated) > 10*time.Minute {
		logger.ErrorWithMetadata("Task queue health check failed - appears stalled", map[string]interface{}{
			"last_updated":     stats.LastUpdated.Format(time.RFC3339),
			"stalled_duration": time.Since(stats.LastUpdated).String(),
		}, nil)
		return fmt.Errorf("task queue appears to be stalled")
	}

	// Check for high failure rate
	if stats.Processed > 0 {
		failureRate := float64(stats.Failed) / float64(stats.Processed+stats.Failed) * 100
		if failureRate > 50 {
			logger.WarnWithMetadata("High task failure rate detected", map[string]interface{}{
				"failure_rate":    fmt.Sprintf("%.2f%%", failureRate),
				"failed_tasks":    stats.Failed,
				"processed_tasks": stats.Processed,
			})
		}
	}

	logger.DebugWithMetadata("Task queue health check passed", map[string]interface{}{
		"stats": stats,
	})

	return nil
}

// QueueSize returns the current number of queued tasks
func (tq *TaskQueue) QueueSize() int {
	size := len(tq.tasks)

	if size > cap(tq.tasks)*3/4 {
		logger.WarnWithMetadata("Task queue is becoming full", map[string]interface{}{
			"current_size":  size,
			"capacity":      cap(tq.tasks),
			"usage_percent": float64(size) / float64(cap(tq.tasks)) * 100,
		})
	}

	return size
}

// IsStarted returns whether the task queue is started
func (tq *TaskQueue) IsStarted() bool {
	tq.mu.RLock()
	defer tq.mu.RUnlock()
	return tq.started
}
