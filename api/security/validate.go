package security

import (
	"errors"
	"fmt"
	"math"
	"time"

	"capacity/api/dates"
	"capacity/api/models"
)

var (
	ErrMissingRange     = errors.New("from and to query params are required (YYYY-MM-DD)")
	ErrInvalidFrom      = errors.New("invalid from date")
	ErrInvalidTo        = errors.New("invalid to date")
	ErrRangeOrder       = errors.New("to must be on or after from")
	ErrRangeTooWide     = errors.New("date range must span at most 52 weeks")
	ErrInvalidPersonID  = errors.New("invalid person id")
	ErrMissingHours     = errors.New("weeklyHours is required")
	ErrNegativeHours    = errors.New("weeklyHours must be >= 0")
	ErrHoursTooLarge    = errors.New("weeklyHours must be <= 168")
	ErrNonFiniteHours   = errors.New("weeklyHours must be a finite number")
)

const maxWeeks = 52

// ParseRange validates and parses from/to query params.
func ParseRange(fromStr, toStr string) (from, to time.Time, err error) {
	if fromStr == "" || toStr == "" {
		return time.Time{}, time.Time{}, ErrMissingRange
	}
	from, err = dates.ParseDate(fromStr)
	if err != nil {
		return time.Time{}, time.Time{}, ErrInvalidFrom
	}
	to, err = dates.ParseDate(toStr)
	if err != nil {
		return time.Time{}, time.Time{}, ErrInvalidTo
	}
	if to.Before(from) {
		return time.Time{}, time.Time{}, ErrRangeOrder
	}
	weeks := dates.WeeksInRange(from, to)
	if len(weeks) > maxWeeks {
		return time.Time{}, time.Time{}, ErrRangeTooWide
	}
	return from, to, nil
}

// ParsePersonID validates a path id.
func ParsePersonID(raw string) (int, error) {
	var id int
	if _, err := fmt.Sscanf(raw, "%d", &id); err != nil || id < 1 {
		return 0, ErrInvalidPersonID
	}
	// Reject leading junk like "12abc" — Sscanf succeeds on prefix.
	if fmt.Sprintf("%d", id) != raw {
		return 0, ErrInvalidPersonID
	}
	return id, nil
}

// ValidateUpdatePerson checks a PATCH body.
func ValidateUpdatePerson(body models.UpdatePersonRequest) (float64, error) {
	if body.WeeklyHours == nil {
		return 0, ErrMissingHours
	}
	h := *body.WeeklyHours
	if math.IsNaN(h) || math.IsInf(h, 0) {
		return 0, ErrNonFiniteHours
	}
	if h < 0 {
		return 0, ErrNegativeHours
	}
	if h > 168 {
		return 0, ErrHoursTooLarge
	}
	return h, nil
}
