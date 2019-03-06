// Christopher Black 2019
// an aedifex project
// ...for learning
// ...for science

package main

import (
	"encoding/json"
	"fmt"
	"io"
	// "io/ioutil"
	"log"
	"net/http"
	"os"
)

var build_id, build_time = "testing", "testing"

// this needs to be wrapped in the function handler
func logRequest(r *http.Request) {
	log.Printf("Request recieved from host: %v", r.Host)
	log.Printf("URI requested: %v", r.URL)
}

func version(w http.ResponseWriter, r *http.Request) {
	// mapD := map[string]int{"apple": 5, "lettuce": 7}
	version := map[string]string{"Build v": build_id, "Build time": build_time}
	payload, _ := jsonIfy(version)
	fmt.Fprintf(w, string(payload))
	// fmt.Fprintf(w, "Build version: %s\n", build_id)
	// fmt.Fprintf(w, "This build was compiled on: %s\n", build_time)
}

func hello(w http.ResponseWriter, r *http.Request) {
	log.Printf("Serving request: %s", r.URL.Path)
	log.Printf("URI requested: %v", r.URL)
	host, _ := os.Hostname()
	fmt.Fprintf(w, "Hello, world!\n")
	fmt.Fprintf(w, "Version: 1.0.0\n")
	fmt.Fprintf(w, "Hostname: %s\n", host)
	w.WriteHeader(http.StatusInternalServerError)
	// fmt.Fprintf(w, "Remote IP: ", r.RemoteAddr[6:])
	// fmt.Fprintf(w, "Remote IP: ", r.RemoteAddr[6:])
	// ip := map[string]interface{}{"Origin": r.RemoteAddr[6:], "Host": r.Host}

}

func test_json(w http.ResponseWriter, r *http.Request) {
	payload, _ := jsonIfy([]string{"apple", "peach", "pear"})
	fmt.Fprintf(w, string(payload))
}

// return 'get' URI
func get(w http.ResponseWriter, r *http.Request) {
	logRequest(r)
	// return GET data
	io.WriteString(w, fmt.Sprintf("Request method: %v\n", r.Method))
	io.WriteString(w, fmt.Sprintf("Request path: %v\n", r.URL.Path))
	io.WriteString(w, fmt.Sprintf("URL 	 path: %v\n", r.URL))
}

// this should write to the http req
func jsonIfy(element interface{}) ([]byte, error) {
	json, err := json.Marshal(element)
	if err != nil {
		return nil, err
	}
	return json, nil
}

func startServer() {
	// config port or default to 8000
	var port string
	if os.Getenv("PORT") != "" {
		port = ":" + os.Getenv("PORT")
	} else {
		port = ":8000"
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", hello)

	mux.HandleFunc("/version", version)

	mux.HandleFunc("/json", test_json)

	log.Printf("Starting server on port: %v", port)
	err := http.ListenAndServe(port, mux)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

	// defer resp.Body.Close()
}

func main() {
	// if no other params
	startServer()
}
