package store

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPoolConnectionsHaveWriteTimeout(t *testing.T) {
	for _, suffix := range []string{"", "?mode=rwc"} {
		t.Run(suffix, func(t *testing.T) {
			st, err := New(filepath.Join(t.TempDir(), "pool.db") + suffix)
			require.NoError(t, err)
			defer st.Close()
			// Hold every connection at once so each query checks a distinct connection.
			for i := 0; i < 4; i++ {
				conn, err := st.DB().Conn(context.Background())
				require.NoError(t, err)
				defer conn.Close()
				var timeout int
				require.NoError(t, conn.QueryRowContext(context.Background(), "PRAGMA busy_timeout").Scan(&timeout))
				require.Equal(t, 5000, timeout)
			}
		})
	}
}
