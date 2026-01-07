package scheduler

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

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
		ticker:  time.NewTicker(time.Minute),
		done:    make(chan struct{}),
	}
}

// Start begins the scheduling loop
func (s *Scheduler) Start(ctx context.Context, handler func(*store.Job) error) error {
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

		// Get the job's schedule
		schedule, err := s.store.GetJobSchedule(job.ID)
		if err != nil {
			log.Printf("Failed to get schedule for job %s: %v\n", job.ID, err)
			continue
		}

		// Check if job should run now
		if s.matcher.Matches(now, schedule) {
			// Get the last run to ensure we don't run multiple times in the same minute
			runs, err := s.store.ListRuns(&job.ID, 1, 0)
			if err == nil && len(runs) > 0 {
				lastRun := runs[0]
				if lastRun.StartedAt != nil && lastRun.StartedAt.Year() == now.Year() &&
					lastRun.StartedAt.Month() == now.Month() &&
					lastRun.StartedAt.Day() == now.Day() &&
					lastRun.StartedAt.Hour() == now.Hour() &&
					lastRun.StartedAt.Minute() == now.Minute() {
					// Already ran this minute
					continue
				}
			}

			// Enqueue job for execution
			s.queue.Enqueue(job)
		}
	}
}

// IsRunning returns true if scheduler is running
func (s *Scheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// Enqueue adds a job to the execution queue (for manual triggers)
func (s *Scheduler) Enqueue(job *store.Job) {
	s.queue.Enqueue(job)
}
