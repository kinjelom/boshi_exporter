package fetchers

import (
	"context"
	"fmt"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"time"
)

// SystemFetcher is responsible for gathering system metrics
type SystemFetcher struct{}

type HostStat struct {
	BootTime uint64
	UpTime   uint64
	Load     *load.AvgStat
}

type CPUStat struct {
	LogicalCores       int            // number of logical CPU cores
	PhysicalCores      int            // number of logical CPU cores
	Times              *cpu.TimesStat // Initial CPU times snapshot (user, system, idle, etc.)
	TimesTimestamp     time.Time      // when Times was read
	NextTimes          *cpu.TimesStat // next CPU times (after at least 1s)
	NextTimesTimestamp time.Time      // when NextTimes was read
}

// CPUUsageRatio calculates the CPU usage percentage between two snapshots, 1=100%.
// Returns 0.0 if either snapshot is missing or invalid.
func (s *CPUStat) CPUUsageRatio() float64 {
	// Guard: need both previous and next snapshots
	if s == nil || s.Times == nil || s.NextTimes == nil {
		return 0.0
	}
	// Guard: ensure timestamps are in the right order
	if !s.NextTimesTimestamp.After(s.TimesTimestamp) {
		return 0.0
	}

	prev := s.Times
	curr := s.NextTimes

	// idle includes idle + iowait
	prevIdle := prev.Idle + prev.Iowait
	currIdle := curr.Idle + curr.Iowait

	// non-idle = user + system + nice + irq + softirq + steal
	prevNonIdle := prev.User + prev.System + prev.Nice + prev.Irq + prev.Softirq + prev.Steal
	currNonIdle := curr.User + curr.System + curr.Nice + curr.Irq + curr.Softirq + curr.Steal

	prevTotal := prevIdle + prevNonIdle
	currTotal := currIdle + currNonIdle

	totalDelta := currTotal - prevTotal
	idleDelta := currIdle - prevIdle

	// Guard: avoid division by zero or negative deltas
	if totalDelta <= 0 {
		return 0.0
	}

	return 1.0 - idleDelta/totalDelta
}

type MemoryStat struct {
	VM         *mem.VirtualMemoryStat // virtual memory statistics
	SwapMemory *mem.SwapMemoryStat    // swap memory
}

type DisksStat struct {
	RootDisk  *disk.UsageStat // root disk usage
	DataDisk  *disk.UsageStat // /var/vcap/data disk usage
	StoreDisk *disk.UsageStat // /var/vcap/store disk usage
}

// SystemStat holds metrics about disk, memory, and CPU usage
type SystemStat struct {
	Host   *HostStat
	CPU    *CPUStat
	Memory *MemoryStat
	Disks  *DisksStat
}

// NewSystemFetcher initializes a new SystemFetcher
func NewSystemFetcher() *SystemFetcher {
	return &SystemFetcher{}
}

// Fetch retrieves current system metrics and returns them
func (m *SystemFetcher) Fetch(ctx context.Context) (*SystemStat, error) {
	return m.fetchSystem(ctx, "/", "/var/vcap/data", "/var/vcap/store")
}

// Fetch retrieves current system metrics and returns them
func (m *SystemFetcher) fetchSystem(ctx context.Context, rootDiskPath, dataDiskPath, storeDiskPath string) (*SystemStat, error) {
	cpuStat, err := m.fetchCPU(ctx)
	if err != nil {
		return nil, err
	}
	hostStat, err := m.fetchHost(ctx)
	if err != nil {
		return nil, err
	}
	memoryStat, err := m.fetchMemory(ctx)
	if err != nil {
		return nil, err
	}
	disksStat, err := m.fetchDisks(ctx, rootDiskPath, dataDiskPath, storeDiskPath)
	if err != nil {
		return nil, err
	}
	err = m.fetchCPUNext(ctx, cpuStat)
	if err != nil {
		return nil, err
	}
	return &SystemStat{Host: hostStat, CPU: cpuStat, Memory: memoryStat, Disks: disksStat}, nil
}

func (m *SystemFetcher) fetchHost(ctx context.Context) (*HostStat, error) {
	stat := &HostStat{}
	if bootTime, err := host.BootTimeWithContext(ctx); err == nil {
		stat.BootTime = bootTime
	} else {
		return nil, fmt.Errorf("failed to get boot time, error: %v", err)
	}
	if uptime, err := host.UptimeWithContext(ctx); err == nil {
		stat.UpTime = uptime
	} else {
		return nil, fmt.Errorf("failed to get uptime, error: %v", err)
	}
	if avg, err := load.AvgWithContext(ctx); err == nil {
		stat.Load = avg
	} else {
		return nil, fmt.Errorf("failed to get load average, error: %v", err)
	}
	return stat, nil
}

func (m *SystemFetcher) fetchCPU(ctx context.Context) (*CPUStat, error) {
	stat := &CPUStat{}
	// CPU cores count
	if counts, err := cpu.CountsWithContext(ctx, true); err == nil {
		stat.LogicalCores = counts
	} else {
		return nil, fmt.Errorf("failed to get CPU logical cores count, error: %v", err)
	}
	if counts, err := cpu.CountsWithContext(ctx, false); err == nil {
		stat.PhysicalCores = counts
	} else {
		return nil, fmt.Errorf("failed to get CPU physical cores count, error: %v", err)
	}
	// CPU times
	if times, err := cpu.TimesWithContext(ctx, false); err == nil && len(times) > 0 {
		stat.Times = &times[0]
		stat.TimesTimestamp = time.Now()
	} else {
		return nil, fmt.Errorf("failed to get CPU times, error: %v", err)
	}
	return stat, nil
}

func (m *SystemFetcher) fetchCPUNext(ctx context.Context, s *CPUStat) error {
	elapsed := time.Since(s.TimesTimestamp)
	if remaining := time.Second - elapsed; remaining > 0 {
		time.Sleep(remaining)
	}

	times, err := cpu.TimesWithContext(ctx, false)
	if err != nil || len(times) == 0 {
		return fmt.Errorf("failed to get next CPU times: %v", err)
	}
	s.NextTimes = &times[0]
	s.NextTimesTimestamp = time.Now()

	return nil
}

func (m *SystemFetcher) fetchMemory(ctx context.Context) (*MemoryStat, error) {
	stat := &MemoryStat{}
	// memory usage
	if vm, err := mem.VirtualMemoryWithContext(ctx); err == nil {
		stat.VM = vm
	} else {
		return nil, fmt.Errorf("failed to get memory, error: %v", err)
	}
	// swap memory
	if sm, err := mem.SwapMemoryWithContext(ctx); err == nil {
		stat.SwapMemory = sm
	} else {
		return nil, fmt.Errorf("failed to get swap memory, error: %v", err)
	}
	return stat, nil
}

func (m *SystemFetcher) fetchDisks(ctx context.Context, rootDiskPath, dataDiskPath, storeDiskPath string) (*DisksStat, error) {
	stat := &DisksStat{}
	// disk usage
	if rootDiskPath != "" {
		if usage, err := disk.UsageWithContext(ctx, rootDiskPath); err == nil {
			stat.RootDisk = usage
		} else {
			return nil, fmt.Errorf("failed to get root disk usage, error: %v", err)
		}
	}
	if dataDiskPath != "" {
		if usage, err := disk.UsageWithContext(ctx, dataDiskPath); err == nil {
			stat.DataDisk = usage
		} else {
			return nil, fmt.Errorf("failed to get data disk usage, error: %v", err)
		}
	}
	if storeDiskPath != "" {
		if usage, err := disk.UsageWithContext(ctx, storeDiskPath); err == nil {
			stat.StoreDisk = usage
		} else {
			return nil, fmt.Errorf("failed to get store disk usage, error: %v", err)
		}
	}
	return stat, nil
}
