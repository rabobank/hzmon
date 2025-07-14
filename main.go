package main

import (
	"fmt"
	"github.com/rabobank/hzmon/conf"
	"github.com/rabobank/hzmon/hz"
	"github.com/rabobank/hzmon/prom"
	"github.com/rabobank/hzmon/util"
	"os"
	"time"
)

func main() {
	if !conf.EnvironmentComplete() {
		os.Exit(8)
	}

	conf.MyIP = util.GetIP()

	fmt.Printf("Starting hzmon (version %s) with a %d second interval using MyIP: %s\n", conf.VERSION, conf.IntervalSecs, conf.MyIP)

	startHttpServer()

	// initial wait time so that with many instances the activity will be spread out
	waitTime := conf.CFInstanceIndex
	fmt.Printf("waiting %d seconds\n", waitTime)
	time.Sleep(time.Duration(waitTime) * time.Second)

	prom.StartPrometheusSender()

	hz.StartProbing()
}
