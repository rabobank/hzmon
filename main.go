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

	fmt.Printf("Starting hzmon (commithash:%s, buildtime:%s) with a %d second interval using MyIP: %s\n", conf.CommitHash, conf.BuildTime, conf.IntervalSecs, conf.MyIP)

	startHttpServer()

	conf.MyIP = util.GetIP()

	// initial wait time so that with many instances the activity will be spread out
	waitTime := conf.CFInstanceIndex
	fmt.Printf("waiting %d seconds\n", waitTime)
	time.Sleep(time.Duration(waitTime) * time.Second)

	prom.StartPrometheusSender()

	hz.StartProbing()
}
