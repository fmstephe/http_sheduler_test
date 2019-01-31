package main

import (
	"flag"
	"fmt"
	"html"
	"log"
	"net/http"
	"time"
)

var (
	delayFlag = flag.Duration("d", time.Duration(0), "The amount of time to pause before sending the response")
)

func main() {
	flag.Parse()
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleRequest)
	log.Fatal(http.ListenAndServe(":9001", mux))
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	if *delayFlag != 0 {
		time.Sleep(*delayFlag)
	}
	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
		taken := time.Since(start)
		fmt.Printf("Duration: %d\n", int64(taken))
	}
}
