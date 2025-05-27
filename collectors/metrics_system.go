package collectors

import (
	"boshi_exporter/config"
	"boshi_exporter/fetchers"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type SystemMetrics struct {
	Load1  *prometheus.GaugeVec
	Load5  *prometheus.GaugeVec
	Load15 *prometheus.GaugeVec

	// CPU
	CpuLogicalCoreCount  *prometheus.GaugeVec
	CpuPhysicalCoreCount *prometheus.GaugeVec

	// Virtual memory
	VmSize       *prometheus.GaugeVec // Total virtual memory in bytes
	VmAvailable  *prometheus.GaugeVec // Available virtual memory in bytes
	VmUsed       *prometheus.GaugeVec // Used virtual memory in bytes
	VmUsageRatio *prometheus.GaugeVec // Percentage of used virtual memory

	// Swap memory
	SwapSize       *prometheus.GaugeVec // Total swap memory in bytes
	SwapUsed       *prometheus.GaugeVec // Used swap memory in bytes
	SwapUsageRatio *prometheus.GaugeVec // Percentage of used swap memory

	// Disk usage
	DiskRootSize       *prometheus.GaugeVec // Total bytes on root filesystem
	DiskRootUsed       *prometheus.GaugeVec // Used bytes on root filesystem
	DiskRootUsageRatio *prometheus.GaugeVec // Percentage of used root filesystem

	DiskDataSize       *prometheus.GaugeVec // Total bytes on /var/vcap/data
	DiskDataUsed       *prometheus.GaugeVec // Used bytes on /var/vcap/data
	DiskDataUsageRatio *prometheus.GaugeVec // Percentage of used /var/vcap/data

	DiskStoreSize       *prometheus.GaugeVec // Total bytes on /var/vcap/store
	DiskStoreUsed       *prometheus.GaugeVec // Used bytes on /var/vcap/store
	DiskStoreUsageRatio *prometheus.GaugeVec // Percentage of used /var/vcap/store}
}

var _ Metrics = (*SystemMetrics)(nil)

func NewSystemMetrics(metricsContext *config.MetricsContext, spec *fetchers.InstanceSpec) *SystemMetrics {
	opts := func(name, help string) prometheus.GaugeOpts {
		return prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   "system",
			Name:        name,
			Help:        help,
			ConstLabels: *NewInstanceLabels(metricsContext, spec),
		}
	}
	return &SystemMetrics{

		Load1:  promauto.NewGaugeVec(opts("load1", "1-minute load average"), []string{}),
		Load5:  promauto.NewGaugeVec(opts("load5", "5-minute load average"), []string{}),
		Load15: promauto.NewGaugeVec(opts("load15", "15-minute load average"), []string{}),

		CpuLogicalCoreCount:  promauto.NewGaugeVec(opts("cpu_logical_core_count", "Number of logical CPU cores"), []string{}),
		CpuPhysicalCoreCount: promauto.NewGaugeVec(opts("cpu_physical_core_count", "Number of physical CPU cores"), []string{}),

		VmSize:       promauto.NewGaugeVec(opts("memory_virtual_size_bytes", "Total virtual memory in bytes"), []string{}),
		VmAvailable:  promauto.NewGaugeVec(opts("memory_virtual_available_bytes", "Available virtual memory in bytes"), []string{}),
		VmUsed:       promauto.NewGaugeVec(opts("memory_virtual_used_bytes", "Used virtual memory in bytes"), []string{}),
		VmUsageRatio: promauto.NewGaugeVec(opts("memory_virtual_usage_ratio", "Used virtual memory fraction (1=100%)"), []string{}),

		SwapSize:       promauto.NewGaugeVec(opts("memory_swap_size_bytes", "Total swap memory in bytes"), []string{}),
		SwapUsed:       promauto.NewGaugeVec(opts("memory_swap_used_bytes", "Used swap memory in bytes"), []string{}),
		SwapUsageRatio: promauto.NewGaugeVec(opts("memory_swap_usage_ratio", "Used swap memory fraction (1=100%)"), []string{}),

		DiskRootSize:       promauto.NewGaugeVec(opts("disk_root_size_bytes", "Total bytes on root filesystem"), []string{}),
		DiskRootUsed:       promauto.NewGaugeVec(opts("disk_root_used_bytes", "Used bytes on root filesystem"), []string{}),
		DiskRootUsageRatio: promauto.NewGaugeVec(opts("disk_root_usage_ratio", "Used root filesystem fraction (1=100%)"), []string{}),

		DiskDataSize:       promauto.NewGaugeVec(opts("disk_data_size_bytes", "Total bytes on /var/vcap/data"), []string{}),
		DiskDataUsed:       promauto.NewGaugeVec(opts("disk_data_used_bytes", "Used bytes on /var/vcap/data"), []string{}),
		DiskDataUsageRatio: promauto.NewGaugeVec(opts("disk_data_usage_ratio", "Used /var/vcap/data fraction (1=100%)"), []string{}),

		DiskStoreSize:       promauto.NewGaugeVec(opts("disk_store_size_bytes", "Total bytes on /var/vcap/store"), []string{}),
		DiskStoreUsed:       promauto.NewGaugeVec(opts("disk_store_used_bytes", "Used bytes on /var/vcap/store"), []string{}),
		DiskStoreUsageRatio: promauto.NewGaugeVec(opts("disk_store_usage_ratio", "Used /var/vcap/store fraction (1=100%)"), []string{}),
	}
}

func (m *SystemMetrics) Emit(stat *fetchers.SystemStat) {
	// Load averages
	if stat.Host.Load != nil {
		m.Load1.With(nil).Set(stat.Host.Load.Load1)
		m.Load5.With(nil).Set(stat.Host.Load.Load5)
		m.Load15.With(nil).Set(stat.Host.Load.Load15)
	}

	// CPU
	m.CpuLogicalCoreCount.With(nil).Set(float64(stat.CPU.LogicalCores))
	m.CpuPhysicalCoreCount.With(nil).Set(float64(stat.CPU.PhysicalCores))

	// Memory
	if stat.Memory.VM != nil {
		m.VmSize.With(nil).Set(float64(stat.Memory.VM.Total))
		m.VmAvailable.With(nil).Set(float64(stat.Memory.VM.Available))
		m.VmUsed.With(nil).Set(float64(stat.Memory.VM.Used))
		m.VmUsageRatio.With(nil).Set(stat.Memory.VM.UsedPercent / 100)
	}

	// Swap
	if stat.Memory.SwapMemory != nil {
		m.SwapSize.With(nil).Set(float64(stat.Memory.SwapMemory.Total))
		m.SwapUsed.With(nil).Set(float64(stat.Memory.SwapMemory.Used))
		m.SwapUsageRatio.With(nil).Set(stat.Memory.SwapMemory.UsedPercent / 100)
	}

	// Disks
	if stat.Disks.RootDisk != nil {
		m.DiskRootSize.With(nil).Set(float64(stat.Disks.RootDisk.Total))
		m.DiskRootUsed.With(nil).Set(float64(stat.Disks.RootDisk.Used))
		m.DiskRootUsageRatio.With(nil).Set(stat.Disks.RootDisk.UsedPercent / 100)
	}
	if stat.Disks.DataDisk != nil {
		m.DiskDataSize.With(nil).Set(float64(stat.Disks.DataDisk.Total))
		m.DiskDataUsed.With(nil).Set(float64(stat.Disks.DataDisk.Used))
		m.DiskDataUsageRatio.With(nil).Set(stat.Disks.DataDisk.UsedPercent / 100)
	}
	if stat.Disks.StoreDisk != nil {
		m.DiskStoreSize.With(nil).Set(float64(stat.Disks.StoreDisk.Total))
		m.DiskStoreUsed.With(nil).Set(float64(stat.Disks.StoreDisk.Used))
		m.DiskStoreUsageRatio.With(nil).Set(stat.Disks.StoreDisk.UsedPercent / 100)
	}
}

func (m *SystemMetrics) Collectors() []prometheus.Collector {
	return ListMetricsCollectors(m)
}
