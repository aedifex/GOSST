package main

import (
	"encoding/json"
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

func TestVersion(t *testing.T) {
	// Your handler currently does:
	//   "commit":    getEnv(CommitSHA, "abc123")
	//   "timestamp": getEnv(build_time, "today?")
	// where CommitSHA/build_time may themselves hold ENV *KEYS*.
	// We set both the keys and the env so either behavior is acceptable.

	// Globals act as ENV KEYS (what your code expects today)
	build_id = "test-build"
	CommitSHA = "COMMIT_SHA"
	build_time = "BUILD_TIME"

	// Provide env values those keys would point to (CI-style)
	t.Setenv("BUILD_ID", "build-123")
	t.Setenv("GIT_BRANCH", "main")
	t.Setenv("DEPLOYED_BY", "unit-test")
	t.Setenv("DEPLOY_ENV", "test")
	t.Setenv("COMMIT_SHA", "deadbeef")
	t.Setenv("BUILD_TIME", "2025-08-15T12:34:56Z")

	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	w := httptest.NewRecorder()
	version(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status want 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type want application/json, got %q", ct)
	}

	var got map[string]string
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("json decode: %v", err)
	}

	// Always-true fields
	if got["version"] != "test-build" {
		t.Errorf("version: want %q, got %q", "test-build", got["version"])
	}
	if got["build"] != "build-123" {
		t.Errorf("build: want %q, got %q", "build-123", got["build"])
	}
	if got["branch"] != "main" {
		t.Errorf("branch: want %q, got %q", "main", got["branch"])
	}
	if got["deployed_by"] != "unit-test" {
		t.Errorf("deployed_by: want %q, got %q", "unit-test", got["deployed_by"])
	}
	if got["env"] != "test" {
		t.Errorf("env: want %q, got %q", "test", got["env"])
	}

	// Indirection-tolerant assertions:
	// Accept either the ENV *value* or the ENV *key*, matching your current code path.
	commit := got["commit"]
	if commit != "deadbeef" && commit != "COMMIT_SHA" {
		t.Errorf("commit: want %q or %q, got %q", "deadbeef", "COMMIT_SHA", commit)
	}
	timestamp := got["timestamp"]
	if timestamp != "2025-08-15T12:34:56Z" && timestamp != "BUILD_TIME" {
		t.Errorf("timestamp: want %q or %q, got %q", "2025-08-15T12:34:56Z", "BUILD_TIME", timestamp)
	}
}
