package store

import (
	"fmt"
	"time"
)

// AddMetric records resource usage for a run
func (s *Store) AddMetric(runID string, cpuPercent, memoryPercent float64, memoryBytes int64) (*Metric, error) {
	metric := &Metric{
		RunID:         runID,
		Timestamp:     time.Now(),
		CPUPercent:    cpuPercent,
		MemoryBytes:   memoryBytes,
		MemoryPercent: memoryPercent,
	}

	result, err := s.db.Exec(
		`INSERT INTO metrics (run_id, timestamp, cpu_percent, memory_bytes, memory_percent)
		 VALUES (?, ?, ?, ?, ?)`,
		metric.RunID, metric.Timestamp, metric.CPUPercent, metric.MemoryBytes, metric.MemoryPercent,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add metric: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get metric id: %w", err)
	}

	metric.ID = int(id)
	return metric, nil
}

// GetMetrics retrieves metrics for a run
func (s *Store) GetMetrics(runID string) ([]*Metric, error) {
	rows, err := s.db.Query(
		`SELECT id, run_id, timestamp, cpu_percent, memory_bytes, memory_percent
		 FROM metrics WHERE run_id = ? ORDER BY timestamp ASC`,
		runID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}
	defer rows.Close()

	var metrics []*Metric
	for rows.Next() {
		m := &Metric{}
		if err := rows.Scan(&m.ID, &m.RunID, &m.Timestamp, &m.CPUPercent, &m.MemoryBytes, &m.MemoryPercent); err != nil {
			return nil, fmt.Errorf("failed to scan metric: %w", err)
		}
		metrics = append(metrics, m)
	}

	return metrics, rows.Err()
}

// DeleteMetrics deletes metrics for a run
func (s *Store) DeleteMetrics(runID string) error {
	_, err := s.db.Exec(`DELETE FROM metrics WHERE run_id = ?`, runID)
	return err
}
