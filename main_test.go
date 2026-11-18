package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"runtime"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// setBuildVars sets the ldflags-stamped build globals for one test and
// restores the previous values when the test finishes, so tests can't leak
// state into each other.
func setBuildVars(t *testing.T, buildID, commit, branch, buildTime, deployedBy string) {
	t.Helper()
	oldID, oldCommit, oldBranch, oldTime, oldBy := BuildID, CommitSHA, GitBranch, BuildTime, DeployedBy
	t.Cleanup(func() {
		BuildID, CommitSHA, GitBranch, BuildTime, DeployedBy = oldID, oldCommit, oldBranch, oldTime, oldBy
	})
	BuildID, CommitSHA, GitBranch, BuildTime, DeployedBy = buildID, commit, branch, buildTime, deployedBy
}

// serve calls a handler directly and returns the recorded response.
func serve(h http.HandlerFunc, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	h(w, req)
	return w
}

func decodeJSON(t *testing.T, w *httptest.ResponseRecorder, v interface{}) {
	t.Helper()
	if err := json.Unmarshal(w.Body.Bytes(), v); err != nil {
		t.Fatalf("body is not valid JSON: %v\nbody: %s", err, w.Body.String())
	}
}

func wantStatus(t *testing.T, w *httptest.ResponseRecorder, want int) {
	t.Helper()
	if w.Code != want {
		t.Fatalf("status: want %d, got %d", want, w.Code)
	}
}

// ---------------------------------------------------------------------------
// /health
// ---------------------------------------------------------------------------

func TestHealth(t *testing.T) {
	cases := []struct {
		name string
		sha  string
	}{
		{"stamped commit", "abc1234"},
		{"unstamped build", ""}, // `go run .` - no ldflags, empty body
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			setBuildVars(t, "", tc.sha, "", "", "")
			w := serve(health, httptest.NewRequest(http.MethodGet, "/health", nil))

			wantStatus(t, w, http.StatusOK)
			if got := w.Body.String(); got != tc.sha {
				t.Errorf("body: want %q, got %q", tc.sha, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// /faux
// ---------------------------------------------------------------------------

func TestFaux(t *testing.T) {
	w := serve(faux, httptest.NewRequest(http.MethodGet, "/faux", nil))

	wantStatus(t, w, http.StatusOK)
	if got, want := w.Body.String(), "CI Demo!"; got != want {
		t.Errorf("body: want %q, got %q", want, got)
	}
}

// ---------------------------------------------------------------------------
// /get
// ---------------------------------------------------------------------------

func TestGet(t *testing.T) {
	methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			w := serve(get, httptest.NewRequest(method, "/get", nil))

			wantStatus(t, w, http.StatusOK)
			var got map[string]string
			decodeJSON(t, w, &got)
			want := map[string]string{"Request method": method}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("body: want %v, got %v", want, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// /runtime
// ---------------------------------------------------------------------------

func TestRuntimeInfo(t *testing.T) {
	w := serve(runtimeInfo, httptest.NewRequest(http.MethodGet, "/runtime", nil))

	wantStatus(t, w, http.StatusOK)
	want := fmt.Sprintf("Ave mundus! I'm running on %s with an %s CPU ", runtime.GOOS, runtime.GOARCH)
	if got := w.Body.String(); got != want {
		t.Errorf("body: want %q, got %q", want, got)
	}
}

// ---------------------------------------------------------------------------
// /whatismyip
// ---------------------------------------------------------------------------

func TestWhatIsMyIP(t *testing.T) {
	cases := []struct {
		name       string
		remoteAddr string
		xff        string
		want       string
	}{
		{"ipv4", "203.0.113.7:51234", "", "203.0.113.7"},
		{"ipv6", "[2001:db8::1]:443", "", "2001:db8::1"},
		{"ipv6 loopback", "[::1]:8080", "", "::1"},
		// NOTE: X-Forwarded-For is ignored. Behind a load balancer / ACA
		// ingress this reports the proxy's IP, not the caller's.
		{"ignores X-Forwarded-For", "10.0.0.5:4000", "198.51.100.9", "10.0.0.5"},
		// NOTE: the SplitHostPort error is swallowed by checkErr, so a bad
		// address yields 200 with an empty body rather than an error.
		{"unparseable address", "not-an-address", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/whatismyip", nil)
			req.RemoteAddr = tc.remoteAddr
			if tc.xff != "" {
				req.Header.Set("X-Forwarded-For", tc.xff)
			}
			w := serve(whatismyip, req)

			wantStatus(t, w, http.StatusOK)
			if got := w.Body.String(); got != tc.want {
				t.Errorf("body: want %q, got %q", tc.want, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// /user-agent
// ---------------------------------------------------------------------------

func TestUserAgent(t *testing.T) {
	// NOTE: the value is r.Header["User-Agent"], a []string - so the JSON
	// is an array (or null), not a plain string.
	cases := []struct {
		name string
		uas  []string
		want string
	}{
		{"single", []string{"gosst-test/1.0"}, `{"User agent":["gosst-test/1.0"]}`},
		{"missing", nil, `{"User agent":null}`},
		{"repeated header", []string{"a/1", "b/2"}, `{"User agent":["a/1","b/2"]}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/user-agent", nil)
			for _, ua := range tc.uas {
				req.Header.Add("User-Agent", ua)
			}
			w := serve(user_agent, req)

			wantStatus(t, w, http.StatusOK)
			if got := w.Body.String(); got != tc.want {
				t.Errorf("body: want %s, got %s", tc.want, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// /version
// ---------------------------------------------------------------------------

func TestVersion(t *testing.T) {
	cases := []struct {
		name                                         string
		buildID, commit, branch, buildTime, deployer string
		envBuildID, envDeployEnv                     string
		want                                         map[string]string
	}{
		{
			name:    "stamped build with env",
			buildID: "v1.2.3", commit: "deadbee", branch: "master",
			buildTime: "2026-11-18T12:00:00Z", deployer: "circleci",
			envBuildID: "circle-42", envDeployEnv: "test",
			want: map[string]string{
				"version": "v1.2.3", "build": "circle-42", "commit": "deadbee",
				"branch": "master", "timestamp": "2026-11-18T12:00:00Z",
				"deployed_by": "circleci", "env": "test",
			},
		},
		{
			name:    "env fallbacks",
			buildID: "v1.2.3", commit: "deadbee", branch: "master",
			buildTime: "2026-11-18T12:00:00Z", deployer: "circleci",
			want: map[string]string{
				"version": "v1.2.3", "build": "local-build", "commit": "deadbee",
				"branch": "master", "timestamp": "2026-11-18T12:00:00Z",
				"deployed_by": "circleci", "env": "dev",
			},
		},
		{
			name: "unstamped binary", // `go run .` - every ldflags var is empty
			want: map[string]string{
				"version": "", "build": "local-build", "commit": "",
				"branch": "", "timestamp": "", "deployed_by": "", "env": "dev",
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			setBuildVars(t, tc.buildID, tc.commit, tc.branch, tc.buildTime, tc.deployer)
			t.Setenv("BUILD_ID", tc.envBuildID)     // "" => getEnv falls back
			t.Setenv("DEPLOY_ENV", tc.envDeployEnv) // "" => getEnv falls back

			w := serve(version, httptest.NewRequest(http.MethodGet, "/version", nil))

			wantStatus(t, w, http.StatusOK)
			if ct := w.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type: want application/json, got %q", ct)
			}
			var got map[string]string
			decodeJSON(t, w, &got)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("payload mismatch\nwant: %v\n got: %v", tc.want, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// /hostz
// ---------------------------------------------------------------------------

func TestHostz(t *testing.T) {
	w := serve(hostz, httptest.NewRequest(http.MethodGet, "/hostz", nil))

	wantStatus(t, w, http.StatusOK)
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type: want application/json, got %q", ct)
	}

	// Mixed value types in the payload, so decode loosely.
	var got map[string]interface{}
	decodeJSON(t, w, &got)

	if len(got) != 8 {
		t.Errorf("want 8 fields, got %d: %v", len(got), got)
	}

	// Strings we can compare exactly against the running process.
	if want, err := os.Hostname(); err == nil {
		if got["hostname"] != want {
			t.Errorf("hostname: want %q, got %v", want, got["hostname"])
		}
	}
	if got["os"] != runtime.GOOS {
		t.Errorf("os: want %q, got %v", runtime.GOOS, got["os"])
	}
	if got["arch"] != runtime.GOARCH {
		t.Errorf("arch: want %q, got %v", runtime.GOARCH, got["arch"])
	}

	// JSON numbers decode as float64, so convert before comparing.
	exact := []struct {
		key  string
		want int
	}{
		{"pid", os.Getpid()},
		{"num_cpu", runtime.NumCPU()},
	}
	for _, tc := range exact {
		n, ok := got[tc.key].(float64)
		if !ok {
			t.Errorf("%s: want a number, got %T", tc.key, got[tc.key])
			continue
		}
		if int(n) != tc.want {
			t.Errorf("%s: want %d, got %d", tc.key, tc.want, int(n))
		}
	}

	// These move around at runtime, so only assert they're sane.
	for _, key := range []string{"gomaxprocs", "goroutines"} {
		if n, ok := got[key].(float64); !ok || n < 1 {
			t.Errorf("%s: want >= 1, got %v", key, got[key])
		}
	}

	// Uptime is a rounded Duration string ("0s", "1m30s") - must parse back.
	uptime, ok := got["uptime"].(string)
	if !ok {
		t.Fatalf("uptime: want a string, got %T", got["uptime"])
	}
	if _, err := time.ParseDuration(uptime); err != nil {
		t.Errorf("uptime %q is not a valid duration: %v", uptime, err)
	}
}

func TestHostzUptime(t *testing.T) {
	old := startTime
	t.Cleanup(func() { startTime = old })
	startTime = time.Now().Add(-90 * time.Second)

	w := serve(hostz, httptest.NewRequest(http.MethodGet, "/hostz", nil))

	var got map[string]interface{}
	decodeJSON(t, w, &got)
	if got["uptime"] != "1m30s" {
		t.Errorf("uptime: want %q, got %v", "1m30s", got["uptime"])
	}
}

// ---------------------------------------------------------------------------
// Helpers in main.go
// ---------------------------------------------------------------------------

func TestGetEnv(t *testing.T) {
	const key = "GOSST_TEST_GETENV"

	t.Run("set", func(t *testing.T) {
		t.Setenv(key, "value")
		if got := getEnv(key, "fallback"); got != "value" {
			t.Errorf("want %q, got %q", "value", got)
		}
	})
	t.Run("empty uses fallback", func(t *testing.T) {
		t.Setenv(key, "")
		if got := getEnv(key, "fallback"); got != "fallback" {
			t.Errorf("want %q, got %q", "fallback", got)
		}
	})
	t.Run("unset uses fallback", func(t *testing.T) {
		t.Setenv(key, "") // registers restore of any pre-existing value
		os.Unsetenv(key)
		if got := getEnv(key, "fallback"); got != "fallback" {
			t.Errorf("want %q, got %q", "fallback", got)
		}
	})
}

func TestJsonIfy(t *testing.T) {
	t.Run("marshals", func(t *testing.T) {
		got, err := jsonIfy(map[string]int{"a": 1})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(got) != `{"a":1}` {
			t.Errorf("want %s, got %s", `{"a":1}`, got)
		}
	})
	t.Run("unmarshalable value errors", func(t *testing.T) {
		got, err := jsonIfy(make(chan int))
		if err == nil {
			t.Fatal("want an error for a channel, got nil")
		}
		if got != nil {
			t.Errorf("want nil bytes on error, got %s", got)
		}
	})
}

func TestCheckErr(t *testing.T) {
	// NOTE: checkErr just hands back its input - it doesn't log, panic or
	// stop the caller, so handlers that rely on it silently ignore errors.
	if checkErr(nil) != nil {
		t.Error("want nil for nil input")
	}
	boom := errors.New("boom")
	if checkErr(boom) != boom {
		t.Error("want the same error back")
	}
}
