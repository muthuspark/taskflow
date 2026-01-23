package store

import (
	"fmt"
	"time"
)

// AddLog adds a log entry for a run
func (s *Store) AddLog(runID, stream, content string) (*LogEntry, error) {
	log := &LogEntry{
		RunID:     runID,
		Timestamp: time.Now(),
		Stream:    stream,
		Content:   content,
	}

	result, err := s.db.Exec(
		`INSERT INTO logs (run_id, timestamp, stream, content) VALUES (?, ?, ?, ?)`,
		log.RunID, log.Timestamp, log.Stream, log.Content,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add log: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get log id: %w", err)
	}

	log.ID = int(id)
	return log, nil
}

// GetLogs retrieves logs for a run
func (s *Store) GetLogs(runID string) ([]*LogEntry, error) {
	return s.GetLogsPaginated(runID, 0, 0)
}

// GetLogsPaginated retrieves logs for a run with limit/offset support.
// If limit is 0, all logs are returned.
func (s *Store) GetLogsPaginated(runID string, limit, offset int) ([]*LogEntry, error) {
	var query string
	var args []interface{}

	if limit > 0 {
		query = `SELECT id, run_id, timestamp, stream, content FROM logs WHERE run_id = ? ORDER BY id ASC LIMIT ? OFFSET ?`
		args = []interface{}{runID, limit, offset}
	} else {
		query = `SELECT id, run_id, timestamp, stream, content FROM logs WHERE run_id = ? ORDER BY id ASC`
		args = []interface{}{runID}
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}
	defer rows.Close()

	logs := make([]*LogEntry, 0)
	for rows.Next() {
		log := &LogEntry{}
		if err := rows.Scan(&log.ID, &log.RunID, &log.Timestamp, &log.Stream, &log.Content); err != nil {
			return nil, fmt.Errorf("failed to scan log: %w", err)
		}
		logs = append(logs, log)
	}

	return logs, rows.Err()
}

// GetLogCount returns the total number of log entries for a run.
func (s *Store) GetLogCount(runID string) (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM logs WHERE run_id = ?`, runID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count logs: %w", err)
	}
	return count, nil
}

// DeleteLogs deletes logs for a run
func (s *Store) DeleteLogs(runID string) error {
	_, err := s.db.Exec(`DELETE FROM logs WHERE run_id = ?`, runID)
	return err
}
