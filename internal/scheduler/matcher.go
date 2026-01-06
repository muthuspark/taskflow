package scheduler

import (
	"time"

	"github.com/taskflow/taskflow/internal/store"
)

// Matcher checks if a given time matches a schedule
type Matcher struct{}

// NewMatcher creates a new schedule matcher
func NewMatcher() *Matcher {
	return &Matcher{}
}

// Matches checks if the given time matches the schedule
func (m *Matcher) Matches(t time.Time, schedule *store.Schedule) bool {
	return m.matchesField(schedule.Years, t.Year()) &&
		m.matchesField(schedule.Months, int(t.Month())) &&
		m.matchesField(schedule.Days, t.Day()) &&
		m.matchesField(schedule.Weekdays, int(t.Weekday())) &&
		m.matchesField(schedule.Hours, t.Hour()) &&
		m.matchesField(schedule.Minutes, t.Minute())
}

// matchesField checks if value is in allowed list (nil/empty means any)
func (m *Matcher) matchesField(allowed []int, value int) bool {
	if allowed == nil || len(allowed) == 0 {
		return true // nil or empty means "any"
	}
	for _, v := range allowed {
		if v == value {
			return true
		}
	}
	return false
}

// NextScheduledTime calculates the next execution time based on schedule
// This is a simplified implementation that checks minute-by-minute
func (m *Matcher) NextScheduledTime(schedule *store.Schedule, from time.Time) time.Time {
	t := from.Add(time.Minute)
	// Truncate to minute boundary
	t = t.Truncate(time.Minute)

	// Check next 365 days for matching time
	for i := 0; i < 365*24*60; i++ {
		if m.Matches(t, schedule) {
			return t
		}
		t = t.Add(time.Minute)
	}

	// No match found in the next year
	return time.Time{}
}
