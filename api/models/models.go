package models

// CapacityResponse is GET /api/capacity.
// Weeks are ISO weeks (Monday–Sunday). Every person appears once with a
// cell for each week in the requested range.
type CapacityResponse struct {
	From  string        `json:"from"`
	To    string        `json:"to"`
	Weeks []WeekHeader  `json:"weeks"`
	Rows  []CapacityRow `json:"rows"`
}

type WeekHeader struct {
	Start string `json:"start"` // Monday YYYY-MM-DD
	End   string `json:"end"`   // Sunday YYYY-MM-DD
}

type CapacityRow struct {
	PersonID    int            `json:"personId"`
	Name        string         `json:"name"`
	WeeklyHours float64        `json:"weeklyHours"`
	Weeks       []CapacityCell `json:"weeks"`
}

type CapacityCell struct {
	WeekStart string  `json:"weekStart"`
	Allocated float64 `json:"allocated"`
	Capacity  float64 `json:"capacity"`
	Over      bool    `json:"over"`
}

// UpdatePersonRequest is PATCH /api/people/{id}.
type UpdatePersonRequest struct {
	WeeklyHours *float64 `json:"weeklyHours"`
}

// Person is the public person payload returned after updates (and usable
// by the grid to patch local capacity without a full refetch).
type Person struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	WeeklyHours float64 `json:"weeklyHours"`
}

// HealthResponse is GET /api/health.
type HealthResponse struct {
	OK     bool `json:"ok"`
	People int  `json:"people"`
}
