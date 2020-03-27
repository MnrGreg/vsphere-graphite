package backend

// InitPrometheus : Set some channels to notify other theads when using Prometheus
import (
	"fmt"
	"log"

	"github.com/prometheus/client_golang/prometheus"
)

// Describe : Implementation of Prometheus Collector.Describe
func (backend *Config) Describe(ch chan<- *prometheus.Desc) {
	prometheus.NewGauge(prometheus.GaugeOpts{Name: "Dummy", Help: "Dummy"}).Describe(ch)
}

// Collect : Implementation of Prometheus Collector.Collect
func (backend *Config) Collect(ch chan<- prometheus.Metric) {

	log.Println("prometheus: requesting metrics")

	request := make(chan Point, 100)
	channels := Channels{Request: &request}

	select {
	case *queries <- channels:
		log.Println("prometheus: requested metrics")
	default:
		log.Println("prometheus: query buffer full. discarding request")
		return
	}

	// points received
	points := 0
	for point := range *channels.Request {
		// increase points
		points++
		// send point to prometheus
		backend.PrometheusSend(ch, point)
	}
	log.Printf("prometheus: sent %d points", points)
}

//PrometheusSend sends a point to prometheus
func (backend *Config) PrometheusSend(ch chan<- prometheus.Metric, point Point) {
	tags := point.GetTags(backend.NoArray, ",")
	labelNames := make([]string, len(tags))
	labelValues := make([]string, len(tags))
	i := 0
	for key, value := range tags {
		labelNames[i] = key
		labelValues[i] = value
		i++
	}
	key := fmt.Sprintf("%s_%s_%s_%s", backend.Prefix, point.Group, point.Counter, point.Rollup)
	desc := prometheus.NewDesc(key, "vSphere collected metric", labelNames, nil)
	metric, err := prometheus.NewConstMetric(desc, prometheus.GaugeValue, float64(point.Value), labelValues...)
	if err != nil {
		log.Println("Error creating prometheus metric")
	}
	ch <- metric
}
