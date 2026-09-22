package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"capacity/api/middleware"
)

func TestSecurityHeaders(t *testing.T) {
	h := middleware.SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatal("missing nosniff")
	}
	if rr.Header().Get("X-Frame-Options") != "DENY" {
		t.Fatal("missing frame options")
	}
	if rr.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("missing cache-control")
	}
}

func TestRecover(t *testing.T) {
	h := middleware.Recover(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status %d", rr.Code)
	}
}

func TestMaxBody(t *testing.T) {
	var readErr error
	h := middleware.MaxBody(8)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 64)
		_, readErr = r.Body.Read(buf)
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("0123456789abcdef"))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if readErr == nil {
		t.Fatal("expected error reading oversized body")
	}
}

func TestChainOrder(t *testing.T) {
	var order []string
	mw := func(name string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name+":in")
				next.ServeHTTP(w, r)
				order = append(order, name+":out")
			})
		}
	}
	h := middleware.Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "handler")
		}),
		mw("a"),
		mw("b"),
	)
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	want := "a:in b:in handler b:out a:out"
	got := strings.Join(order, " ")
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
