package scheduler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/taskflow/taskflow/internal/store"
)

// TestMatchesField tests single field matching
func TestMatchesField(t *testing.T) {
	m := NewMatcher()

	tests := []struct {
		name     string
		allowed  []int
		value    int
		expected bool
	}{
		{
			name:     "empty list matches any",
			allowed:  []int{},
			value:    5,
			expected: true,
		},
		{
			name:     "nil list matches any",
			allowed:  nil,
			value:    5,
			expected: true,
		},
		{
			name:     "value in list",
			allowed:  []int{1, 5, 10},
			value:    5,
			expected: true,
		},
		{
			name:     "value not in list",
			allowed:  []int{1, 10},
			value:    5,
			expected: false,
		},
		{
			name:     "single value matches",
			allowed:  []int{5},
			value:    5,
			expected: true,
		},
		{
			name:     "zero value matches",
			allowed:  []int{0, 5, 10},
			value:    0,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.matchesField(tt.allowed, tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMatches tests complete schedule matching
func TestMatches(t *testing.T) {
	m := NewMatcher()

	tests := []struct {
		name     string
		time     time.Time
		schedule *store.Schedule
		expected bool
	}{
		{
			name: "all fields nil matches any time",
			time: time.Date(2026, time.January, 15, 14, 30, 0, 0, time.UTC),
			schedule: &store.Schedule{
				Years:    nil,
				Months:   nil,
				Days:     nil,
				Weekdays: nil,
				Hours:    nil,
				Minutes:  nil,
			},
			expected: true,
		},
		{
			name: "specific month and day",
			time: time.Date(2026, time.March, 15, 14, 30, 0, 0, time.UTC),
			schedule: &store.Schedule{
				Years:    nil,
				Months:   []int{3},     // March
				Days:     []int{15},    // 15th
				Weekdays: nil,
				Hours:    nil,
				Minutes:  nil,
			},
			expected: true,
		},
		{
			name: "specific month no match",
			time: time.Date(2026, time.February, 15, 14, 30, 0, 0, time.UTC),
			schedule: &store.Schedule{
				Years:    nil,
				Months:   []int{3},     // March only
				Days:     []int{15},
				Weekdays: nil,
				Hours:    nil,
				Minutes:  nil,
			},
			expected: false,
		},
		{
			name: "specific time and minute",
			time: time.Date(2026, time.January, 15, 14, 30, 0, 0, time.UTC),
			schedule: &store.Schedule{
				Years:    nil,
				Months:   nil,
				Days:     nil,
				Weekdays: nil,
				Hours:    []int{14},    // 14:00
				Minutes:  []int{30},    // :30
			},
			expected: true,
		},
		{
			name: "wrong minute no match",
			time: time.Date(2026, time.January, 15, 14, 31, 0, 0, time.UTC),
			schedule: &store.Schedule{
				Years:    nil,
				Months:   nil,
				Days:     nil,
				Weekdays: nil,
				Hours:    []int{14},
				Minutes:  []int{30},
			},
			expected: false,
		},
		{
			name: "weekday match",
			time: time.Date(2026, time.January, 15, 14, 30, 0, 0, time.UTC), // Thursday = 4
			schedule: &store.Schedule{
				Years:    nil,
				Months:   nil,
				Days:     nil,
				Weekdays: []int{4},    // Thursday
				Hours:    nil,
				Minutes:  nil,
			},
			expected: true,
		},
		{
			name: "weekday no match",
			time: time.Date(2026, time.January, 15, 14, 30, 0, 0, time.UTC), // Thursday = 4
			schedule: &store.Schedule{
				Years:    nil,
				Months:   nil,
				Days:     nil,
				Weekdays: []int{0, 1, 2}, // Sun-Tue only
				Hours:    nil,
				Minutes:  nil,
			},
			expected: false,
		},
		{
			name: "multiple months match",
			time: time.Date(2026, time.June, 15, 14, 30, 0, 0, time.UTC),
			schedule: &store.Schedule{
				Years:    nil,
				Months:   []int{3, 6, 9, 12},    // Every 3 months
				Days:     nil,
				Weekdays: nil,
				Hours:    nil,
				Minutes:  nil,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.Matches(tt.time, tt.schedule)
			assert.Equal(t, tt.expected, result, "for time %v with schedule %v", tt.time, tt.schedule)
		})
	}
}

// TestNextScheduledTime tests calculating next execution time
func TestNextScheduledTime(t *testing.T) {
	m := NewMatcher()

	// Start time: Jan 15, 2026 14:30
	from := time.Date(2026, time.January, 15, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		schedule *store.Schedule
		from     time.Time
		// We check that next scheduled is after from and matches schedule
		expectAfter bool
	}{
		{
			name: "find next matching minute",
			schedule: &store.Schedule{
				Years:    nil,
				Months:   nil,
				Days:     nil,
				Weekdays: nil,
				Hours:    []int{14, 15}, // 14:00 and 15:00
				Minutes:  []int{0, 30},   // :00 and :30
			},
			from:        from,
			expectAfter: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next := m.NextScheduledTime(tt.schedule, tt.from)

			// Should find a time
			assert.False(t, next.IsZero(), "should find next scheduled time")

			// Should be after the from time
			if tt.expectAfter {
				assert.True(t, next.After(tt.from), "next should be after from")
			}

			// Should match the schedule
			assert.True(t, m.Matches(next, tt.schedule), "next time should match schedule")
		})
	}
}

// TestEdgeCases tests edge cases in schedule matching
func TestEdgeCases(t *testing.T) {
	m := NewMatcher()

	tests := []struct {
		name     string
		time     time.Time
		schedule *store.Schedule
		expected bool
	}{
		{
			name: "leap year date",
			time: time.Date(2024, time.February, 29, 0, 0, 0, 0, time.UTC),
			schedule: &store.Schedule{
				Months: []int{2},
				Days:   []int{29},
			},
			expected: true,
		},
		{
			name: "midnight",
			time: time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC),
			schedule: &store.Schedule{
				Hours:   []int{0},
				Minutes: []int{0},
			},
			expected: true,
		},
		{
			name: "last hour of day",
			time: time.Date(2026, time.January, 15, 23, 59, 0, 0, time.UTC),
			schedule: &store.Schedule{
				Hours:   []int{23},
				Minutes: []int{59},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.Matches(tt.time, tt.schedule)
			assert.Equal(t, tt.expected, result)
		})
	}
}
