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
	parm1 := request.URL.Query()["parm1"]
	_, _ = writer.Write([]byte(fmt.Sprintf("parm1: %s\n", parm1)))
	stop := request.URL.Query()["stop"]
	if len(stop) > 0 {
		fmt.Printf("Stop server requested by %s\n", request.RemoteAddr)
		conf.StopRequested = true
	}
}
