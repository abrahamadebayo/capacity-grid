package services

import (
	"context"
	"errors"
	"time"

	"capacity/api/dates"
	"capacity/api/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")

// CapacityService loads allocation vs capacity for a date range.
type CapacityService struct {
	db *pgxpool.Pool
}

func NewCapacityService(db *pgxpool.Pool) *CapacityService {
	return &CapacityService{db: db}
}

// capacityQuery complexity (P people, A assignments, W weeks in range):
//
//   Allocation uses the (start_date, end_date) index once per week:
//     O(W · (log A + K_w)) ≈ O(W log A + K) where K_w is overlaps that week
//     and K = Σ K_w. Day-count math inside the sum is O(1) per overlapping row.
//
//   People ⨯ weeks is O(P · W) and unavoidable for a dense grid (zeros included).
//   Response assembly in Go is also O(P · W).
//
// A single whole-window filter + LATERAL week expansion was tried; with this
// seed Postgres fell back to scanning all A rows via person_id, which is
// worse for small W. Per-week index probes stay near the matching subset
// (≈4.7k of 126k for the default 3-week window).
const capacityQuery = `
WITH weeks AS (
  SELECT generate_series(
    date_trunc('week', $1::date)::date,
    date_trunc('week', $2::date)::date,
    '1 week'::interval
  )::date AS week_start
),
allocation AS (
  SELECT
    a.person_id,
    w.week_start,
    coalesce(sum(
      a.hours_per_day * (
        least(a.end_date, w.week_start + 6) -
        greatest(a.start_date, w.week_start) + 1
      )
    ), 0)::float8 AS allocated
  FROM weeks w
  JOIN assignments a
    ON a.start_date <= w.week_start + 6
   AND a.end_date   >= w.week_start
  GROUP BY a.person_id, w.week_start
)
SELECT
  p.id,
  p.name,
  p.weekly_hours::float8,
  w.week_start,
  coalesce(al.allocated, 0)::float8 AS allocated
FROM people p
CROSS JOIN weeks w
LEFT JOIN allocation al
  ON al.person_id = p.id AND al.week_start = w.week_start
ORDER BY p.name, p.id, w.week_start
`

// Get returns capacity for every person across ISO weeks in [from, to].
func (s *CapacityService) Get(ctx context.Context, from, to time.Time) (models.CapacityResponse, error) {
	weeks := dates.WeeksInRange(from, to)

	rows, err := s.db.Query(ctx, capacityQuery, from, to)
	if err != nil {
		return models.CapacityResponse{}, err
	}
	defer rows.Close()

	weekHeaders := make([]models.WeekHeader, len(weeks))
	for i, ws := range weeks {
		weekHeaders[i] = models.WeekHeader{
			Start: dates.FormatDate(ws),
			End:   dates.FormatDate(ws.AddDate(0, 0, 6)),
		}
	}

	byPerson := map[int]*models.CapacityRow{}
	var order []int

	for rows.Next() {
		var (
			id          int
			name        string
			weeklyHours float64
			weekStart   time.Time
			allocated   float64
		)
		if err := rows.Scan(&id, &name, &weeklyHours, &weekStart, &allocated); err != nil {
			return models.CapacityResponse{}, err
		}

		row, ok := byPerson[id]
		if !ok {
			row = &models.CapacityRow{
				PersonID:    id,
				Name:        name,
				WeeklyHours: weeklyHours,
				Weeks:       make([]models.CapacityCell, 0, len(weeks)),
			}
			byPerson[id] = row
			order = append(order, id)
		}

		row.Weeks = append(row.Weeks, models.CapacityCell{
			WeekStart: dates.FormatDate(weekStart),
			Allocated: allocated,
			Capacity:  weeklyHours,
			Over:      allocated > weeklyHours,
		})
	}
	if err := rows.Err(); err != nil {
		return models.CapacityResponse{}, err
	}

	out := make([]models.CapacityRow, 0, len(order))
	for _, id := range order {
		out = append(out, *byPerson[id])
	}

	return models.CapacityResponse{
		From:  dates.FormatDate(from),
		To:    dates.FormatDate(to),
		Weeks: weekHeaders,
		Rows:  out,
	}, nil
}

// PeopleService updates people.
type PeopleService struct {
	db *pgxpool.Pool
}

func NewPeopleService(db *pgxpool.Pool) *PeopleService {
	return &PeopleService{db: db}
}

// UpdateWeeklyHours sets a person's capacity and returns the updated row.
func (s *PeopleService) UpdateWeeklyHours(ctx context.Context, id int, weeklyHours float64) (models.Person, error) {
	var person models.Person
	err := s.db.QueryRow(
		ctx,
		`UPDATE people
		 SET weekly_hours = $1
		 WHERE id = $2
		 RETURNING id, name, weekly_hours::float8`,
		weeklyHours,
		id,
	).Scan(&person.ID, &person.Name, &person.WeeklyHours)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Person{}, ErrNotFound
	}
	if err != nil {
		return models.Person{}, err
	}
	return person, nil
}

// HealthService reports basic readiness.
type HealthService struct {
	db *pgxpool.Pool
}

func NewHealthService(db *pgxpool.Pool) *HealthService {
	return &HealthService{db: db}
}

func (s *HealthService) Check(ctx context.Context) (models.HealthResponse, error) {
	var people int
	if err := s.db.QueryRow(ctx, `SELECT count(*) FROM people`).Scan(&people); err != nil {
		return models.HealthResponse{}, err
	}
	return models.HealthResponse{OK: true, People: people}, nil
}
