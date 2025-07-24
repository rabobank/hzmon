package prom

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rabobank/hzmon/hz"
	"net/http"
)

const (
	metricName = "rabo_hzmon_responsetime"
	metricHelp = "Response time of Hazelcast operations in microseconds"
)

type HZMonCollector struct {
	HzMetric *prometheus.Desc
}

func StartExporter() {
	//HzMetrics["0/get"] = HzMetric{LastUpdated: time.Now(), Operation: "get", SourceIP: "1.2.3.4", InstanceIndex: "0", RespTime: 1234}
	//HzMetrics["0/put"] = HzMetric{LastUpdated: time.Now(), Operation: "put", SourceIP: "1.2.4.5", InstanceIndex: "1", RespTime: 1245}
	fmt.Println("starting Prometheus exporter...")
	reg := prometheus.NewRegistry()
	reg.MustRegister(newHZMonCollector())
	handler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{EnableOpenMetrics: false})
	http.Handle("/metrics", handler)
}

func newHZMonCollector() *HZMonCollector {
	return &HZMonCollector{HzMetric: prometheus.NewDesc(metricName, metricHelp, []string{"operation", "sourceIP", "instanceIndex"}, nil)}
}

// Describe - Each and every collector must implement the Describe function. It essentially writes all descriptors to the exporter desc channel.
func (collector *HZMonCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.HzMetric
}

func (collector *HZMonCollector) Collect(ch chan<- prometheus.Metric) {
	if currentMetrics, err := hz.GetMetricsFromHZ(); err != nil {
		fmt.Printf("failed to get metrics from Hazelcast: %s\n", err)
		return
	} else {
		// iterate over the measurements and send them to the channel
		fmt.Printf("Collecting %d HzMetrics\n", len(currentMetrics))
		for _, metric := range currentMetrics {
			ch <- prometheus.MustNewConstMetric(collector.HzMetric, prometheus.GaugeValue, float64(metric.RespTime), metric.Operation, metric.SourceIP, metric.InstanceIndex)
		}
	}
}
