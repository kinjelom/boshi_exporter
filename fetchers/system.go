package fetchers

import (
	"context"
	"fmt"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
)

// SystemFetcher is responsible for gathering system metrics
type SystemFetcher struct{}

type HostStat struct {
	Load *load.AvgStat
}

type CPUStat struct {
	LogicalCores  int // number of logical CPU cores
	PhysicalCores int // number of logical CPU cores
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
	hostStat, err := m.fetchHost(ctx)
	if err != nil {
		return nil, err
	}
	cpuStat, err := m.fetchCPU(ctx)
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
	return &SystemStat{Host: hostStat, CPU: cpuStat, Memory: memoryStat, Disks: disksStat}, nil
}

func (m *SystemFetcher) fetchHost(ctx context.Context) (*HostStat, error) {
	stat := &HostStat{}
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
	return stat, nil
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
