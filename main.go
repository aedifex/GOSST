// Christopher Black 2021
// An aedifex project
// ...for learning
// ...for science

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"
)

// These values will be used to version the binary.
// var build_id, build_time = "dev", "dev"
var BuildTime string
var BuildID string
var CommitSHA string
var GitBranch string
var DeployedBy string

// Stamped when the process boots so we can report uptime.
var startTime = time.Now()

// Takes an element, returns an array of bytes in JSON format.
func jsonIfy(element interface{}) ([]byte, error) {
	json, err := json.Marshal(element)
	if err != nil {
		return nil, err
	}
	return json, nil
}

// writeJSON sends v as a 200 application/json response, or a 500 if it can't
// be marshalled. Every JSON route goes through here so they all behave the same.
// The payload is written as raw bytes - never passed to Fprintf as a format
// string - so a '%' in the data can't be misread as a formatting verb.
func writeJSON(w http.ResponseWriter, v interface{}) {
	payload, err := jsonIfy(v)
	if err != nil {
		http.Error(w, "error generating JSON payload", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(payload)
}

// Used for basic health checks, returning a 200 if the app is up and running.
func health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(CommitSHA))
}

// Canary route: if this answers, the pipeline actually shipped.
func faux(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("CI Demo!"))
}

// Echoes the request back, httpbin-style: method, path, query args, headers
// and the caller's IP. Registered for every method - echoing it is the point.
func get(w http.ResponseWriter, r *http.Request) {
	headers := make(map[string]string, len(r.Header))
	for name, values := range r.Header {
		headers[name] = strings.Join(values, ", ")
	}
	origin, _ := clientIP(r) // best effort: empty if it can't be determined

	writeJSON(w, map[string]interface{}{
		"method":  r.Method,
		"path":    r.URL.Path,
		"args":    r.URL.Query(),
		"headers": headers,
		"origin":  origin,
	})
}

// Returns OS & ARCH info about the host.
func runtimeInfo(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Ave mundus! I'm running on %s with an %s CPU ", runtime.GOOS, runtime.GOARCH)
}

// Returns identity & vitals for the host that served this request.
// Handy behind a load balancer: hit it a few times and watch the
// hostname change as you're bounced between replicas.
func hostz(w http.ResponseWriter, r *http.Request) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	writeJSON(w, map[string]interface{}{
		"hostname":   hostname,
		"pid":        os.Getpid(),
		"uptime":     time.Since(startTime).Round(time.Second).String(),
		"os":         runtime.GOOS,
		"arch":       runtime.GOARCH,
		"go_version": runtime.Version(),
		"num_cpu":    runtime.NumCPU(),
		"gomaxprocs": runtime.GOMAXPROCS(0),
		"goroutines": runtime.NumGoroutine(),
	})
}

// Returns the request's User-Agent as a plain string ("" if none was sent).
func user_agent(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{"user_agent": r.UserAgent()})
}

// Helper function to get env vars and set a default fallback value
func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Printf("WARN: %s not set, using fallback '%s'\n", key, fallback)
		return fallback
	}
	return val
}

// runtimeEnv is build metadata supplied by whatever runs the binary, rather
// than stamped into it. It's read once when the router is built - not on every
// request, which also used to log a WARN per hit when the vars were unset.
type runtimeEnv struct {
	BuildID   string
	DeployEnv string
}

func loadRuntimeEnv() runtimeEnv {
	return runtimeEnv{
		BuildID:   getEnv("BUILD_ID", "local-build"),
		DeployEnv: getEnv("DEPLOY_ENV", "dev"),
	}
}

// Returns build identity. Two IDs, two sources:
//   - "version" is stamped at compile time via -ldflags (CI sets circle-<build #>)
//   - "build" is the BUILD_ID env var at runtime (set by whatever runs the container)
func versionHandler(env runtimeEnv) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]string{
			"version":     BuildID,
			"build":       env.BuildID,
			"commit":      CommitSHA,
			"branch":      GitBranch,
			"timestamp":   BuildTime,
			"deployed_by": DeployedBy,
			"env":         env.DeployEnv,
			"go_version":  runtime.Version(),
		})
	}
}

// clientIP works out the caller's IP. Behind a proxy or load balancer (e.g.
// Azure Container Apps ingress) the TCP connection comes from the proxy, so the
// left-most X-Forwarded-For entry wins when it's a valid IP. XFF is set by the
// client side and can be spoofed - fine for a demo, never use it for auth.
func clientIP(r *http.Request) (string, error) {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		first := strings.TrimSpace(strings.Split(xff, ",")[0])
		if ip := net.ParseIP(first); ip != nil {
			return ip.String(), nil
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", err
	}
	return host, nil
}

// Return requester's IP address.
func whatismyip(w http.ResponseWriter, r *http.Request) {
	ip, err := clientIP(r)
	if err != nil {
		http.Error(w, "could not determine client IP", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	io.WriteString(w, ip)
}

// allowGET restricts h to GET and HEAD, answering anything else with a 405.
func allowGET(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		h.ServeHTTP(w, r)
	})
}

// newMux builds the router with every route the service exposes.
// Shared by startServer and the tests so the two can never drift apart.
//
// Go 1.22+ method patterns: "GET /x" also answers HEAD. Other methods fall
// through to the "/" catch-all, which allowGET turns into a 405.
func newMux() *http.ServeMux {
	mux := http.NewServeMux()

	// *** Multiplexer && Routes ***
	mux.HandleFunc("GET /whatismyip", whatismyip)
	mux.HandleFunc("/get", get) // any method, on purpose: it echoes the method back
	mux.HandleFunc("GET /runtime", runtimeInfo)
	mux.HandleFunc("GET /hostz", hostz)
	mux.HandleFunc("GET /version", versionHandler(loadRuntimeEnv()))
	mux.HandleFunc("GET /user-agent", user_agent)
	mux.HandleFunc("GET /health", health)
	mux.HandleFunc("GET /faux", faux)

	// Serve static html
	//mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	//	http.ServeFile(w, r, "./static/index.html")
	//})
	// Plain "/" rather than "GET /": the router rejects "GET /" next to the
	// any-method "/get" as ambiguous, so allowGET enforces the method instead.
	mux.Handle("/", allowGET(http.FileServer(http.Dir("./static"))))

	return mux
}

// newServer wraps the router in an http.Server with timeouts. The old
// http.ListenAndServe had none, so a slow or stalled client could hold a
// connection (and a goroutine) open forever.
func newServer(addr string) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           newMux(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

// runServer serves on ln until ctx is cancelled, then stops accepting new
// connections and gives in-flight requests up to `drain` to finish. Container
// platforms send SIGTERM before killing a replica, so this is what keeps a
// rolling deploy from dropping requests mid-flight.
func runServer(ctx context.Context, srv *http.Server, ln net.Listener, drain time.Duration) error {
	errCh := make(chan error, 1)
	go func() { errCh <- srv.Serve(ln) }()

	select {
	case err := <-errCh:
		return err // the server died on its own
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), drain)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

func startServer() {
	// PORT is a good example of elements we'd like to be able to configure.
	var port string
	if os.Getenv("PORT") != "" {
		port = ":" + os.Getenv("PORT")
	} else {
		port = ":8000"
	}

	srv := newServer(port)
	ln, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal("Listen: ", err)
	}

	// Ctrl-C locally, SIGTERM from Docker / Kubernetes / Container Apps.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("Starting server version: %v on port: %v", BuildID, port)
	if err := runServer(ctx, srv, ln, 10*time.Second); err != nil {
		log.Fatal("Server: ", err)
	}
	log.Println("Server stopped cleanly")
}

func main() {
	startServer()
}
