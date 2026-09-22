package dates

import "time"

// WeeksInRange returns Monday dates for every ISO week overlapping [from, to].
// Matches Postgres date_trunc('week', ...), which uses Monday as week start.
func WeeksInRange(from, to time.Time) []time.Time {
	start := MondayOf(from)
	end := MondayOf(to)
	var weeks []time.Time
	for d := start; !d.After(end); d = d.AddDate(0, 0, 7) {
		weeks = append(weeks, d)
	}
	return weeks
}

// MondayOf returns the Monday of the ISO week containing t (UTC, date-only).
func MondayOf(t time.Time) time.Time {
	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	wd := int(t.Weekday())
	if wd == 0 {
		wd = 7
	}
	return t.AddDate(0, 0, -(wd - 1))
}

// ParseDate parses YYYY-MM-DD as a UTC calendar date.
func ParseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}

// FormatDate formats t as YYYY-MM-DD.
func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}
