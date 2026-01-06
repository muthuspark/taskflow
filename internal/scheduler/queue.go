package scheduler

import (
	"log"
	"sync"

	"github.com/taskflow/taskflow/internal/store"
)

// JobQueue manages sequential job execution
type JobQueue struct {
	jobs    chan *store.Job
	running bool
	mu      sync.RWMutex
	done    chan struct{}
}

// NewJobQueue creates a new job queue
func NewJobQueue() *JobQueue {
	return &JobQueue{
		jobs: make(chan *store.Job, 100),
		done: make(chan struct{}),
	}
}

// Enqueue adds a job to the queue
func (jq *JobQueue) Enqueue(job *store.Job) {
	jq.jobs <- job
}

// Start begins processing queued jobs
func (jq *JobQueue) Start(handler func(*store.Job) error) {
	jq.mu.Lock()
	jq.running = true
	jq.mu.Unlock()

	go func() {
		for {
			select {
			case job := <-jq.jobs:
				if job != nil {
					if err := handler(job); err != nil {
						log.Printf("Error handling job %s: %v\n", job.ID, err)
					}
				}
			case <-jq.done:
				return
			}
		}
	}()
}

// Stop stops the queue
func (jq *JobQueue) Stop() {
	jq.mu.Lock()
	jq.running = false
	jq.mu.Unlock()
	close(jq.done)
}

// IsRunning returns true if the queue is running
func (jq *JobQueue) IsRunning() bool {
	jq.mu.RLock()
	defer jq.mu.RUnlock()
	return jq.running
}
