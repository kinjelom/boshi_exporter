package fetchers

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// MonitFetcher invokes `monit status`
type MonitFetcher struct {
	monitPath string // /var/vcap/bosh/bin/monit
	reBanner  *regexp.Regexp
	reSection *regexp.Regexp
	reMetric  *regexp.Regexp
}

// MonitProcessStatus represents the status of a single process or the system
type MonitProcessStatus struct {
	Status           string // service status (e.g., running)
	MonitoringStatus string // whether monitoring is active (e.g., monitored)
	PID              string // process ID
	ParentPID        string // parent process ID

	Uptime                 time.Duration // uptime since last start
	Children               int           // number of child processes
	MemoryUsedBytes        uint64        // memory used in bytes
	MemoryUsedBytesTotal   uint64        // total memory usage in bytes
	MemoryUsedPercent      float64       // memory usage as a percentage
	MemoryUsedPercentTotal float64       // total memory percentage
	CPUUsedPercent         float64       // CPU usage as a percentage
	CPUUsedPercentTotal    float64       // total CPU percentage

	DataCollected time.Time // timestamp when data was collected
}

// MonitSystemStatus holds parsed metrics from “bosh monit”
type MonitSystemStatus struct {
	Status           string // service status (e.g., "running")
	MonitoringStatus string // whether monitoring is active (e.g., "monitored")

	LoadAvg1          float64   // 1-minute load average
	LoadAvg5          float64   // 5-minute load average
	LoadAvg15         float64   // 15-minute load average
	CPUUserPercent    float64   // CPU time spent in user mode (%)
	CPUSystemPercent  float64   // CPU time spent in kernel/system mode (%)
	CPUIOWaitPercent  float64   // CPU time spent waiting for I/O (%)
	MemoryUsedBytes   uint64    // memory used in bytes
	MemoryUsedPercent float64   // memory used as a percentage of total
	SwapUsedBytes     uint64    // swap used in bytes
	SwapUsedPercent   float64   // swap used as a percentage of total
	DataCollected     time.Time // timestamp when data was collected
}

// MonitStat maps process or system names to their status
type MonitStat struct {
	Version   string
	Uptime    time.Duration
	Processes map[string]MonitProcessStatus
	System    MonitSystemStatus
}

// NewMonitFetcher creates a new MonitFetcher with the given path to the monit binary
func NewMonitFetcher(monitPath string) *MonitFetcher {
	return &MonitFetcher{
		monitPath: monitPath,
		// regex to check monit version, uptime
		reBanner: regexp.MustCompile(`^The Monit daemon (\S+) uptime: (.*)$`),
		// regex to detect section start
		reSection: regexp.MustCompile(`^(\S+) '(.*)'$`),
		// regex to detect process sections
		reMetric: regexp.MustCompile(`^\s{2}(\S+(\s\S+)*)\s{2,}(.*)$`),
	}
}

// Fetch parses the output of `monit status` and returns a MonitStat map
func (m *MonitFetcher) Fetch(ctx context.Context) (stat *MonitStat, err error) {
	const timeoutSec = 5
	ctx, cancel := context.WithTimeout(ctx, timeoutSec*time.Second)
	defer cancel()

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic during Fetch: %v", r)
		}
	}()

	cmd := exec.CommandContext(ctx, m.monitPath, "status")
	output, execErr := cmd.CombinedOutput()
	if execErr != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, fmt.Errorf("monit status timed out after %d s: %w", timeoutSec, execErr)
		}
		return nil, fmt.Errorf("failed to execute monit: %w (output: %s)", execErr, strings.TrimSpace(string(output)))
	}

	stat, parseErr := m.parseData(string(output))
	if parseErr != nil {
		return nil, fmt.Errorf("failed to parse monit output: %w", parseErr)
	}

	return stat, nil
}

func (m *MonitFetcher) parseData(data string) (stat *MonitStat, err error) {
	stat = &MonitStat{
		Processes: make(map[string]MonitProcessStatus),
	}
	scanner := bufio.NewScanner(strings.NewReader(data))
	bannerLine := ""
	for scanner.Scan() {
		bannerLine = strings.TrimSpace(scanner.Text())
		if bannerLine != "" {
			break
		}
	}
	if banner := m.reBanner.FindStringSubmatch(bannerLine); banner != nil {
		if len(banner) == 3 {
			stat.Version = banner[1]
			if d, err := parseUptime(banner[2]); err == nil {
				stat.Uptime = d
			}
		}
	}
	if stat.Version == "" {
		return nil, fmt.Errorf("unsupported monit output: %s", bannerLine)
	}

	var currentServiceKind string
	var currentServiceName string
	var currentProcessStatus MonitProcessStatus
	var currentSystemStatus MonitSystemStatus
	commitService := func() {
		if currentServiceName != "" {
			if currentServiceKind == "Process" {
				stat.Processes[currentServiceName] = currentProcessStatus
			}
			if currentServiceKind == "System" {
				stat.System = currentSystemStatus
			}
		}
		currentProcessStatus = MonitProcessStatus{}
		currentSystemStatus = MonitSystemStatus{}
	}
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) != "" {
			if section := m.reSection.FindStringSubmatch(line); section != nil {
				commitService()
				if len(section) == 3 {
					currentServiceKind = section[1]
					currentServiceName = section[2]
				}
			} else if metric := m.reMetric.FindStringSubmatch(line); metric != nil {
				if len(metric) == 4 {
					key := strings.TrimSpace(metric[1])
					val := strings.TrimSpace(metric[3])
					if currentServiceKind == "Process" {
						currentProcessStatus.parseMetricEntry(key, val)
					} else if currentServiceKind == "System" {
						currentSystemStatus.parseMetricEntry(key, val)
					}
				}
			}
		}
	}
	commitService()

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning data: %w", err)
	}
	return stat, nil
}

func (p *MonitProcessStatus) parseMetricEntry(key, val string) {
	switch key {
	case "status":
		p.Status = val
	case "monitoring status":
		p.MonitoringStatus = val
	case "pid":
		p.PID = val
	case "parent pid":
		p.ParentPID = val
	case "uptime":
		if d, err := parseUptime(val); err == nil {
			p.Uptime = d
		}
	case "children":
		p.Children = atoi(val)
	case "memory kilobytes":
		p.MemoryUsedBytes = atouint(val) * 1024
	case "memory kilobytes total":
		p.MemoryUsedBytesTotal = atouint(val) * 1024
	case "memory percent":
		p.MemoryUsedPercent = parsePercent(val)
	case "memory percent total":
		p.MemoryUsedPercentTotal = parsePercent(val)
	case "cpu percent":
		p.CPUUsedPercent = parsePercent(val)
	case "cpu percent total":
		p.CPUUsedPercentTotal = parsePercent(val)
	case "data collected":
		if t, err := time.Parse("Mon Jan 2 15:04:05 2006", val); err == nil {
			p.DataCollected = t
		}
	}
}

// parseMetricEntry maps a single “key: value” line into struct fields
func (s *MonitSystemStatus) parseMetricEntry(key, val string) {
	switch key {
	case "status":
		s.Status = val

	case "monitoring status":
		s.MonitoringStatus = val

	case "load average":
		// val example: "[0.09] [0.10] [0.04]"
		_, _ = fmt.Sscanf(val, "[%f] [%f] [%f]", &s.LoadAvg1, &s.LoadAvg5, &s.LoadAvg15)

	case "cpu":
		// val example: "2.0%us 4.0%sy 0.0%wa"
		_, _ = fmt.Sscanf(val, "%f%%us %f%%sy %f%%wa",
			&s.CPUUserPercent, &s.CPUSystemPercent, &s.CPUIOWaitPercent)

	case "memory usage":
		// val example: "227240 kB [23.0%]"
		var kb uint64
		_, _ = fmt.Sscanf(val, "%d kB [%f%%]", &kb, &s.MemoryUsedPercent)
		s.MemoryUsedBytes = kb * 1024

	case "swap usage":
		// val example: "256 kB [0.0%]"
		var kb uint64
		_, _ = fmt.Sscanf(val, "%d kB [%f%%]", &kb, &s.SwapUsedPercent)
		s.SwapUsedBytes = kb * 1024

	case "data collected":
		// val example: "Fri May 23 11:03:35 2025"
		if t, err := time.Parse("Mon Jan 2 15:04:05 2006", val); err == nil {
			s.DataCollected = t
		}
	}
}

// atoi converts string to int, returns 0 on error
func atoi(s string) int {
	i, _ := strconv.Atoi(strings.TrimSpace(s))
	return i
}
func atouint(s string) uint64 {
	i, _ := strconv.Atoi(strings.TrimSpace(s))
	return uint64(i)
}

// parsePercent removes '%' and converts to float64, returns 0 on error
func parsePercent(s string) float64 {
	s = strings.TrimSuffix(s, "%")
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// parseUptime converts a string like "2h12m" or "19h17m" to time.Duration
func parseUptime(s string) (time.Duration, error) {
	// remove spaces to match time.ParseDuration format
	s = strings.ReplaceAll(s, " ", "")
	return time.ParseDuration(s)
}
