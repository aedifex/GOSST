// Christopher Black 2021
// An aedifex project
// ...for learning
// ...for science

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
)

// These values will be used to version the binary.
var build_id, build_time = "dev", "dev"
var git_commit string

// Used for basic health checks, returning a 200 if the app is up and running
http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
})

// Takes an element, returns an array of bytes in json fmt.
func jsonIfy(element interface{}) ([]byte, error) {
	json, err := json.Marshal(element)
	if err != nil {
		return nil, err
	}
	return json, nil
}

// Return 'get' URI in the body of the response.
func get(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{"Request method": r.Method}
	payload, err := jsonIfy(resp)
	checkErr(err)
	fmt.Fprintf(w, string(payload))
}

// Returns OS & ARCH info about the host.
func runtimeInfo(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Ave mundus! I'm running on %s with an %s CPU ", runtime.GOOS, runtime.GOARCH)
}

// Returns user agent associated with request.
func user_agent(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{"User agent": r.Header["User-Agent"]}
	payload, _ := jsonIfy(resp)
	io.WriteString(w, string(payload))
}

// Returns binary version in the form of SHA1 && compile time.
func version(w http.ResponseWriter, r *http.Request) {
	version := map[string]string{"Build version": build_id, "Build time": build_time, "GitRev": git_commit}
	payload, _ := jsonIfy(version)
	fmt.Fprintf(w, string(payload))
}

// Simple helper function. Some might find it useful...
// others pointless.
func checkErr(e error) error {
	if e != nil {
		return e
	}
	return nil
}

// Return requester's IP address.
func whatismyip(w http.ResponseWriter, r *http.Request) {
	ipAddress, _, err := net.SplitHostPort(r.RemoteAddr)
	checkErr(err)
	fmt.Fprintf(w, "%s", ipAddress)
}

func startServer() {
	// PORT is a good example of elements we'd like to be able to configure.
	var port string
	if os.Getenv("PORT") != "" {
		port = ":" + os.Getenv("PORT")
	} else {
		port = ":8000"
	}

	mux := http.NewServeMux()

	// *** Multiplexer && Routes ***
	// Basically mux is a mapping of a string in the form of a
	// request URL to a function that takes Response/Request Writers.
	// Functions with a larger/more complex scope will be defined in routes.go
	mux.HandleFunc("/whatismyip", whatismyip)
	mux.HandleFunc("/get", get)
	mux.HandleFunc("/runtime", runtimeInfo)
	mux.HandleFunc("/version", version)
	mux.HandleFunc("/user-agent", user_agent)
	// mux.HandleFunc("/health", user_agent)

	// Serve static html
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.FileServer(http.Dir("./static")).ServeHTTP(w, r)
	})

	log.Printf("Starting server version: %v on port: %v", build_id, port)
	err := http.ListenAndServe(port, mux)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func main() {
	startServer()
}
