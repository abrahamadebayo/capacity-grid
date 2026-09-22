package dates_test

import (
	"testing"
	"time"

	"capacity/api/dates"
)

func TestMondayOf(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"2025-12-29", "2025-12-29"}, // Monday
		{"2025-12-31", "2025-12-29"}, // Wednesday
		{"2026-01-04", "2025-12-29"}, // Sunday
		{"2026-01-05", "2026-01-05"}, // next Monday
	}
	for _, tc := range cases {
		got := dates.MondayOf(mustDate(t, tc.in))
		if dates.FormatDate(got) != tc.want {
			t.Errorf("MondayOf(%s) = %s, want %s", tc.in, dates.FormatDate(got), tc.want)
		}
	}
}

func TestWeeksInRange(t *testing.T) {
	from := mustDate(t, "2025-12-29")
	to := mustDate(t, "2026-01-16")
	weeks := dates.WeeksInRange(from, to)
	if len(weeks) != 3 {
		t.Fatalf("got %d weeks, want 3", len(weeks))
	}
	want := []string{"2025-12-29", "2026-01-05", "2026-01-12"}
	for i, w := range want {
		if dates.FormatDate(weeks[i]) != w {
			t.Errorf("week[%d] = %s, want %s", i, dates.FormatDate(weeks[i]), w)
		}
	}
}

func TestWeeksInRangeSingleDay(t *testing.T) {
	d := mustDate(t, "2026-01-07")
	weeks := dates.WeeksInRange(d, d)
	if len(weeks) != 1 || dates.FormatDate(weeks[0]) != "2026-01-05" {
		t.Fatalf("unexpected weeks: %v", weeks)
	}
}

func mustDate(t *testing.T, s string) time.Time {
	t.Helper()
	d, err := dates.ParseDate(s)
	if err != nil {
		t.Fatal(err)
	}
	return d
}
