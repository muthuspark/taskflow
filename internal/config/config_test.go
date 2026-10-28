package config

import "testing"

func TestWorkerCount(t *testing.T) {
	for _, tc := range []struct {
		value string
		want  int
	}{{"", 3}, {"1", 1}, {"5", 5}, {"0", 3}, {"-2", 3}, {"invalid", 3}} {
		t.Run("workers="+tc.value, func(t *testing.T) {
			t.Setenv("WORKER_COUNT", tc.value)
			if got := Load().WorkerCount; got != tc.want {
				t.Fatalf("got %d workers, want %d", got, tc.want)
			}
		})
	}
}
