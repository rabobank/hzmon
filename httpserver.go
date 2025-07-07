package main

import (
	"fmt"
	"github.com/rabobank/hzmon/conf"
	"net/http"
)

func startHttpServer() {
	go func() {
		http.HandleFunc("/", handleRequest)
		fmt.Println("Http server started on port 8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			panic(err)
		}
	}()
}

func handleRequest(writer http.ResponseWriter, request *http.Request) {
	_ = writer // prevent unused variable warning
	stop := request.URL.Query()["stop"]
	if len(stop) > 0 {
		fmt.Printf("stop probing requested by %s\n", request.RemoteAddr)
		conf.StopRequested = true
	}
	start := request.URL.Query()["start"]
	if len(start) > 0 {
		fmt.Printf("start probing requested by %s\n", request.RemoteAddr)
		conf.StopRequested = false
	}
	debugoff := request.URL.Query()["debugoff"]
	if len(debugoff) > 0 {
		fmt.Printf("debugoff requested by %s\n", request.RemoteAddr)
		conf.Debug = false
	}
	debugon := request.URL.Query()["debugon"]
	if len(debugon) > 0 {
		fmt.Printf("debugon requested by %s\n", request.RemoteAddr)
		conf.Debug = true
	}
	_, _ = fmt.Fprintf(writer, "hzmon running:%t, debug:%t.  Use /?start or /?stop to control probing, /?debugon or /?debugoff to control debug mode (use header X-Cf-App-Instance=<ix> to direct to all the app instances).\n", !conf.StopRequested, conf.Debug)
}
