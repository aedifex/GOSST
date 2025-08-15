package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFaux(t *testing.T) {
	// Create a fake HTTP GET request to /faux
	req := httptest.NewRequest(http.MethodGet, "/faux", nil)

	// Response recorder to capture handler output
	w := httptest.NewRecorder()

	// Call the handler directly
	faux(w, req)

	// Check the status code
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Check the body content
	expected := "CI Demo!"
	if w.Body.String() != expected {
		t.Errorf("expected body %q, got %q", expected, w.Body.String())
	}
}
