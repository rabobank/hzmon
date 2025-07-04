package main

import (
	"fmt"
	"github.com/rabobank/hzmon/conf"
	"github.com/rabobank/hzmon/hz"
	"github.com/rabobank/hzmon/util"
	"os"
	"time"
)

func main() {
	if !conf.EnvironmentComplete() {
		os.Exit(8)
	}

	conf.MyIP = util.GetIP()
	fmt.Printf("Starting hzmon (commithash:%s, buildtime:%s) with a %d second interval using MyIP: %s\n", conf.CommitHash, conf.BuildTime, conf.IntervalSecs, conf.MyIP)

	startHttpServer()

	hz.StartProbing()

	for {
		if conf.StopRequested {
			fmt.Println("Stopping server")
			os.Exit(0)
		}
		time.Sleep(2 * time.Second)
	}
}
