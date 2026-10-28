package scheduler

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/taskflow/taskflow/internal/store"
)

func receive(t *testing.T, ch <-chan string) string {
	t.Helper()
	select {
	case value := <-ch:
		return value
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for worker")
		return ""
	}
}

// Duplicate jobs must not occupy every worker and block unrelated queued work.
func TestWorkersSerializeSameJobWithoutBlockingOtherJobs(t *testing.T) {
	q := NewJobQueue(2)
	started := make(chan string, 4)
	finished := make(chan string, 4)
	release := map[string]chan struct{}{}
	for _, id := range []string{"a1", "a2", "b", "c"} {
		release[id] = make(chan struct{}, 1)
	}
	t.Cleanup(func() {
		q.Stop()
		for _, ch := range release {
			ch <- struct{}{}
		}
	})
	for _, item := range []struct{ job, run string }{{"a", "a1"}, {"a", "a2"}, {"b", "b"}, {"c", "c"}} {
		q.EnqueueWithRun(&store.Job{ID: item.job}, &store.Run{ID: item.run})
	}
	q.Start(func(job *store.Job, run *store.Run) error {
		started <- run.ID
		<-release[run.ID]
		finished <- run.ID
		return errors.New("handler failure must release its reservation")
	})
	first, second := receive(t, started), receive(t, started)
	if !((first == "a1" && second == "b") || (first == "b" && second == "a1")) {
		t.Fatalf("first jobs = %s, %s; want a1 and b", first, second)
	}
	release["b"] <- struct{}{}
	if got := receive(t, started); got != "c" {
		t.Fatalf("next job = %s, want c while a1 is still running", got)
	}
	release["a1"] <- struct{}{}
	if got := receive(t, started); got != "a2" {
		t.Fatalf("next job = %s, want a2 after a1 finishes", got)
	}
	release["a2"] <- struct{}{}
	release["c"] <- struct{}{}
	for i := 0; i < 4; i++ {
		receive(t, finished)
	}
}

func TestWorkerLimit(t *testing.T) {
	for _, workers := range []int{1, 3} {
		t.Run(fmt.Sprint(workers), func(t *testing.T) {
			q := NewJobQueue(workers)
			started := make(chan string, 6)
			finished := make(chan string, 6)
			release := make(chan struct{})
			var mu sync.Mutex
			active, maximum := 0, 0
			t.Cleanup(func() { q.Stop(); close(release) })
			for i := 0; i < 6; i++ {
				q.Enqueue(&store.Job{ID: fmt.Sprint(i)})
			}
			q.Start(func(job *store.Job, _ *store.Run) error {
				mu.Lock()
				active++
				if active > maximum {
					maximum = active
				}
				mu.Unlock()
				started <- job.ID
				<-release
				mu.Lock()
				active--
				mu.Unlock()
				finished <- job.ID
				return nil
			})
			for i := 0; i < workers; i++ {
				receive(t, started)
			}
			select {
			case id := <-started:
				t.Fatalf("job %s exceeded worker limit %d", id, workers)
			case <-time.After(50 * time.Millisecond):
			}
			// Release one slot at a time; queued jobs should fill each free slot.
			for i := workers; i < 6; i++ {
				release <- struct{}{}
				receive(t, started)
			}
			for i := 0; i < workers; i++ {
				release <- struct{}{}
			}
			for i := 0; i < 6; i++ {
				receive(t, finished)
			}
			mu.Lock()
			defer mu.Unlock()
			if maximum != workers {
				t.Fatalf("maximum concurrency = %d, want %d", maximum, workers)
			}
		})
	}
}

func TestStopAndRepeatedStart(t *testing.T) {
	q := NewJobQueue(2)
	started := make(chan string, 3)
	release := make(chan struct{})
	t.Cleanup(func() { q.Stop(); close(release) })
	q.Enqueue(&store.Job{ID: "a"})
	q.Enqueue(&store.Job{ID: "b"})
	q.Enqueue(&store.Job{ID: "c"})
	handler := func(job *store.Job, _ *store.Run) error {
		started <- job.ID
		<-release
		return nil
	}
	q.Start(handler)
	q.Start(handler)
	receive(t, started)
	receive(t, started)
	q.Stop()
	q.Stop()
	q.Enqueue(&store.Job{ID: "after-stop"})
	q.Start(handler)
	if q.IsRunning() {
		t.Fatal("stopped queue is running")
	}
	select {
	case id := <-started:
		t.Fatalf("unexpected job %s after repeated start or stop", id)
	case <-time.After(50 * time.Millisecond):
	}
}
