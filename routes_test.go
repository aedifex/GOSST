package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Tests in this file go through the real router (newMux) over a real HTTP
// connection, so they cover routing, static files and headers - not just
// the handler functions.

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(newMux())
	t.Cleanup(srv.Close)
	return srv
}

func fetch(t *testing.T, method, url string, headers map[string]string) (*http.Response, string) {
	t.Helper()
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return resp, string(body)
}

func TestRoutes(t *testing.T) {
	setBuildVars(t, "v-test", "abc1234", "master", "2026-11-18T00:00:00Z", "test")
	srv := newTestServer(t)

	cases := []struct {
		name        string
		path        string
		status      int
		contentType string // prefix match
	}{
		{"health", "/health", http.StatusOK, "text/plain"},
		{"faux", "/faux", http.StatusOK, "text/plain"},
		{"get", "/get", http.StatusOK, "text/plain"}, // NOTE: JSON body, no JSON Content-Type
		{"runtime", "/runtime", http.StatusOK, "text/plain"},
		{"whatismyip", "/whatismyip", http.StatusOK, "text/plain"},
		{"user-agent", "/user-agent", http.StatusOK, "text/plain"}, // NOTE: JSON body, no JSON Content-Type
		{"version", "/version", http.StatusOK, "application/json"},
		{"hostz", "/hostz", http.StatusOK, "application/json"},
		{"root", "/", http.StatusOK, "text/html"},
		{"unknown path", "/nope", http.StatusNotFound, "text/plain"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp, _ := fetch(t, http.MethodGet, srv.URL+tc.path, nil)
			if resp.StatusCode != tc.status {
				t.Errorf("status: want %d, got %d", tc.status, resp.StatusCode)
			}
			if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, tc.contentType) {
				t.Errorf("Content-Type: want prefix %q, got %q", tc.contentType, ct)
			}
		})
	}
}

func TestStaticSite(t *testing.T) {
	srv := newTestServer(t)

	t.Run("index", func(t *testing.T) {
		resp, body := fetch(t, http.MethodGet, srv.URL+"/", nil)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status: want 200, got %d", resp.StatusCode)
		}
		if !strings.Contains(body, "<title>Globex</title>") {
			t.Error("index.html: expected <title>Globex</title>")
		}
	})
	t.Run("asset", func(t *testing.T) {
		resp, _ := fetch(t, http.MethodGet, srv.URL+"/images/favicon.png", nil)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status: want 200, got %d", resp.StatusCode)
		}
		if ct := resp.Header.Get("Content-Type"); ct != "image/png" {
			t.Errorf("Content-Type: want image/png, got %q", ct)
		}
	})
	t.Run("missing asset", func(t *testing.T) {
		resp, _ := fetch(t, http.MethodGet, srv.URL+"/images/does-not-exist.png", nil)
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("status: want 404, got %d", resp.StatusCode)
		}
	})
}

// NOTE: handlers don't check the method, so every route answers any verb.
// This pins current behaviour; tighten it if you add method restrictions.
func TestAnyMethodAccepted(t *testing.T) {
	srv := newTestServer(t)

	cases := []struct{ method, path string }{
		{http.MethodPost, "/health"},
		{http.MethodDelete, "/faux"},
		{http.MethodPut, "/version"},
	}
	for _, tc := range cases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			resp, _ := fetch(t, tc.method, srv.URL+tc.path, nil)
			if resp.StatusCode != http.StatusOK {
				t.Errorf("status: want 200, got %d", resp.StatusCode)
			}
		})
	}
}

func TestEchoRoutesOverHTTP(t *testing.T) {
	srv := newTestServer(t)

	t.Run("whatismyip sees loopback", func(t *testing.T) {
		_, body := fetch(t, http.MethodGet, srv.URL+"/whatismyip", nil)
		if body != "127.0.0.1" {
			t.Errorf("body: want %q, got %q", "127.0.0.1", body)
		}
	})
	t.Run("user-agent echoes client header", func(t *testing.T) {
		_, body := fetch(t, http.MethodGet, srv.URL+"/user-agent",
			map[string]string{"User-Agent": "gosst-ci/1.0"})
		if want := `{"User agent":["gosst-ci/1.0"]}`; body != want {
			t.Errorf("body: want %s, got %s", want, body)
		}
	})
}

func TestServer_SlowishSmoke(t *testing.T) {
	srv := newTestServer(t)

	// Small loop + sleep so the test takes measurable time (~300ms) in CI.
	endpoints := []string{"/faux", "/health", "/runtime", "/version", "/hostz", "/"}
	for i := 0; i < 20; i++ {
		for _, ep := range endpoints {
			resp, _ := fetch(t, http.MethodGet, srv.URL+ep, nil)
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("%s: expected 200, got %d", ep, resp.StatusCode)
			}
		}
		time.Sleep(15 * time.Millisecond)
	}
}
