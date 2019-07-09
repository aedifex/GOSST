// Christopher Black 2019
// An aedifex project
// ...for learning
// ...for science

package main

import (
	"encoding/json"
	"fmt"
	"io"
	// "io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	// "errors"
)

var build_id = "dev"
var build_time = "dev"

// Takes an element, returns an array of bytes
// in json fmt.
func jsonIfy(element interface{}) ([]byte, error) {
	json, err := json.Marshal(element)
	if err != nil {
		return nil, err
	}
	return json, nil
}

// return 'get' URI
func get(w http.ResponseWriter, r *http.Request) {
	// logRequest(r)
	// return GET data

	resp := map[string]interface{}{"Request method": r.Method}
	payload, err := jsonIfy(resp)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		// I believe error terminates the program...
		log.Fatal(err)
	}
	//io.WriteString(w, fmt.Sprintf("Request method: %v\n", r.Method))
	//io.WriteString(w, fmt.Sprintf("Request path: %v\n", r.URL.Path))
	//io.WriteString(w, fmt.Sprintf("URL 	 path: %v\n", r.URL))
	fmt.Fprintf(w, string(payload))
}

func runtimeInfo(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Ave mundus! I'm running on %s with an %s CPU ", runtime.GOOS, runtime.GOARCH)
}

func user_agentHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{"User agent": r.Header["User-Agent"]}
	payload, _ := jsonIfy(resp)
	io.WriteString(w, string(payload))
}

func version(w http.ResponseWriter, r *http.Request) {
	version := map[string]string{"Build version": build_id, "Build time": build_time}
	payload, _ := jsonIfy(version)
	fmt.Fprintf(w, string(payload))
}

func checkErr(e error) error {
	if e != nil {
		return e
	}
	return nil
}

// return requester's ip address
func whatismyip(w http.ResponseWriter, r *http.Request) {
	ipAddress, _, err := net.SplitHostPort(r.RemoteAddr)
	checkErr(err)
	fmt.Fprintf(w, "%s", ipAddress)
}

func startServer() {

	// show at start time and when requested
	binary_version := map[string]string{"Build v": build_id, "Build time": build_time}

	// config port or default to 8000
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

	log.Printf("Starting server version: %v on port: %v", binary_version["Build v"], port)
	err := http.ListenAndServe(port, mux)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func main() {
	startServer()
}
