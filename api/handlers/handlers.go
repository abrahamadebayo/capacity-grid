package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"capacity/api/models"
	"capacity/api/security"
	"capacity/api/services"
)

func isCanceled(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

// CapacityReader loads capacity for a range.
type CapacityReader interface {
	Get(ctx context.Context, from, to time.Time) (models.CapacityResponse, error)
}

// PersonUpdater updates a person's weekly hours.
type PersonUpdater interface {
	UpdateWeeklyHours(ctx context.Context, id int, weeklyHours float64) (models.Person, error)
}

// HealthChecker reports API readiness.
type HealthChecker interface {
	Check(ctx context.Context) (models.HealthResponse, error)
}

// API groups HTTP handlers.
type API struct {
	Capacity CapacityReader
	People   PersonUpdater
	Health   HealthChecker
}

func (a *API) HandleCapacity(w http.ResponseWriter, r *http.Request) {
	from, to, err := security.ParseRange(r.URL.Query().Get("from"), r.URL.Query().Get("to"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := a.Capacity.Get(r.Context(), from, to)
	if err != nil {
		if isCanceled(err) {
			return
		}
		log.Printf("capacity: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to load capacity")
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (a *API) HandleUpdatePerson(w http.ResponseWriter, r *http.Request) {
	id, err := security.ParsePersonID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var body models.UpdatePersonRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	hours, err := security.ValidateUpdatePerson(body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	person, err := a.People.UpdateWeeklyHours(r.Context(), id, hours)
	if errors.Is(err, services.ErrNotFound) {
		writeError(w, http.StatusNotFound, "person not found")
		return
	}
	if err != nil {
		log.Printf("update person: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to update person")
		return
	}
	writeJSON(w, http.StatusOK, person)
}

func (a *API) HandleHealth(w http.ResponseWriter, r *http.Request) {
	resp, err := a.Health.Check(r.Context())
	if err != nil {
		log.Printf("health: %v", err)
		writeError(w, http.StatusInternalServerError, "unhealthy")
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
