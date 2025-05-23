package fetchers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

const sampleOutput1 = `The Monit daemon 5.2.5 uptime: 19h 17m

Process 'gitea'
  status                            running
  monitoring status                 monitored
  pid                               65763
  parent pid                        1
  uptime                            2h 12m
  children                          1
  memory kilobytes                  992
  memory kilobytes total            328184
  memory percent                    0.0%
  memory percent total              16.3%
  cpu percent                       0.0%
  cpu percent total                 0.0%
  data collected                    Tue May 20 10:33:22 2025

Process 'blackbox'
  status                            running
  monitoring status                 monitored
  pid                               65799
  parent pid                        1
  uptime                            2h 12m
  children                          0
  memory kilobytes                  9992
  memory kilobytes total            9992
  memory percent                    0.4%
  memory percent total              0.4%
  cpu percent                       0.4%
  cpu percent total                 0.4%
  data collected                    Tue May 20 10:33:22 2025

System 'system_c9020371-a500-4399-8b2f-765cdf01635e'
  status                            running
  monitoring status                 monitored
  data collected                    Tue May 20 10:33:22 2025
`

func TestMonitFetcher_ParsesProcessesAndSystem(t *testing.T) {
	fetcher := NewMonitFetcher("fake-path")
	statusMap, err := fetcher.parseData(sampleOutput1)
	assert.NoError(t, err, "Fetch should complete without error")

	// Verify process 'gitea'
	gitea, exists := statusMap["gitea"]
	assert.True(t, exists, "entry for process 'gitea' should exist")
	assert.Equal(t, "running", gitea.Status)
	assert.Equal(t, "monitored", gitea.MonitoringStatus)
	assert.Equal(t, "65763", gitea.PID)
	assert.Equal(t, 1, gitea.Children)
	assert.InDelta(t, 0.0, gitea.MemoryPercent, 1e-6)
	assert.InDelta(t, 16.3, gitea.MemoryPercentTotal, 1e-6)

	// Verify process 'blackbox'
	blackbox, exists := statusMap["blackbox"]
	assert.True(t, exists, "entry for process 'blackbox' should exist")
	assert.Equal(t, 9992*1024, blackbox.MemoryBytes)
	assert.InDelta(t, 0.4, blackbox.CPUPercent, 1e-6)

	// Verify system entry
	systemEntry, exists := statusMap["system_c9020371-a500-4399-8b2f-765cdf01635e"]
	assert.True(t, exists, "system entry should exist")
	assert.Equal(t, "running", systemEntry.Status)
	assert.Equal(t, "monitored", systemEntry.MonitoringStatus)
}

func TestMonitFetcher_BinNotFound(t *testing.T) {
	fetcher := NewMonitFetcher("/does/not/exist")
	_, err := fetcher.Fetch(context.Background())
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}
