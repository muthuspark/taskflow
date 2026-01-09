package scheduler

import (
	"log"
	"sync"

	internal "github.com/taskflow/taskflow/internal"
	"github.com/taskflow/taskflow/internal/store"
)

// QueueItem represents a job to be executed, with an optional pre-created run
type QueueItem struct {
	Job *store.Job
	Run *store.Run // Optional: if set, use this run instead of creating a new one
}

// JobQueue manages sequential job execution
type JobQueue struct {
	items   chan *QueueItem
	running bool
	mu      sync.RWMutex
	done    chan struct{}
}

// NewJobQueue creates a new job queue
func NewJobQueue() *JobQueue {
	return &JobQueue{
		items: make(chan *QueueItem, internal.JobQueueChannelSize),
		done:  make(chan struct{}),
	}
}

// Enqueue adds a job to the queue (creates new run during execution)
func (jq *JobQueue) Enqueue(job *store.Job) {
	jq.items <- &QueueItem{Job: job, Run: nil}
}

// EnqueueWithRun adds a job with a pre-created run to the queue
func (jq *JobQueue) EnqueueWithRun(job *store.Job, run *store.Run) {
	jq.items <- &QueueItem{Job: job, Run: run}
}

// Start begins processing queued jobs
func (jq *JobQueue) Start(handler func(*store.Job, *store.Run) error) {
	jq.mu.Lock()
	jq.running = true
	jq.mu.Unlock()

	go func() {
		for {
			select {
			case item := <-jq.items:
				if item != nil && item.Job != nil {
					if err := handler(item.Job, item.Run); err != nil {
						log.Printf("Error handling job %s: %v\n", item.Job.ID, err)
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
