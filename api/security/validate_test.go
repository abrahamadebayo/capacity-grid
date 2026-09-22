package security_test

import (
	"math"
	"testing"

	"capacity/api/models"
	"capacity/api/security"
)

func TestParseRange(t *testing.T) {
	_, _, err := security.ParseRange("", "2026-01-01")
	if err != security.ErrMissingRange {
		t.Errorf("want ErrMissingRange, got %v", err)
	}

	_, _, err = security.ParseRange("2026-01-10", "2026-01-01")
	if err != security.ErrRangeOrder {
		t.Errorf("want ErrRangeOrder, got %v", err)
	}

	_, _, err = security.ParseRange("not-a-date", "2026-01-01")
	if err != security.ErrInvalidFrom {
		t.Errorf("want ErrInvalidFrom, got %v", err)
	}

	from, to, err := security.ParseRange("2025-12-29", "2026-01-18")
	if err != nil {
		t.Fatal(err)
	}
	if from.Format("2006-01-02") != "2025-12-29" || to.Format("2006-01-02") != "2026-01-18" {
		t.Errorf("unexpected parse: %v %v", from, to)
	}
}

func TestParseRangeTooWide(t *testing.T) {
	_, _, err := security.ParseRange("2025-01-01", "2026-12-31")
	if err != security.ErrRangeTooWide {
		t.Errorf("want ErrRangeTooWide, got %v", err)
	}
}

func TestParsePersonID(t *testing.T) {
	id, err := security.ParsePersonID("42")
	if err != nil || id != 42 {
		t.Errorf("got %d %v", id, err)
	}
	if _, err := security.ParsePersonID("0"); err != security.ErrInvalidPersonID {
		t.Errorf("want invalid for 0, got %v", err)
	}
	if _, err := security.ParsePersonID("12x"); err != security.ErrInvalidPersonID {
		t.Errorf("want invalid for 12x, got %v", err)
	}
	if _, err := security.ParsePersonID("-3"); err != security.ErrInvalidPersonID {
		t.Errorf("want invalid for -3, got %v", err)
	}
}

func TestValidateUpdatePerson(t *testing.T) {
	if _, err := security.ValidateUpdatePerson(models.UpdatePersonRequest{}); err != security.ErrMissingHours {
		t.Errorf("want missing, got %v", err)
	}

	neg := -1.0
	if _, err := security.ValidateUpdatePerson(models.UpdatePersonRequest{WeeklyHours: &neg}); err != security.ErrNegativeHours {
		t.Errorf("want negative, got %v", err)
	}

	big := 200.0
	if _, err := security.ValidateUpdatePerson(models.UpdatePersonRequest{WeeklyHours: &big}); err != security.ErrHoursTooLarge {
		t.Errorf("want too large, got %v", err)
	}

	nan := math.NaN()
	if _, err := security.ValidateUpdatePerson(models.UpdatePersonRequest{WeeklyHours: &nan}); err != security.ErrNonFiniteHours {
		t.Errorf("want non-finite, got %v", err)
	}

	ok := 32.0
	got, err := security.ValidateUpdatePerson(models.UpdatePersonRequest{WeeklyHours: &ok})
	if err != nil || got != 32 {
		t.Errorf("got %v %v", got, err)
	}
}
