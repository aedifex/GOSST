// Christopher Black 2019
// an aedifex project
// ...for learning
// ...for science

package main

import (
	// "encoding/json"
	"fmt"
	// "io"
	// "io/ioutil"
	"log"
	"net/http"
	"os"
)

var build_id, build_time = "testing", "testing"

func logRequest(r *http.Request) {
	log.Printf("Request recieved from host: %v", r.Host)
	log.Printf("URI requested: %v", r.URL)
}

func version(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Build version: %s\n", build_id)
	fmt.Fprintf(w, "This build was compiled on: %s\n", build_time)
}

func hello(w http.ResponseWriter, r *http.Request) {
	log.Printf("Serving request: %s", r.URL.Path)
	log.Printf("URI requested: %v", r.URL)
	host, _ := os.Hostname()
	fmt.Fprintf(w, "Hello, world!\n")
	fmt.Fprintf(w, "Version: 1.0.0\n")
	fmt.Fprintf(w, "Hostname: %s\n", host)
	// fmt.Fprintf(w, "Remote IP: ", r.RemoteAddr[6:])
	// fmt.Fprintf(w, "Remote IP: ", r.RemoteAddr[6:])
	// ip := map[string]interface{}{"Origin": r.RemoteAddr[6:], "Host": r.Host}

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

	log.Printf("Starting server on port: %v", port)
	err := http.ListenAndServe(port, mux)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func main() {
	// if no other params
	startServer()
}
