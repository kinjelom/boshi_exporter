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

func TestMonitFetcher_ParsesSuccessfulSampleOutput1(t *testing.T) {
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
  memory percent                    0.1%
  memory percent total              16.4%
  cpu percent                       0.0%
  cpu percent total                 0.0%
  data collected                    Tue May 20 10:33:22 2025

System 'system_da9acacc-dac6-4ee3-9388-83f4451415c2'
  status                            running
  monitoring status                 monitored
  load average                      [0.09] [0.10] [0.04]
  cpu                               2.0%us 4.0%sy 0.1%wa
  memory usage                      227240 kB [23.0%]
  swap usage                        256 kB [0.1%]
  data collected                    Fri May 23 11:03:35 2025
`
	fetcher := NewMonitFetcher("fake-path")
	stat, err := fetcher.parseData(sampleOutput)
	assert.NoError(t, err, "Fetch should complete without error")

	// Verify process
	process, exists := stat.Processes["boshi_exporter"]
	assert.True(t, exists, "entry for process 'boshi_exporter' should exist")
	assert.Equal(t, "running", process.Status)
	assert.Equal(t, "monitored", process.MonitoringStatus)
	assert.Equal(t, "65763", process.PID)
	assert.Equal(t, "1", process.ParentPID)
	assert.Equal(t, "2h12m0s", process.Uptime.String())
	assert.Equal(t, 1, process.Children)
	assert.Equal(t, uint64(992*1024), process.MemoryUsedBytes)
	assert.Equal(t, uint64(328184*1024), process.MemoryUsedBytesTotal)
	assert.InDelta(t, 0.1, process.MemoryUsedPercent, 1e-6)
	assert.InDelta(t, 16.4, process.MemoryUsedPercentTotal, 1e-6)
	assert.Equal(t, "2025-05-20 10:33:22", process.DataCollected.Format("2006-01-02 15:04:05"))

	// Verify system entry
	sys := stat.System
	assert.Equal(t, "running", sys.Status)
	assert.Equal(t, "monitored", sys.MonitoringStatus)
	assert.InDelta(t, 0.09, sys.LoadAvg1, 1e-6)
	assert.InDelta(t, 0.1, sys.LoadAvg5, 1e-6)
	assert.InDelta(t, 0.04, sys.LoadAvg15, 1e-6)
	assert.InDelta(t, 2, sys.CPUUserPercent, 1e-6)
	assert.InDelta(t, 4, sys.CPUSystemPercent, 1e-6)
	assert.InDelta(t, 0.1, sys.CPUIOWaitPercent, 1e-6)
	assert.Equal(t, uint64(227240*1024), sys.MemoryUsedBytes)
	assert.InDelta(t, 23, sys.MemoryUsedPercent, 1e-6)
	assert.Equal(t, uint64(256*1024), sys.SwapUsedBytes)
	assert.InDelta(t, 0.1, sys.SwapUsedPercent, 1e-6)
	assert.Equal(t, "2025-05-23 11:03:35", sys.DataCollected.Format("2006-01-02 15:04:05"))
}

func TestMonitFetcher_ParsesSuccessfulSampleOutput2(t *testing.T) {
	sampleOutput := `The Monit daemon 5.2.5 uptime: 2d 17h 13m 

Process 'elasticsearch'
  status                            running
  monitoring status                 monitored
  pid                               4701
  parent pid                        1
  uptime                            2d 17h 12m 
  children                          2
  memory kilobytes                  23448
  memory kilobytes total            1309768
  memory percent                    1.1%
  memory percent total              65.1%
  cpu percent                       0.9%
  cpu percent total                 3.8%
  data collected                    Tue May 27 07:12:02 2025

File 'nfs_mounter'
  status                            accessible
  monitoring status                 monitored
  permission                        644
  uid                               0
  gid                               0
  timestamp                         Sat May 24 13:59:17 2025
  size                              68 B
  data collected                    Tue May 27 07:12:02 2025

Process 'blackbox'
  status                            running
  monitoring status                 monitored
  pid                               4754
  parent pid                        1
  uptime                            2d 17h 12m 
  children                          0
  memory kilobytes                  7904
  memory kilobytes total            7904
  memory percent                    0.3%
  memory percent total              0.3%
  cpu percent                       0.0%
  cpu percent total                 0.0%
  data collected                    Tue May 27 07:12:02 2025

System 'system_3c903c64-5da8-481f-a37a-80ceabbae891'
  status                            running
  monitoring status                 monitored
  load average                      [1.35] [1.50] [1.40]
  cpu                               5.1%us 5.1%sy 8.5%wa
  memory usage                      1358016 kB [67.5%]
  swap usage                        143360 kB [7.1%]
  data collected                    Tue May 27 07:12:02 2025
`
	fetcher := NewMonitFetcher("fake-path")
	stat, err := fetcher.parseData(sampleOutput)
	assert.NoError(t, err, "Fetch should complete without error")

	processName := "elasticsearch"
	process, exists := stat.Processes[processName]
	assert.True(t, exists, "entry for process '"+processName+"' should exist")
	assert.Equal(t, "running", process.Status)
	assert.Equal(t, "monitored", process.MonitoringStatus)
	assert.Equal(t, "4701", process.PID)

	processName = "blackbox"
	process, exists = stat.Processes[processName]
	assert.True(t, exists, "entry for process '"+processName+"' should exist")
	assert.Equal(t, "running", process.Status)
	assert.Equal(t, "monitored", process.MonitoringStatus)
	assert.Equal(t, "4754", process.PID)

	sys := stat.System
	assert.True(t, exists, "system entry should exist")
	assert.Equal(t, "running", sys.Status)
	assert.Equal(t, "monitored", sys.MonitoringStatus)
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
	stat, err := fetcher.parseData(sampleOutput)
	assert.NoError(t, err, "Fetch should complete without error")

	// Verify process 'boshi_exporter'
	boshi, exists := stat.Processes["boshi_exporter"]
	assert.True(t, exists, "entry for process 'boshi_exporter' should exist")
	assert.Equal(t, "not monitored - unmonitor pending", boshi.Status)
	assert.Equal(t, "not monitored", boshi.MonitoringStatus)

	// Verify process 'blackbox'
	blackbox, exists := stat.Processes["blackbox"]
	assert.True(t, exists, "entry for process 'blackbox' should exist")
	assert.Equal(t, "not monitored", blackbox.Status)
	assert.Equal(t, "not monitored", blackbox.MonitoringStatus)

	// Verify system entry
	sys := stat.System
	assert.True(t, exists, "system entry should exist")
	assert.Equal(t, "running", sys.Status)
	assert.Equal(t, "monitored", sys.MonitoringStatus)
}
func TestMonitFetcher_ParsesJumpBoxOutput(t *testing.T) {
	sampleOutput := `The Monit daemon 5.2.5 uptime: 57m 

Process 'blackbox'
  status                            running
  monitoring status                 monitored
  pid                               3203
  parent pid                        1
  uptime                            57m 
  children                          0
  memory kilobytes                  8476
  memory kilobytes total            8476
  memory percent                    0.8%
  memory percent total              0.8%
  cpu percent                       0.9%
  cpu percent total                 0.9%
  data collected                    Tue May 27 14:18:45 2025

Process 'boshi_exporter'
  status                            running
  monitoring status                 monitored
  pid                               3217
  parent pid                        1
  uptime                            57m 
  children                          0
  memory kilobytes                  15396
  memory kilobytes total            15396
  memory percent                    1.5%
  memory percent total              1.5%
  cpu percent                       0.0%
  cpu percent total                 0.0%
  data collected                    Tue May 27 14:18:45 2025

Process 'node_exporter'
  status                            running
  monitoring status                 monitored
  pid                               3236
  parent pid                        1
  uptime                            57m 
  children                          0
  memory kilobytes                  20192
  memory kilobytes total            20192
  memory percent                    2.0%
  memory percent total              2.0%
  cpu percent                       0.0%
  cpu percent total                 0.0%
  data collected                    Tue May 27 14:18:45 2025

System 'system_62f5a758-a81e-473e-571c-3ebe27d904b9'
  status                            running
  monitoring status                 monitored
  load average                      [0.00] [0.06] [0.10]
  cpu                               1.8%us 1.3%sy 0.1%wa
  memory usage                      438460 kB [44.7%]
  swap usage                        3328 kB [0.3%]
  data collected                    Tue May 27 14:18:45 2025
`
	fetcher := NewMonitFetcher("fake-path")
	stat, err := fetcher.parseData(sampleOutput)
	assert.NoError(t, err, "Fetch should complete without error")

	blackbox, exists := stat.Processes["blackbox"]
	assert.True(t, exists, "entry for process 'blackbox' should exist")
	assert.Equal(t, "running", blackbox.Status)

	boshiExporter, exists := stat.Processes["boshi_exporter"]
	assert.True(t, exists, "entry for process 'boshi_exporter' should exist")
	assert.Equal(t, "running", boshiExporter.Status)

	nodeExporter, exists := stat.Processes["node_exporter"]
	assert.True(t, exists, "entry for process 'node_exporter' should exist")
	assert.Equal(t, "running", nodeExporter.Status)

	// Verify system entry
	sys := stat.System
	assert.True(t, exists, "system entry should exist")
	assert.Equal(t, "running", sys.Status)
}

func TestMonitFetcher_ParsesDirectorOutput(t *testing.T) {
	sampleOutput := `The Monit daemon 5.2.5 uptime: 38m 

Process 'nats'
  status                            running
  monitoring status                 monitored
  pid                               22394
  parent pid                        1
  uptime                            38m 
  children                          1
  memory kilobytes                  900
  memory kilobytes total            35408
  memory percent                    0.0%
  memory percent total              0.8%
  cpu percent                       0.0%
  cpu percent total                 0.4%
  data collected                    Tue May 27 14:17:05 2025

Process 'bosh_nats_sync'
  status                            running
  monitoring status                 monitored
  pid                               23768
  parent pid                        1
  uptime                            37m 
  children                          1
  memory kilobytes                  884
  memory kilobytes total            43848
  memory percent                    0.0%
  memory percent total              1.0%
  cpu percent                       0.0%
  cpu percent total                 1.9%
  data collected                    Tue May 27 14:17:05 2025

Process 'postgres'
  status                            running
  monitoring status                 monitored
  pid                               22476
  parent pid                        1
  uptime                            38m 
  children                          65
  memory kilobytes                  904
  memory kilobytes total            1392724
  memory percent                    0.0%
  memory percent total              34.7%
  cpu percent                       0.0%
  cpu percent total                 0.8%
  data collected                    Tue May 27 14:17:05 2025

Process 'blobstore_nginx'
  status                            running
  monitoring status                 monitored
  pid                               22521
  parent pid                        1
  uptime                            38m 
  children                          3
  memory kilobytes                  880
  memory kilobytes total            22688
  memory percent                    0.0%
  memory percent total              0.5%
  cpu percent                       0.0%
  cpu percent total                 0.0%
  data collected                    Tue May 27 14:17:05 2025

Process 'director'
  status                            running
  monitoring status                 monitored
  pid                               22586
  parent pid                        1
  uptime                            38m 
  children                          4
  memory kilobytes                  904
  memory kilobytes total            575060
  memory percent                    0.0%
  memory percent total              14.3%
  cpu percent                       0.0%
  cpu percent total                 2.7%
  data collected                    Tue May 27 14:17:05 2025

Process 'worker_1'
  status                            running
  monitoring status                 monitored
  pid                               22596
  parent pid                        1
  uptime                            38m 
  children                          0
  memory kilobytes                  67344
  memory kilobytes total            67344
  memory percent                    1.6%
  memory percent total              1.6%
  cpu percent                       0.0%
  cpu percent total                 0.0%
  data collected                    Tue May 27 14:17:05 2025

Process 'worker_2'
  status                            running
  monitoring status                 monitored
  pid                               22606
  parent pid                        1
  uptime                            38m 
  children                          0
  memory kilobytes                  69308
  memory kilobytes total            69308
  memory percent                    1.7%
  memory percent total              1.7%
  cpu percent                       0.0%
  cpu percent total                 0.0%
  data collected                    Tue May 27 14:17:05 2025

Process 'director_scheduler'
  status                            running
  monitoring status                 monitored
  pid                               22870
  parent pid                        1
  uptime                            38m 
  children                          1
  memory kilobytes                  876
  memory kilobytes total            71800
  memory percent                    0.0%
  memory percent total              1.7%
  cpu percent                       0.0%
  cpu percent total                 0.0%
  data collected                    Tue May 27 14:17:05 2025

Process 'metrics_server'
  status                            running
  monitoring status                 monitored
  pid                               24820
  parent pid                        1
  uptime                            37m 
  children                          1
  memory kilobytes                  860
  memory kilobytes total            70664
  memory percent                    0.0%
  memory percent total              1.7%
  cpu percent                       0.0%
  cpu percent total                 0.0%
  data collected                    Tue May 27 14:17:05 2025

Process 'director_sync_dns'
  status                            running
  monitoring status                 monitored
  pid                               22958
  parent pid                        1
  uptime                            38m 
  children                          1
  memory kilobytes                  952
  memory kilobytes total            62184
  memory percent                    0.0%
  memory percent total              1.5%
  cpu percent                       0.0%
  cpu percent total                 0.0%
  data collected                    Tue May 27 14:17:05 2025

Process 'director_nginx'
  status                            running
  monitoring status                 monitored
  pid                               23026
  parent pid                        1
  uptime                            38m 
  children                          3
  memory kilobytes                  856
  memory kilobytes total            26664
  memory percent                    0.0%
  memory percent total              0.6%
  cpu percent                       0.0%
  cpu percent total                 1.3%
  data collected                    Tue May 27 14:17:05 2025

Process 'health_monitor'
  status                            running
  monitoring status                 monitored
  pid                               24871
  parent pid                        1
  uptime                            37m 
  children                          1
  memory kilobytes                  860
  memory kilobytes total            60560
  memory percent                    0.0%
  memory percent total              1.5%
  cpu percent                       0.0%
  cpu percent total                 0.4%
  port response time                0.005s to localhost:25923/healthz [HTTP via TCP]
  data collected                    Tue May 27 14:17:05 2025

Process 'uaa'
  status                            running
  monitoring status                 monitored
  pid                               23117
  parent pid                        1
  uptime                            38m 
  children                          2
  memory kilobytes                  860
  memory kilobytes total            468668
  memory percent                    0.0%
  memory percent total              11.7%
  cpu percent                       0.0%
  cpu percent total                 3.9%
  data collected                    Tue May 27 14:17:05 2025

Process 'credhub'
  status                            running
  monitoring status                 monitored
  pid                               24891
  parent pid                        1
  uptime                            37m 
  children                          12
  memory kilobytes                  458224
  memory kilobytes total            474112
  memory percent                    11.4%
  memory percent total              11.8%
  cpu percent                       0.4%
  cpu percent total                 0.4%
  data collected                    Tue May 27 14:17:05 2025

System 'system_9cd7df3f-27dc-4614-6df9-a542b2c43ba9'
  status                            running
  monitoring status                 monitored
  load average                      [2.01] [2.48] [2.62]
  cpu                               11.0%us 3.1%sy 0.4%wa
  memory usage                      3142212 kB [78.4%]
  swap usage                        254976 kB [6.3%]
  data collected                    Tue May 27 14:17:05 2025
`
	fetcher := NewMonitFetcher("fake-path")
	stat, err := fetcher.parseData(sampleOutput)
	assert.NoError(t, err, "Fetch should complete without error")

	director, exists := stat.Processes["director"]
	assert.True(t, exists, "entry for process 'director' should exist")
	assert.Equal(t, "running", director.Status)

	worker1, exists := stat.Processes["worker_1"]
	assert.True(t, exists, "entry for process 'worker_1' should exist")
	assert.Equal(t, "running", worker1.Status)

	worker2, exists := stat.Processes["worker_2"]
	assert.True(t, exists, "entry for process 'worker_2' should exist")
	assert.Equal(t, "running", worker2.Status)

	// Verify system entry
	sys := stat.System
	assert.True(t, exists, "system entry should exist")
	assert.Equal(t, "running", sys.Status)
}
