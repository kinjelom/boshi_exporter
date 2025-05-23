package fetchers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMonitFetcher_BinNotFound(t *testing.T) {
	fetcher := NewMonitFetcher("/does/not/exist")
	_, err := fetcher.Fetch(context.Background())
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestMonitFetcher_ParsesSuccessfulSampleOutput(t *testing.T) {
	sampleOutput := `The Monit daemon 5.2.5 uptime: 19h 17m

Process 'boshi_exporter'
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

System 'system_da9acacc-dac6-4ee3-9388-83f4451415c2'
  status                            running
  monitoring status                 monitored
  load average                      [0.09] [0.10] [0.04]
  cpu                               2.0%us 4.0%sy 0.0%wa
  memory usage                      227240 kB [23.0%]
  swap usage                        256 kB [0.0%]
  data collected                    Fri May 23 11:03:35 2025
`
	fetcher := NewMonitFetcher("fake-path")
	statusMap, err := fetcher.parseData(sampleOutput)
	assert.NoError(t, err, "Fetch should complete without error")

	// Verify process 'boshi'
	boshi, exists := statusMap["boshi_exporter"]
	assert.True(t, exists, "entry for process 'boshi' should exist")
	assert.Equal(t, "running", boshi.Status)
	assert.Equal(t, "monitored", boshi.MonitoringStatus)
	assert.Equal(t, "65763", boshi.PID)
	assert.Equal(t, 1, boshi.Children)
	assert.InDelta(t, 0.0, boshi.MemoryPercent, 1e-6)
	assert.InDelta(t, 16.3, boshi.MemoryPercentTotal, 1e-6)

	// Verify process 'blackbox'
	blackbox, exists := statusMap["blackbox"]
	assert.True(t, exists, "entry for process 'blackbox' should exist")
	assert.Equal(t, 9992*1024, blackbox.MemoryBytes)
	assert.InDelta(t, 0.4, blackbox.CPUPercent, 1e-6)

	// Verify system entry
	systemEntry, exists := statusMap["system_da9acacc-dac6-4ee3-9388-83f4451415c2"]
	assert.True(t, exists, "system entry should exist")
	assert.Equal(t, "running", systemEntry.Status)
	assert.Equal(t, "monitored", systemEntry.MonitoringStatus)
}

func TestMonitFetcher_ParsesFailedSampleOutput(t *testing.T) {
	sampleOutput := `The Monit daemon 5.2.5 uptime: 8m 

Process 'boshi_exporter'
  status                            not monitored - unmonitor pending
  monitoring status                 not monitored
  data collected                    Fri May 23 11:03:35 2025

Process 'blackbox'
  status                            not monitored
  monitoring status                 not monitored
  data collected                    Fri May 23 11:03:35 2025

System 'system_da9acacc-dac6-4ee3-9388-83f4451415c2'
  status                            running
  monitoring status                 monitored
  data collected                    Tue May 20 10:33:22 2025
`
	fetcher := NewMonitFetcher("fake-path")
	statusMap, err := fetcher.parseData(sampleOutput)
	assert.NoError(t, err, "Fetch should complete without error")

	// Verify process 'boshi'
	boshi, exists := statusMap["boshi_exporter"]
	assert.True(t, exists, "entry for process 'boshi' should exist")
	assert.Equal(t, "not monitored - unmonitor pending", boshi.Status)
	assert.Equal(t, "not monitored", boshi.MonitoringStatus)

	// Verify process 'blackbox'
	blackbox, exists := statusMap["blackbox"]
	assert.True(t, exists, "entry for process 'blackbox' should exist")
	assert.Equal(t, "not monitored", blackbox.Status)
	assert.Equal(t, "not monitored", blackbox.MonitoringStatus)

	// Verify system entry
	systemEntry, exists := statusMap["system_da9acacc-dac6-4ee3-9388-83f4451415c2"]
	assert.True(t, exists, "system entry should exist")
	assert.Equal(t, "running", systemEntry.Status)
	assert.Equal(t, "monitored", systemEntry.MonitoringStatus)
}
