package store

import (
	"time"
)

// DailyExecutionStats represents execution statistics for a single day
type DailyExecutionStats struct {
	Date         string  `json:"date"`
	TotalRuns    int     `json:"total_runs"`
	SuccessCount int     `json:"success_count"`
	FailureCount int     `json:"failure_count"`
	TimeoutCount int     `json:"timeout_count"`
	SuccessRate  float64 `json:"success_rate"`
	AvgDuration  int64   `json:"avg_duration_ms"`
}

// JobStats represents statistics for a specific job
type JobStats struct {
	JobID        string  `json:"job_id"`
	JobName      string  `json:"job_name"`
	TotalRuns    int     `json:"total_runs"`
	SuccessCount int     `json:"success_count"`
	FailureCount int     `json:"failure_count"`
	SuccessRate  float64 `json:"success_rate"`
	AvgDuration  int64   `json:"avg_duration_ms"`
	MinDuration  int64   `json:"min_duration_ms"`
	MaxDuration  int64   `json:"max_duration_ms"`
	LastRunAt    *string `json:"last_run_at"`
}

// DurationDataPoint represents a single duration measurement over time
type DurationDataPoint struct {
	Date        string `json:"date"`
	AvgDuration int64  `json:"avg_duration_ms"`
	MinDuration int64  `json:"min_duration_ms"`
	MaxDuration int64  `json:"max_duration_ms"`
	RunCount    int    `json:"run_count"`
}

// GetExecutionTrends returns daily execution statistics for the specified number of days
func (s *Store) GetExecutionTrends(days int) ([]*DailyExecutionStats, error) {
	startDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02")

	query := `
		SELECT
			date(started_at) as date,
			COUNT(*) as total_runs,
			SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) as success_count,
			SUM(CASE WHEN status = 'failure' THEN 1 ELSE 0 END) as failure_count,
			SUM(CASE WHEN status = 'timeout' THEN 1 ELSE 0 END) as timeout_count,
			COALESCE(AVG(duration_ms), 0) as avg_duration
		FROM runs
		WHERE started_at IS NOT NULL
		AND date(started_at) >= ?
		AND status IN ('success', 'failure', 'timeout')
		GROUP BY date(started_at)
		ORDER BY date(started_at) ASC
	`

	rows, err := s.db.Query(query, startDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*DailyExecutionStats
	for rows.Next() {
		var stat DailyExecutionStats
		var avgDuration float64
		if err := rows.Scan(&stat.Date, &stat.TotalRuns, &stat.SuccessCount, &stat.FailureCount, &stat.TimeoutCount, &avgDuration); err != nil {
			return nil, err
		}
		stat.AvgDuration = int64(avgDuration)
		if stat.TotalRuns > 0 {
			stat.SuccessRate = float64(stat.SuccessCount) / float64(stat.TotalRuns)
		}
		stats = append(stats, &stat)
	}

	return stats, rows.Err()
}

// GetJobStats returns execution statistics for all jobs
func (s *Store) GetJobStats() ([]*JobStats, error) {
	query := `
		SELECT
			j.id,
			j.name,
			COUNT(r.id) as total_runs,
			COALESCE(SUM(CASE WHEN r.status = 'success' THEN 1 ELSE 0 END), 0) as success_count,
			COALESCE(SUM(CASE WHEN r.status IN ('failure', 'timeout') THEN 1 ELSE 0 END), 0) as failure_count,
			COALESCE(AVG(r.duration_ms), 0) as avg_duration,
			COALESCE(MIN(r.duration_ms), 0) as min_duration,
			COALESCE(MAX(r.duration_ms), 0) as max_duration,
			MAX(r.started_at) as last_run_at
		FROM jobs j
		LEFT JOIN runs r ON j.id = r.job_id AND r.status IN ('success', 'failure', 'timeout')
		GROUP BY j.id, j.name
		ORDER BY total_runs DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*JobStats
	for rows.Next() {
		var stat JobStats
		var avgDuration, minDuration, maxDuration float64
		if err := rows.Scan(&stat.JobID, &stat.JobName, &stat.TotalRuns, &stat.SuccessCount, &stat.FailureCount, &avgDuration, &minDuration, &maxDuration, &stat.LastRunAt); err != nil {
			return nil, err
		}
		stat.AvgDuration = int64(avgDuration)
		stat.MinDuration = int64(minDuration)
		stat.MaxDuration = int64(maxDuration)
		if stat.TotalRuns > 0 {
			stat.SuccessRate = float64(stat.SuccessCount) / float64(stat.TotalRuns)
		}
		stats = append(stats, &stat)
	}

	return stats, rows.Err()
}

// GetJobDurationTrends returns duration trends for a specific job over time
func (s *Store) GetJobDurationTrends(jobID string, days int) ([]*DurationDataPoint, error) {
	startDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02")

	query := `
		SELECT
			date(started_at) as date,
			AVG(duration_ms) as avg_duration,
			MIN(duration_ms) as min_duration,
			MAX(duration_ms) as max_duration,
			COUNT(*) as run_count
		FROM runs
		WHERE job_id = ?
		AND started_at IS NOT NULL
		AND date(started_at) >= ?
		AND status IN ('success', 'failure', 'timeout')
		AND duration_ms IS NOT NULL
		GROUP BY date(started_at)
		ORDER BY date(started_at) ASC
	`

	rows, err := s.db.Query(query, jobID, startDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []*DurationDataPoint
	for rows.Next() {
		var point DurationDataPoint
		var avgDuration, minDuration, maxDuration float64
		if err := rows.Scan(&point.Date, &avgDuration, &minDuration, &maxDuration, &point.RunCount); err != nil {
			return nil, err
		}
		point.AvgDuration = int64(avgDuration)
		point.MinDuration = int64(minDuration)
		point.MaxDuration = int64(maxDuration)
		points = append(points, &point)
	}

	return points, rows.Err()
}

// GetOverallStats returns overall system statistics
func (s *Store) GetOverallStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total runs and success rate
	var totalRuns, successCount, failureCount int
	var avgDuration float64
	err := s.db.QueryRow(`
		SELECT
			COUNT(*),
			COALESCE(SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status IN ('failure', 'timeout') THEN 1 ELSE 0 END), 0),
			COALESCE(AVG(duration_ms), 0)
		FROM runs
		WHERE status IN ('success', 'failure', 'timeout')
	`).Scan(&totalRuns, &successCount, &failureCount, &avgDuration)
	if err != nil {
		return nil, err
	}

	stats["total_runs"] = totalRuns
	stats["success_count"] = successCount
	stats["failure_count"] = failureCount
	stats["avg_duration_ms"] = int64(avgDuration)
	if totalRuns > 0 {
		stats["success_rate"] = float64(successCount) / float64(totalRuns)
	} else {
		stats["success_rate"] = 0.0
	}

	// Runs in last 24 hours
	var last24h int
	err = s.db.QueryRow(`
		SELECT COUNT(*) FROM runs
		WHERE started_at >= datetime('now', '-24 hours')
		AND status IN ('success', 'failure', 'timeout')
	`).Scan(&last24h)
	if err != nil {
		return nil, err
	}
	stats["runs_last_24h"] = last24h

	// Runs in last 7 days
	var last7d int
	err = s.db.QueryRow(`
		SELECT COUNT(*) FROM runs
		WHERE started_at >= datetime('now', '-7 days')
		AND status IN ('success', 'failure', 'timeout')
	`).Scan(&last7d)
	if err != nil {
		return nil, err
	}
	stats["runs_last_7d"] = last7d

	// Total jobs
	var totalJobs int
	err = s.db.QueryRow(`SELECT COUNT(*) FROM jobs`).Scan(&totalJobs)
	if err != nil {
		return nil, err
	}
	stats["total_jobs"] = totalJobs

	// Active jobs (enabled)
	var activeJobs int
	err = s.db.QueryRow(`SELECT COUNT(*) FROM jobs WHERE enabled = 1`).Scan(&activeJobs)
	if err != nil {
		return nil, err
	}
	stats["active_jobs"] = activeJobs

	return stats, nil
}
