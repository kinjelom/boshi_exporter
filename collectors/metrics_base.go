package collectors

import (
	"boshi_exporter/config"
	"boshi_exporter/fetchers"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type BaseMetrics struct {
	BuildInfo    *prometheus.GaugeVec
	InstanceInfo *prometheus.GaugeVec
}

var _ Metrics = (*BaseMetrics)(nil)

func NewBaseMetrics(programName, programVersion string, metricsContext *config.MetricsContext, spec *fetchers.InstanceSpec) *BaseMetrics {
	opts := func(name, help string, constantLabels *prometheus.Labels) prometheus.GaugeOpts {
		return prometheus.GaugeOpts{
			Namespace:   namespace,
			Name:        name,
			Help:        help,
			ConstLabels: *constantLabels,
		}
	}
	return &BaseMetrics{
		BuildInfo: promauto.NewGaugeVec(
			opts("build_info", "Program build information", &prometheus.Labels{
				programNameLabel:    programName,
				programVersionLabel: programVersion,
			}),
			[]string{},
		),
		InstanceInfo: promauto.NewGaugeVec(
			opts("instance_info", "Bosh instance information", NewInstanceLabels(metricsContext, spec)),
			[]string{}),
	}
}

func (m *BaseMetrics) Emit() {
	m.BuildInfo.With(nil).Set(1)
	m.InstanceInfo.With(nil).Set(1)
}

func (m *BaseMetrics) Collectors() []prometheus.Collector {
	return ListMetricsCollectors(m)
}
