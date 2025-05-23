package collectors

import (
	"boshi_exporter/config"
	"boshi_exporter/fetchers"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	monitProcessNameLabel      = "process_name"
	monitMonitoringStatusLabel = "monitoring_status"
	monitProcessStatusLabel    = "process_status"
	monitProcessPidLabel       = "process_pid"
	monitProcessParentPidLabel = "process_parent_pid"
)

type MonitMetrics struct {
	StatusInfo                *prometheus.GaugeVec
	Uptime                    *prometheus.GaugeVec
	ChildrenCount             *prometheus.GaugeVec
	MemoryUsedBytes           *prometheus.GaugeVec
	MemoryUsedBytesTotal      *prometheus.GaugeVec
	MemoryUsageRatio          *prometheus.GaugeVec
	MemoryUsageRatioTotal     *prometheus.GaugeVec
	CPUUsageRatio             *prometheus.GaugeVec
	CPUUsageRatioTotal        *prometheus.GaugeVec
	CollectedTimestampSeconds *prometheus.GaugeVec
}

var _ Metrics = (*MonitMetrics)(nil)

func NewMonitMetrics(metricsContext *config.MetricsContext, spec *fetchers.InstanceSpec) *MonitMetrics {
	instanceLabels := NewInstanceLabels(metricsContext, spec)
	opts := func(name, help string, constantLabels *prometheus.Labels) prometheus.GaugeOpts {
		return prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "monit_process",
			Name:        name,
			Help:        help,
			ConstLabels: *constantLabels,
		}
	}
	minLabels := []string{monitProcessNameLabel, monitProcessPidLabel}
	fullLabels := []string{monitProcessNameLabel, monitMonitoringStatusLabel, monitProcessStatusLabel, monitProcessPidLabel, monitProcessParentPidLabel}
	return &MonitMetrics{

		StatusInfo:                promauto.NewGaugeVec(opts("status_info", "Monit process and monitoring status information", instanceLabels), fullLabels),
		Uptime:                    promauto.NewGaugeVec(opts("uptime_seconds", "Monit process uptime since last start (seconds)", instanceLabels), minLabels),
		ChildrenCount:             promauto.NewGaugeVec(opts("children_count", "Number of child processes", instanceLabels), minLabels),
		MemoryUsedBytes:           promauto.NewGaugeVec(opts("memory_used_bytes", "Process memory used in bytes", instanceLabels), minLabels),
		MemoryUsedBytesTotal:      promauto.NewGaugeVec(opts("memory_used_bytes_total", "Total process (with subprocesses) memory used in bytes", instanceLabels), minLabels),
		MemoryUsageRatio:          promauto.NewGaugeVec(opts("memory_usage_ratio", "Process memory usage fraction (1=100%)", instanceLabels), minLabels),
		MemoryUsageRatioTotal:     promauto.NewGaugeVec(opts("memory_usage_ratio_total", "Total process (with subprocesses) memory usage fraction (1=100%)", instanceLabels), minLabels),
		CPUUsageRatio:             promauto.NewGaugeVec(opts("cpu_usage_ratio", "Process CPU usage fraction (1=100%)", instanceLabels), minLabels),
		CPUUsageRatioTotal:        promauto.NewGaugeVec(opts("cpu_usage_ratio_total", "Total process (with subprocesses) CPU usage fraction (1=100%)", instanceLabels), minLabels),
		CollectedTimestampSeconds: promauto.NewGaugeVec(opts("collected_timestamp_seconds", "Data collection time as Unix timestamp (seconds).", instanceLabels), minLabels),
	}
}

func (m *MonitMetrics) Emit(stat fetchers.MonitStat) {
	for name, status := range stat {
		fullLabels := prometheus.Labels{
			monitProcessNameLabel:      name,
			monitMonitoringStatusLabel: status.MonitoringStatus,
			monitProcessStatusLabel:    status.Status,
			monitProcessPidLabel:       status.PID,
			monitProcessParentPidLabel: status.ParentPID,
		}
		m.StatusInfo.With(fullLabels).Set(1)

		minLabels := prometheus.Labels{
			monitProcessNameLabel: name,
			monitProcessPidLabel:  status.PID,
		}
		m.Uptime.With(minLabels).Set(status.Uptime.Seconds())
		m.ChildrenCount.With(minLabels).Set(float64(status.Children))
		m.MemoryUsedBytes.With(minLabels).Set(float64(status.MemoryBytes))
		m.MemoryUsedBytesTotal.With(minLabels).Set(float64(status.MemoryBytesTotal))
		m.MemoryUsageRatio.With(minLabels).Set(status.MemoryPercent / 100)
		m.MemoryUsageRatioTotal.With(minLabels).Set(status.MemoryPercentTotal / 100)
		m.CPUUsageRatio.With(minLabels).Set(status.CPUPercent / 100)
		m.CPUUsageRatioTotal.With(minLabels).Set(status.CPUPercentTotal / 100)
		m.CollectedTimestampSeconds.With(minLabels).Set(float64(status.DataCollected.Unix()))
	}
}

func (m *MonitMetrics) Collectors() []prometheus.Collector {
	return ListMetricsCollectors(m)
}
