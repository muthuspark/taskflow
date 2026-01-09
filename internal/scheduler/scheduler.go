package scheduler

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	internal "github.com/taskflow/taskflow/internal"
	"github.com/taskflow/taskflow/internal/store"
)

// Scheduler manages job scheduling and execution
type Scheduler struct {
	store   *store.Store
	queue   *JobQueue
	matcher *Matcher
	ticker  *time.Ticker
	done    chan struct{}
	mu      sync.RWMutex
	running bool
}

// New creates a new scheduler
func New(st *store.Store) *Scheduler {
	return &Scheduler{
		store:   st,
		queue:   NewJobQueue(),
		matcher: NewMatcher(),
		ticker:  time.NewTicker(internal.SchedulerCheckInterval),
		done:    make(chan struct{}),
	}
}

// Start begins the scheduling loop
func (s *Scheduler) Start(ctx context.Context, handler func(*store.Job, *store.Run) error) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("scheduler is already running")
	}
	s.running = true
	s.mu.Unlock()

	s.queue.Start(handler)

	go s.run(ctx)

	log.Println("Scheduler started")
	return nil
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	close(s.done)
	s.ticker.Stop()
	s.queue.Stop()

	log.Println("Scheduler stopped")
}

// run executes the scheduling loop
func (s *Scheduler) run(ctx context.Context) {
	for {
		select {
		case <-s.ticker.C:
			s.checkAndScheduleJobs()
		case <-s.done:
			return
		case <-ctx.Done():
			return
		}
	}
}

// checkAndScheduleJobs checks all jobs and schedules those that should run
func (s *Scheduler) checkAndScheduleJobs() {
	jobs, err := s.store.ListJobs(nil)
	if err != nil {
		log.Printf("Failed to list jobs: %v\n", err)
		return
	}

	now := time.Now()

	for _, job := range jobs {
		if !job.Enabled {
			continue
		}

		schedule, err := s.store.GetJobSchedule(job.ID)
		if err != nil {
			log.Printf("Failed to get schedule for job %s: %v\n", job.ID, err)
			continue
		}

		if !s.matcher.Matches(now, schedule) {
			continue
		}

		if s.alreadyRanThisMinute(job.ID, now) {
			continue
		}

		s.queue.Enqueue(job)
	}
}

// alreadyRanThisMinute checks if a job has already run in the current minute
func (s *Scheduler) alreadyRanThisMinute(jobID string, now time.Time) bool {
	runs, err := s.store.ListRuns(&jobID, 1, 0)
	if err != nil || len(runs) == 0 {
		return false
	}

	lastRun := runs[0]
	if lastRun.StartedAt == nil {
		return false
	}

	// Use time.Truncate for cleaner minute-precision comparison
	return lastRun.StartedAt.Truncate(time.Minute).Equal(now.Truncate(time.Minute))
}

// IsRunning returns true if scheduler is running
func (s *Scheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// Enqueue adds a job to the execution queue (for scheduled triggers)
func (s *Scheduler) Enqueue(job *store.Job) {
	s.queue.Enqueue(job)
}

// EnqueueWithRun adds a job with a pre-created run to the queue (for manual triggers)
func (s *Scheduler) EnqueueWithRun(job *store.Job, run *store.Run) {
	s.queue.EnqueueWithRun(job, run)
}
