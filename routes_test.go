package main

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Tests in this file go through the real router (newMux) or server over a
// real HTTP connection, so they cover routing, methods, static files, headers
// and shutdown - not just the handler functions.

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
		{"get", "/get", http.StatusOK, "application/json"},
		{"runtime", "/runtime", http.StatusOK, "text/plain"},
		{"whatismyip", "/whatismyip", http.StatusOK, "text/plain"},
		{"user-agent", "/user-agent", http.StatusOK, "application/json"},
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

// Routes are GET-only (Go 1.22 "GET /x" patterns plus allowGET on the static
// catch-all), so HEAD works and anything else is a 405 - except /get, which
// takes any method.
func TestMethodRouting(t *testing.T) {
	srv := newTestServer(t)

	cases := []struct {
		method, path string
		status       int
	}{
		{http.MethodPost, "/health", http.StatusMethodNotAllowed},
		{http.MethodDelete, "/faux", http.StatusMethodNotAllowed},
		{http.MethodPut, "/version", http.StatusMethodNotAllowed},
		{http.MethodPost, "/", http.StatusMethodNotAllowed},
		{http.MethodHead, "/health", http.StatusOK},
		{http.MethodHead, "/", http.StatusOK},
		{http.MethodPost, "/get", http.StatusOK},
		{http.MethodDelete, "/get", http.StatusOK},
	}
	for _, tc := range cases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			resp, body := fetch(t, tc.method, srv.URL+tc.path, nil)
			if resp.StatusCode != tc.status {
				t.Fatalf("status: want %d, got %d", tc.status, resp.StatusCode)
			}
			if tc.status == http.StatusMethodNotAllowed {
				if allow := resp.Header.Get("Allow"); !strings.Contains(allow, "GET") {
					t.Errorf("405 should advertise GET in Allow, got %q", allow)
				}
			}
			if tc.method == http.MethodHead && body != "" {
				t.Errorf("HEAD must not return a body, got %d bytes", len(body))
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
	t.Run("whatismyip honours X-Forwarded-For", func(t *testing.T) {
		_, body := fetch(t, http.MethodGet, srv.URL+"/whatismyip",
			map[string]string{"X-Forwarded-For": "198.51.100.9"})
		if body != "198.51.100.9" {
			t.Errorf("body: want %q, got %q", "198.51.100.9", body)
		}
	})
	t.Run("user-agent echoes client header", func(t *testing.T) {
		_, body := fetch(t, http.MethodGet, srv.URL+"/user-agent",
			map[string]string{"User-Agent": "gosst-ci/1.0"})
		if want := `{"user_agent":"gosst-ci/1.0"}`; body != want {
			t.Errorf("body: want %s, got %s", want, body)
		}
	})
	t.Run("get echoes query args", func(t *testing.T) {
		_, body := fetch(t, http.MethodGet, srv.URL+"/get?demo=1", nil)
		if !strings.Contains(body, `"args":{"demo":["1"]}`) {
			t.Errorf("want args to contain demo=1, got %s", body)
		}
	})
}

// ---------------------------------------------------------------------------
// Server: timeouts + graceful shutdown
// ---------------------------------------------------------------------------

func TestServerTimeouts(t *testing.T) {
	srv := newServer(":0")
	for name, d := range map[string]time.Duration{
		"ReadHeaderTimeout": srv.ReadHeaderTimeout,
		"ReadTimeout":       srv.ReadTimeout,
		"WriteTimeout":      srv.WriteTimeout,
		"IdleTimeout":       srv.IdleTimeout,
	} {
		if d <= 0 {
			t.Errorf("%s must be set, got %v", name, d)
		}
	}
	if srv.Handler == nil {
		t.Error("Handler must be the router, got nil")
	}
}

// Cancelling the context (SIGTERM in production) must let an in-flight
// request finish before runServer returns.
func TestGracefulShutdown(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	started := make(chan struct{})
	srv := newServer("")
	srv.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		time.Sleep(300 * time.Millisecond) // still working when shutdown begins
		io.WriteString(w, "done")
	})

	ctx, cancel := context.WithCancel(context.Background())
	runErr := make(chan error, 1)
	go func() { runErr <- runServer(ctx, srv, ln, 5*time.Second) }()

	type result struct {
		status int
		body   string
		err    error
	}
	reqDone := make(chan result, 1)
	go func() {
		resp, err := http.Get("http://" + ln.Addr().String() + "/slow")
		if err != nil {
			reqDone <- result{err: err}
			return
		}
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		reqDone <- result{status: resp.StatusCode, body: string(b)}
	}()

	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("request never reached the handler")
	}
	cancel() // simulate SIGTERM mid-request

	r := <-reqDone
	if r.err != nil {
		t.Fatalf("in-flight request was dropped: %v", r.err)
	}
	if r.status != http.StatusOK || r.body != "done" {
		t.Errorf("in-flight request: want 200 %q, got %d %q", "done", r.status, r.body)
	}
	if err := <-runErr; err != nil {
		t.Errorf("runServer: want nil after clean shutdown, got %v", err)
	}
}

func TestServer_SlowishSmoke(t *testing.T) {
	srv := newTestServer(t)

	// Small loop + sleep so the test takes measurable time (~300ms) in CI.
	endpoints := []string{"/faux", "/health", "/runtime", "/version", "/hostz", "/get", "/"}
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
