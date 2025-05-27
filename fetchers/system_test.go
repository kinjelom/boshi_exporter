package fetchers

import (
	"context"
	"os"
	"testing"
)

func TestSystemFetcher_FetchSystem(t *testing.T) {
	t.Parallel()
	stat, err := NewSystemFetcher().fetchSystem(context.Background(), "", "", "")
	if err != nil {
		t.Fatalf("fetchSystem() returned error: %v", err)
	}

	if stat.Host == nil {
		t.Error("expected HOST to be non-nil")
	}
	if stat.CPU == nil {
		t.Error("expected CPU to be non-nil")
	}
	if stat.Memory == nil {
		t.Error("expected Memory to be non-nil")
	}
	if stat.Disks == nil {
		t.Error("expected Disks to be non-nil")
	}
}

func TestSystemFetcher_FetchHost(t *testing.T) {
	t.Parallel()
	stat, err := NewSystemFetcher().fetchHost(context.Background())
	if err != nil {
		t.Fatalf("fetchHost() returned error: %v", err)
	}

	if stat == nil {
		t.Fatal("fetchHost returned nil")
	}

	if stat.Load == nil {
		t.Error("expected Load to be non-nil")
	}
}

func TestSystemFetcher_FetchCPU(t *testing.T) {
	t.Parallel()
	stat, err := NewSystemFetcher().fetchCPU(context.Background())
	if err != nil {
		t.Fatalf("fetchCPU() returned error: %v", err)
	}

	if stat == nil {
		t.Fatal("fetchCPU returned nil")
	}
	if stat.LogicalCores <= 0 {
		t.Errorf("expected LogicalCores > 0, got %d", stat.LogicalCores)
	}
	if stat.PhysicalCores <= 0 {
		t.Errorf("expected PhysicalCores > 0, got %d", stat.PhysicalCores)
	}
}

func TestSystemFetcher_FetchMemory(t *testing.T) {
	t.Parallel()
	stat, err := NewSystemFetcher().fetchMemory(context.Background())
	if err != nil {
		t.Fatalf("fetchMemory() returned error: %v", err)
	}

	if stat == nil {
		t.Fatal("fetchMemory returned nil")
	}

	if stat.VM == nil {
		t.Error("expected VM (virtual memory) to be non-nil")
	} else if stat.VM.Total == 0 {
		t.Error("expected VM.Total > 0")
	}

	if stat.SwapMemory == nil {
		t.Error("expected SwapMemory to be non-nil")
	}
}

func TestSystemFetcher_FetchDisks(t *testing.T) {
	t.Parallel()

	// Use temp dirs to guarantee existence
	tmpData, err := os.MkdirTemp("", "data")
	if err != nil {
		t.Fatalf("failed to create temp dir for data: %v", err)
	}
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(tmpData)

	tmpStore, err := os.MkdirTemp("", "store")
	if err != nil {
		t.Fatalf("failed to create temp dir for store: %v", err)
	}
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(tmpStore)

	// rootDiskPath use OS temp dir (always exists)
	rootPath := os.TempDir()

	stat, err := NewSystemFetcher().fetchDisks(context.Background(), rootPath, tmpData, tmpStore)
	if err != nil {
		t.Fatalf("fetchDisks() returned error: %v", err)
	}

	if stat.RootDisk == nil {
		t.Error("expected RootDisk to be non-nil")
	} else {
		if stat.RootDisk.Total == 0 {
			t.Error("expected RootDisk.Total > 0")
		}
		if stat.RootDisk.Used > stat.RootDisk.Total {
			t.Errorf("RootDisk.Used (%d) cannot exceed Total (%d)", stat.RootDisk.Used, stat.RootDisk.Total)
		}
	}

	if stat.DataDisk == nil {
		t.Error("expected DataDisk to be non-nil")
	} else if stat.DataDisk.Total == 0 {
		t.Error("expected DataDisk.Total > 0")
	}

	if stat.StoreDisk == nil {
		t.Error("expected StoreDisk to be non-nil")
	} else if stat.StoreDisk.Total == 0 {
		t.Error("expected StoreDisk.Total > 0")
	}
}
