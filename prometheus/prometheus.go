package prom

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/rabobank/hzmon/conf"
	"github.com/rabobank/hzmon/util"
	"log"
	"time"
)

const (
	metricName = "rabo_hzmon_responsetime"
	metricHelp = "Response time of Hazelcast operations in microseconds"
)

var (
	metric = prometheus.NewGauge(prometheus.GaugeOpts{Name: metricName, Help: metricHelp})
)

func StartPrometheusSender() {
	channel := time.Tick(time.Duration(conf.PushGatewayIntervalSecs) * time.Second)
	go func() {
		for range channel {
			pushMetrics()
		}
	}()
}

func pushMetrics() {
	metric.Set(float64(conf.HzGetTime))
	pusher := push.New(conf.PushGatewayURL, "rabo_hzmon").Grouping("sourceIP", conf.MyIP).Grouping("operation", "get").Collector(metric)
	if err := pusher.Push(); err != nil {
		log.Printf("failed to send metrics to %s: %v\n", conf.PushGatewayURL, err.Error())
	} else {
		util.LogDebug(fmt.Sprintf("metrics sent to push gateway @ %s: %v", conf.PushGatewayURL, metric))
	}
}
