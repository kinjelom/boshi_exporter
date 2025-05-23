package collectors

import (
	"boshi_exporter/config"
	"boshi_exporter/fetchers"
	"github.com/prometheus/client_golang/prometheus"
	"testing"
)

func TestMetrics_ListMetricsCollectors(t *testing.T) {
	data := []byte(`{
	"deployment": "test-dev",
	"name": "exporters",
	"index": 0,
	"id": "b36ca4c9-80ce-426a-998e-8b23fd50efe3",
	"az": "z2",
	"other_field": "x"
}`)
	fetcher := fetchers.NewInstanceSpecFetcher("fake-path")
	spec, err := fetcher.FetchData(data)
	if err != nil {
		t.Fatalf("Fetch() returned error: %v", err)
	}
	metricsContext := &config.MetricsContext{
		Namespace:   "boshi",
		Environment: "test",
		BoshName:    "test-director",
		BoshUuid:    "123",
	}

	var list []prometheus.Collector
	var metrics Metrics

	metrics = NewBaseMetrics("test", "test", metricsContext, spec)
	list = ListMetricsCollectors(metrics)
	if len(list) <= 0 {
		t.Errorf("expected base collectors, got '%v'", list)
	}

	metrics = NewMonitMetrics(metricsContext, spec)
	list = ListMetricsCollectors(metrics)
	if len(list) <= 0 {
		t.Errorf("expected monit collectors, got '%v'", list)
	}

	metrics = NewSystemMetrics(metricsContext, spec)
	list = ListMetricsCollectors(metrics)
	if len(list) <= 0 {
		t.Errorf("expected system collectors, got '%v'", list)
	}
}
