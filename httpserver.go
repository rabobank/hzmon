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
	stop := request.URL.Query()["stop"]
	if len(stop) > 0 {
		fmt.Printf("Stop server requested by %s\n", request.RemoteAddr)
		conf.StopRequested = true
	}
	start := request.URL.Query()["start"]
	if len(start) > 0 {
		fmt.Printf("start server requested by %s\n", request.RemoteAddr)
		conf.StopRequested = false
	}
}
