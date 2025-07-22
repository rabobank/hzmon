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

var metricGet prometheus.Gauge
var metricPut prometheus.Gauge
var pushGateway *push.Pusher

func StartPrometheusSender() {
	fmt.Printf("starting Prometheus sender with interval %d\n", conf.PushGatewayIntervalSecs)
	metricGet = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name:        metricName,
			Help:        metricHelp,
			ConstLabels: map[string]string{"operation": "get", "sourceIP": conf.MyIP, "instanceIndex": fmt.Sprintf("%d", conf.CFInstanceIndex)}},
	)
	metricPut = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name:        metricName,
			Help:        metricHelp,
			ConstLabels: map[string]string{"operation": "put", "sourceIP": conf.MyIP, "instanceIndex": fmt.Sprintf("%d", conf.CFInstanceIndex)}},
	)
	registry := prometheus.NewRegistry()
	getCollector := prometheus.Collector(metricGet)
	putCollector := prometheus.Collector(metricPut)
	registry.MustRegister(getCollector, putCollector)
	pushGateway = push.New(conf.PushGatewayURL, "rabo_hzmon").Collector(getCollector).Collector(putCollector)

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
	if len(conf.HzGetTimes) > 0 && len(conf.HzPutTimes) > 0 {
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

		metricGet.Set(float64(averageGet))
		metricPut.Set(float64(averagePut))

		if err := pushGateway.Push(); err != nil {
			log.Printf("failed to send metrics to %s: %v\n", conf.PushGatewayURL, err.Error())
		} else {
			util.LogDebug(fmt.Sprintf("metrics (get=%d,put=%d) sent to push gateway @ %s", averageGet, averagePut, conf.PushGatewayURL))
		}
	}
}
