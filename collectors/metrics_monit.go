package collectors

import (
	"boshi_exporter/config"
	"boshi_exporter/fetchers"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	monitVersionLabel          = "monit_version"
	monitMonitoringStatusLabel = "monitoring_status"
	monitServiceStatusLabel    = "service_status"
	monitProcessNameLabel      = "process_name"
	monitProcessPidLabel       = "process_pid"
	monitProcessParentPidLabel = "process_parent_pid"
)

type MonitMetrics struct {
	MonitInfo   *prometheus.GaugeVec
	MonitUptime *prometheus.GaugeVec

	SysStatusInfo                *prometheus.GaugeVec
	SysUptimeSeconds             *prometheus.GaugeVec
	SysLoadAvg1                  *prometheus.GaugeVec
	SysLoadAvg5                  *prometheus.GaugeVec
	SysLoadAvg15                 *prometheus.GaugeVec
	SysCPUUserRatio              *prometheus.GaugeVec
	SysCPUSystemRatio            *prometheus.GaugeVec
	SysCPUWaitRatio              *prometheus.GaugeVec
	SysMemoryUsedBytes           *prometheus.GaugeVec
	SysMemoryUsageRatio          *prometheus.GaugeVec
	SysSwapUsedBytes             *prometheus.GaugeVec
	SysSwapUsageRatio            *prometheus.GaugeVec
	SysCollectedTimestampSeconds *prometheus.GaugeVec

	ProcStatusInfo                *prometheus.GaugeVec
	ProcUptime                    *prometheus.GaugeVec
	ProcChildrenCount             *prometheus.GaugeVec
	ProcMemoryUsedBytes           *prometheus.GaugeVec
	ProcMemoryUsedBytesTotal      *prometheus.GaugeVec
	ProcMemoryUsageRatio          *prometheus.GaugeVec
	ProcMemoryUsageRatioTotal     *prometheus.GaugeVec
	ProcCPUUsageRatio             *prometheus.GaugeVec
	ProcCPUUsageRatioTotal        *prometheus.GaugeVec
	ProcCollectedTimestampSeconds *prometheus.GaugeVec
}

var _ Metrics = (*MonitMetrics)(nil)

func NewMonitMetrics(metricsContext *config.MetricsContext, spec *fetchers.InstanceSpec) *MonitMetrics {
	instanceLabels := NewInstanceLabels(metricsContext, spec)
	opts := func(name, help string, constantLabels *prometheus.Labels) prometheus.GaugeOpts {
		return prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "monit",
			Name:        name,
			Help:        help,
			ConstLabels: *constantLabels,
		}
	}
	monitInfoLabels := []string{monitVersionLabel}
	var monitLabels []string
	sysInfoLabels := []string{monitMonitoringStatusLabel, monitServiceStatusLabel}
	var sysLabels []string
	procInfoLabels := []string{monitProcessNameLabel, monitProcessPidLabel}
	procLabels := []string{monitProcessNameLabel, monitMonitoringStatusLabel, monitServiceStatusLabel, monitProcessPidLabel, monitProcessParentPidLabel}
	return &MonitMetrics{
		// Monit metrics
		MonitInfo:   promauto.NewGaugeVec(opts("monit_info", "The Monit daemon information", instanceLabels), monitInfoLabels),
		MonitUptime: promauto.NewGaugeVec(opts("uptime_seconds", "Monit uptime since last start (seconds)", instanceLabels), monitLabels),

		// System metrics
		SysStatusInfo:                promauto.NewGaugeVec(opts("system_status_info", "System status info (e.g., running, monitored)", instanceLabels), sysInfoLabels),
		SysUptimeSeconds:             promauto.NewGaugeVec(opts("system_uptime_seconds", "System uptime since last boot (seconds)", instanceLabels), sysLabels),
		SysLoadAvg1:                  promauto.NewGaugeVec(opts("system_load_avg_1m", "System 1-minute load average", instanceLabels), sysLabels),
		SysLoadAvg5:                  promauto.NewGaugeVec(opts("system_load_avg_5m", "System 5-minute load average", instanceLabels), sysLabels),
		SysLoadAvg15:                 promauto.NewGaugeVec(opts("system_load_avg_15m", "System 15-minute load average", instanceLabels), sysLabels),
		SysCPUUserRatio:              promauto.NewGaugeVec(opts("system_cpu_user_ratio", "System CPU time spent in user mode fraction (1=100%)", instanceLabels), sysLabels),
		SysCPUSystemRatio:            promauto.NewGaugeVec(opts("system_cpu_system_ratio", "System CPU time spent in kernel/system mode fraction (1=100%)", instanceLabels), sysLabels),
		SysCPUWaitRatio:              promauto.NewGaugeVec(opts("system_cpu_wait_ratio", "System CPU time spent waiting for I/O fraction (1=100%)", instanceLabels), sysLabels),
		SysMemoryUsedBytes:           promauto.NewGaugeVec(opts("system_memory_used_bytes", "System memory used in bytes", instanceLabels), sysLabels),
		SysMemoryUsageRatio:          promauto.NewGaugeVec(opts("system_memory_usage_ratio", "System memory used as a fraction of total (1=100%)", instanceLabels), sysLabels),
		SysSwapUsedBytes:             promauto.NewGaugeVec(opts("system_swap_used_bytes", "System swap used in bytes", instanceLabels), sysLabels),
		SysSwapUsageRatio:            promauto.NewGaugeVec(opts("system_swap_usage_ratio", "System swap used as a fraction of total (1=100%)", instanceLabels), sysLabels),
		SysCollectedTimestampSeconds: promauto.NewGaugeVec(opts("system_collected_timestamp_seconds", "System data collection time as Unix timestamp (seconds).", instanceLabels), sysLabels),

		// Process metrics
		ProcStatusInfo:                promauto.NewGaugeVec(opts("process_status_info", "Monit process and monitoring status information", instanceLabels), procLabels),
		ProcUptime:                    promauto.NewGaugeVec(opts("process_uptime_seconds", "Monit process uptime since last start (seconds)", instanceLabels), procInfoLabels),
		ProcChildrenCount:             promauto.NewGaugeVec(opts("process_children_count", "Number of child processes", instanceLabels), procInfoLabels),
		ProcMemoryUsedBytes:           promauto.NewGaugeVec(opts("process_memory_used_bytes", "Process memory used in bytes", instanceLabels), procInfoLabels),
		ProcMemoryUsedBytesTotal:      promauto.NewGaugeVec(opts("process_memory_used_bytes_total", "Total process (with subprocesses) memory used in bytes", instanceLabels), procInfoLabels),
		ProcMemoryUsageRatio:          promauto.NewGaugeVec(opts("process_memory_usage_ratio", "Process memory usage fraction (1=100%)", instanceLabels), procInfoLabels),
		ProcMemoryUsageRatioTotal:     promauto.NewGaugeVec(opts("process_memory_usage_ratio_total", "Total process (with subprocesses) memory usage fraction (1=100%)", instanceLabels), procInfoLabels),
		ProcCPUUsageRatio:             promauto.NewGaugeVec(opts("process_cpu_usage_ratio", "Process CPU usage fraction (1=100%)", instanceLabels), procInfoLabels),
		ProcCPUUsageRatioTotal:        promauto.NewGaugeVec(opts("process_cpu_usage_ratio_total", "Total process (with subprocesses) CPU usage fraction (1=100%)", instanceLabels), procInfoLabels),
		ProcCollectedTimestampSeconds: promauto.NewGaugeVec(opts("process_collected_timestamp_seconds", "Process data collection time as Unix timestamp (seconds).", instanceLabels), procInfoLabels),
	}
}

func (m *MonitMetrics) Emit(stat *fetchers.MonitStat) {
	m.MonitInfo.With(prometheus.Labels{monitVersionLabel: stat.Version}).Set(1)
	m.MonitUptime.With(prometheus.Labels{}).Set(stat.Uptime.Seconds())
	m.SysStatusInfo.With(prometheus.Labels{monitMonitoringStatusLabel: stat.System.MonitoringStatus, monitServiceStatusLabel: stat.System.Status}).Set(1)
	m.SysLoadAvg1.With(prometheus.Labels{}).Set(stat.System.LoadAvg1)
	m.SysLoadAvg5.With(prometheus.Labels{}).Set(stat.System.LoadAvg5)
	m.SysLoadAvg15.With(prometheus.Labels{}).Set(stat.System.LoadAvg15)
	m.SysCPUUserRatio.With(prometheus.Labels{}).Set(stat.System.CPUUserPercent / 100)
	m.SysCPUSystemRatio.With(prometheus.Labels{}).Set(stat.System.CPUSystemPercent / 100)
	m.SysCPUWaitRatio.With(prometheus.Labels{}).Set(stat.System.CPUWaitPercent / 100)
	m.SysMemoryUsedBytes.With(prometheus.Labels{}).Set(float64(stat.System.MemoryUsedBytes))
	m.SysMemoryUsageRatio.With(prometheus.Labels{}).Set(stat.System.MemoryUsedPercent / 100)
	m.SysSwapUsedBytes.With(prometheus.Labels{}).Set(float64(stat.System.SwapUsedBytes))
	m.SysSwapUsageRatio.With(prometheus.Labels{}).Set(stat.System.SwapUsedPercent / 100)
	m.SysCollectedTimestampSeconds.With(prometheus.Labels{}).Set(float64(stat.System.DataCollected.Unix()))

	for name, status := range stat.Processes {
		procInfoLabels := prometheus.Labels{
			monitProcessNameLabel:      name,
			monitMonitoringStatusLabel: status.MonitoringStatus,
			monitServiceStatusLabel:    status.Status,
			monitProcessPidLabel:       status.PID,
			monitProcessParentPidLabel: status.ParentPID,
		}
		m.ProcStatusInfo.With(procInfoLabels).Set(1)

		procLabels := prometheus.Labels{
			monitProcessNameLabel: name,
			monitProcessPidLabel:  status.PID,
		}
		m.ProcUptime.With(procLabels).Set(status.Uptime.Seconds())
		m.ProcChildrenCount.With(procLabels).Set(float64(status.Children))
		m.ProcMemoryUsedBytes.With(procLabels).Set(float64(status.MemoryUsedBytes))
		m.ProcMemoryUsedBytesTotal.With(procLabels).Set(float64(status.MemoryUsedBytesTotal))
		m.ProcMemoryUsageRatio.With(procLabels).Set(status.MemoryUsedPercent / 100)
		m.ProcMemoryUsageRatioTotal.With(procLabels).Set(status.MemoryUsedPercentTotal / 100)
		m.ProcCPUUsageRatio.With(procLabels).Set(status.CPUUsedPercent / 100)
		m.ProcCPUUsageRatioTotal.With(procLabels).Set(status.CPUUsedPercentTotal / 100)
		m.ProcCollectedTimestampSeconds.With(procLabels).Set(float64(status.DataCollected.Unix()))
	}
}

func (m *MonitMetrics) Collectors() []prometheus.Collector {
	return ListMetricsCollectors(m)
}
