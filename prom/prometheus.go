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

func StartPrometheusSender() {
	fmt.Printf("starting Prometheus sender with interval %d\n", conf.PushGatewayIntervalSecs)
	channel := time.Tick(time.Duration(conf.PushGatewayIntervalSecs) * time.Second)
	go func() {
		for range channel {
			if !conf.StopRequested {
				pushMetrics()
			}
		}
	}()
}

func pushMetrics() {
	conf.HzSliceLock.Lock()
	// determine the average, and the empty the slice again
	var total int64
	var averageGet, averagePut int64
	for _, value := range conf.HzGetTimes {
		total = total + value
	}
	averageGet = total / int64(len(conf.HzGetTimes))
	total = 0
	for _, value := range conf.HzPutTimes {
		total = total + value
	}
	averagePut = total / int64(len(conf.HzPutTimes))
	conf.HzGetTimes = conf.HzGetTimes[:0]
	conf.HzPutTimes = conf.HzPutTimes[:0]
	conf.HzSliceLock.Unlock()

	metricGet := prometheus.NewGauge(prometheus.GaugeOpts{Name: metricName, Help: metricHelp})
	metricPut := prometheus.NewGauge(prometheus.GaugeOpts{Name: metricName, Help: metricHelp})
	metricGet.Set(float64(averageGet))
	metricPut.Set(float64(averagePut))

	if err := push.New(conf.PushGatewayURL, "rabo_hzmon").Grouping("sourceIP", conf.MyIP).Grouping("operation", "get").Collector(metricGet).Push(); err != nil {
		log.Printf("failed to send get metrics to %s: %v\n", conf.PushGatewayURL, err.Error())
	} else {
		util.LogDebug(fmt.Sprintf("get metric (%d) sent to push gateway @ %s", averageGet, conf.PushGatewayURL))
	}
	if err := push.New(conf.PushGatewayURL, "rabo_hzmon").Grouping("sourceIP", conf.MyIP).Grouping("operation", "put").Collector(metricPut).Push(); err != nil {
		log.Printf("failed to send put metrics to %s: %v\n", conf.PushGatewayURL, err.Error())
	} else {
		util.LogDebug(fmt.Sprintf("put metric (%d) sent to push gateway @ %s", averagePut, conf.PushGatewayURL))
	}
}
