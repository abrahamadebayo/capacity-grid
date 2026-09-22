package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"capacity/api/handlers"
	"capacity/api/models"
	"capacity/api/services"
)

type fakeCapacity struct {
	resp models.CapacityResponse
	err  error
}

func (f fakeCapacity) Get(ctx context.Context, from, to time.Time) (models.CapacityResponse, error) {
	return f.resp, f.err
}

type fakePeople struct {
	person models.Person
	err    error
}

func (f fakePeople) UpdateWeeklyHours(ctx context.Context, id int, weeklyHours float64) (models.Person, error) {
	if f.err != nil {
		return models.Person{}, f.err
	}
	p := f.person
	p.ID = id
	p.WeeklyHours = weeklyHours
	return p, nil
}

type fakeHealth struct {
	resp models.HealthResponse
	err  error
}

func (f fakeHealth) Check(ctx context.Context) (models.HealthResponse, error) {
	return f.resp, f.err
}

func TestHandleCapacityOK(t *testing.T) {
	api := &handlers.API{
		Capacity: fakeCapacity{resp: models.CapacityResponse{
			From: "2025-12-29",
			To:   "2026-01-18",
			Rows: []models.CapacityRow{{PersonID: 1, Name: "Ana", WeeklyHours: 40}},
		}},
	}
	req := httptest.NewRequest(http.MethodGet, "/api/capacity?from=2025-12-29&to=2026-01-18", nil)
	rr := httptest.NewRecorder()
	api.HandleCapacity(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rr.Code, rr.Body.String())
	}
	var body models.CapacityResponse
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if len(body.Rows) != 1 || body.Rows[0].Name != "Ana" {
		t.Fatalf("unexpected body: %+v", body)
	}
}

func TestHandleCapacityBadRange(t *testing.T) {
	api := &handlers.API{Capacity: fakeCapacity{}}
	req := httptest.NewRequest(http.MethodGet, "/api/capacity?from=2026-01-10&to=2026-01-01", nil)
	rr := httptest.NewRecorder()
	api.HandleCapacity(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d", rr.Code)
	}
}

func TestHandleUpdatePersonOK(t *testing.T) {
	api := &handlers.API{
		People: fakePeople{person: models.Person{Name: "Ana"}},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("PATCH /api/people/{id}", api.HandleUpdatePerson)

	req := httptest.NewRequest(http.MethodPatch, "/api/people/1", strings.NewReader(`{"weeklyHours":32}`))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rr.Code, rr.Body.String())
	}
	var person models.Person
	_ = json.NewDecoder(rr.Body).Decode(&person)
	if person.WeeklyHours != 32 || person.ID != 1 {
		t.Fatalf("unexpected person: %+v", person)
	}
}

func TestHandleUpdatePersonNotFound(t *testing.T) {
	api := &handlers.API{
		People: fakePeople{err: services.ErrNotFound},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("PATCH /api/people/{id}", api.HandleUpdatePerson)

	req := httptest.NewRequest(http.MethodPatch, "/api/people/999", strings.NewReader(`{"weeklyHours":32}`))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("status %d", rr.Code)
	}
}

func TestHandleUpdatePersonRejectsUnknownFields(t *testing.T) {
	api := &handlers.API{People: fakePeople{person: models.Person{Name: "Ana"}}}
	mux := http.NewServeMux()
	mux.HandleFunc("PATCH /api/people/{id}", api.HandleUpdatePerson)

	req := httptest.NewRequest(http.MethodPatch, "/api/people/1", strings.NewReader(`{"weeklyHours":32,"admin":true}`))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d", rr.Code)
	}
}

func TestHandleHealth(t *testing.T) {
	api := &handlers.API{Health: fakeHealth{resp: models.HealthResponse{OK: true, People: 500}}}
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rr := httptest.NewRecorder()
	api.HandleHealth(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d", rr.Code)
	}
}

func TestHandleCapacityServiceError(t *testing.T) {
	api := &handlers.API{Capacity: fakeCapacity{err: errors.New("db down")}}
	req := httptest.NewRequest(http.MethodGet, "/api/capacity?from=2025-12-29&to=2026-01-18", nil)
	rr := httptest.NewRecorder()
	api.HandleCapacity(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status %d", rr.Code)
	}
}
