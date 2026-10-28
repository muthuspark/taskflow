package scheduler

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/taskflow/taskflow/internal/executor"
	"github.com/taskflow/taskflow/internal/store"
)

func TestWorkersExecuteAndPersistConcurrentRuns(t *testing.T) {
	root := t.TempDir()
	st, err := store.New(filepath.Join(root, "runs.db"))
	require.NoError(t, err)
	defer st.Close()
	exec := executor.New(st)
	started := make(chan string, 2)
	finished := make(chan string, 2)
	exec.SetStatusBroadcaster(func(id, status string) {
		if status == "running" {
			started <- id
		}
	})
	q := NewJobQueue(2)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer q.Stop()
	runs := make([]*store.Run, 0, 2)
	for _, name := range []string{"alpha", "beta"} {
		job, err := st.CreateJob(&store.Job{
			Name:           name,
			WorkingDir:     root,
			TimeoutSeconds: 5,
			Script:         "while [ ! -f release ]; do sleep 0.01; done; for i in {1..100}; do echo log-$i; done",
		})
		require.NoError(t, err)
		run, err := st.CreateRun(job.ID, "manual")
		require.NoError(t, err)
		runs = append(runs, run)
		q.EnqueueWithRun(job, run)
	}
	q.Start(func(job *store.Job, run *store.Run) error {
		err := exec.Execute(ctx, run, job)
		finished <- run.ID
		return err
	})
	// Neither script can complete until both runs are concurrently running.
	receive(t, started)
	receive(t, started)
	for _, run := range runs {
		stored, err := st.GetRun(run.ID)
		require.NoError(t, err)
		require.Equal(t, "running", stored.Status)
		// The API's original run object must remain safe to serialize.
		require.Equal(t, "pending", run.Status)
	}
	require.NoError(t, os.WriteFile(filepath.Join(root, "release"), nil, 0600))
	receive(t, finished)
	receive(t, finished)
	for _, run := range runs {
		stored, err := st.GetRun(run.ID)
		require.NoError(t, err)
		require.Equal(t, "success", stored.Status)
		var count int
		require.NoError(t, st.DB().QueryRow("SELECT count(*) FROM logs WHERE run_id = ?", run.ID).Scan(&count))
		require.Positive(t, count, "each concurrent run should persist its logs")
	}
}
