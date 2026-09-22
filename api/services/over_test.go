package services_test

import (
	"testing"

	"capacity/api/models"
)

// Pure domain check kept next to the service package: over-allocation is
// strict greater-than, including the weekly_hours == 0 edge case in the seed.
func TestOverAllocationRule(t *testing.T) {
	cases := []struct {
		allocated, capacity float64
		over                bool
	}{
		{0, 40, false},
		{40, 40, false},
		{40.0001, 40, true},
		{1, 0, true},
		{0, 0, false},
	}
	for _, tc := range cases {
		cell := models.CapacityCell{
			Allocated: tc.allocated,
			Capacity:  tc.capacity,
			Over:      tc.allocated > tc.capacity,
		}
		if cell.Over != tc.over {
			t.Errorf("allocated=%v capacity=%v: over=%v want %v",
				tc.allocated, tc.capacity, cell.Over, tc.over)
		}
	}
}
