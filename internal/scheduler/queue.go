package scheduler

import (
	"log"
	"sync"

	internal "github.com/taskflow/taskflow/internal"
	"github.com/taskflow/taskflow/internal/store"
)

// QueueItem represents a job to be executed, with an optional pre-created run.
type QueueItem struct {
	Job *store.Job
	Run *store.Run // Optional: if set, use this run instead of creating a new one
}

// JobQueue runs independent jobs in parallel, serializing runs of the same job.
type JobQueue struct {
	items   []*QueueItem
	active  map[string]bool
	workers int
	running bool
	stopped bool
	mu      sync.Mutex
	changed *sync.Cond
}

// NewJobQueue creates a bounded queue with the requested number of workers.
func NewJobQueue(workers int) *JobQueue {
	if workers < 1 {
		workers = 1
	}
	jq := &JobQueue{workers: workers, active: make(map[string]bool)}
	jq.changed = sync.NewCond(&jq.mu)
	return jq
}

// Enqueue adds a job to the queue (creates new run during execution).
func (jq *JobQueue) Enqueue(job *store.Job) {
	jq.EnqueueWithRun(job, nil)
}

// EnqueueWithRun adds a job with an optional pre-created run to the queue.
func (jq *JobQueue) EnqueueWithRun(job *store.Job, run *store.Run) {
	if job == nil {
		return
	}
	jq.mu.Lock()
	defer jq.mu.Unlock()
	for len(jq.items) >= internal.JobQueueChannelSize && !jq.stopped {
		jq.changed.Wait()
	}
	if jq.stopped {
		return
	}
	// Workers own their copies; HTTP handlers can safely serialize the original run.
	jobCopy := *job
	var runCopy *store.Run
	if run != nil {
		copy := *run
		runCopy = &copy
	}
	jq.items = append(jq.items, &QueueItem{Job: &jobCopy, Run: runCopy})
	jq.changed.Broadcast()
}

// Start begins processing queued jobs. Calling Start again has no effect.
func (jq *JobQueue) Start(handler func(*store.Job, *store.Run) error) {
	jq.mu.Lock()
	defer jq.mu.Unlock()
	if jq.running || jq.stopped {
		return
	}
	jq.running = true
	for i := 0; i < jq.workers; i++ {
		go jq.work(handler)
	}
}

func (jq *JobQueue) work(handler func(*store.Job, *store.Run) error) {
	for {
		item := jq.next()
		if item == nil {
			return
		}
		if err := handler(item.Job, item.Run); err != nil {
			log.Printf("Error handling job %s: %v\n", item.Job.ID, err)
		}
		jq.mu.Lock()
		delete(jq.active, item.Job.ID)
		jq.changed.Broadcast()
		jq.mu.Unlock()
	}
}

// next reserves the oldest runnable item without blocking unrelated jobs behind
// another run of an active job. The reservation lasts until its handler returns.
func (jq *JobQueue) next() *QueueItem {
	jq.mu.Lock()
	defer jq.mu.Unlock()
	for !jq.stopped {
		for i, item := range jq.items {
			if jq.active[item.Job.ID] {
				continue
			}
			jq.active[item.Job.ID] = true
			copy(jq.items[i:], jq.items[i+1:])
			jq.items[len(jq.items)-1] = nil
			jq.items = jq.items[:len(jq.items)-1]
			jq.changed.Broadcast()
			return item
		}
		jq.changed.Wait()
	}
	return nil
}

// Stop prevents new jobs from starting; handlers already running can finish.
func (jq *JobQueue) Stop() {
	jq.mu.Lock()
	defer jq.mu.Unlock()
	jq.running = false
	jq.stopped = true
	jq.changed.Broadcast()
}

// IsRunning returns true if the queue is running.
func (jq *JobQueue) IsRunning() bool {
	jq.mu.Lock()
	defer jq.mu.Unlock()
	return jq.running
}
