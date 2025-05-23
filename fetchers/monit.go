package fetchers

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// MonitFetcher invokes `monit status`
type MonitFetcher struct {
	monitPath  string // /var/vcap/bosh/bin/monit
	kvSplitter *regexp.Regexp
	reProcess  *regexp.Regexp
	reSystem   *regexp.Regexp
}

// MonitProcessStatus represents the status of a single process or the system
type MonitProcessStatus struct {
	Status           string // process status (e.g., running)
	MonitoringStatus string // whether monitoring is active (e.g., monitored)
	PID              string // process ID
	ParentPID        string // parent process ID

	Uptime             time.Duration // uptime since last start
	Children           int           // number of child processes
	MemoryBytes        int           // memory usage in bytes
	MemoryBytesTotal   int           // total memory usage in bytes
	MemoryPercent      float64       // memory usage as a percentage
	MemoryPercentTotal float64       // total memory percentage
	CPUPercent         float64       // CPU usage as a percentage
	CPUPercentTotal    float64       // total CPU percentage
	DataCollected      time.Time     // timestamp when data was collected
}

// MonitStat maps process or system names to their status
type MonitStat map[string]MonitProcessStatus

// NewMonitFetcher creates a new MonitFetcher with the given path to the monit binary
func NewMonitFetcher(monitPath string) *MonitFetcher {
	return &MonitFetcher{
		monitPath: monitPath,
		// regex to split on two or more spaces (separates key and value)
		kvSplitter: regexp.MustCompile(`\s{2,}`),
		// regex to detect process sections
		reProcess: regexp.MustCompile(`^Process '(.+)'`),
		// regex to detect a system section
		reSystem: regexp.MustCompile(`^System '(.+)'`),
	}
}

// Fetch parses the output of `monit status` and returns a MonitStat map
func (m *MonitFetcher) Fetch(ctx context.Context) (MonitStat, error) {
	// Execute the monit command
	out, err := exec.CommandContext(ctx, m.monitPath, "status").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute monit: %w", err)
	}
	return m.parseData(string(out))
}

func (m *MonitFetcher) parseData(data string) (MonitStat, error) {
	scanner := bufio.NewScanner(strings.NewReader(data))

	statusMap := make(MonitStat)
	var currentName string
	var currentStatus MonitProcessStatus

	// commit adds the currentStatus to statusMap under currentName
	commit := func() {
		if currentName != "" {
			statusMap[currentName] = currentStatus
		}
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// new process section
		if sm := m.reProcess.FindStringSubmatch(line); sm != nil {
			commit()
			currentName = sm[1]
			currentStatus = MonitProcessStatus{}
			continue
		}
		// system section
		if sm := m.reSystem.FindStringSubmatch(line); sm != nil {
			commit()
			currentName = sm[1]
			currentStatus = MonitProcessStatus{}
			continue
		}
		// split into key and value
		parts := m.kvSplitter.Split(line, 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch key {
		case "status":
			currentStatus.Status = val
		case "monitoring status":
			currentStatus.MonitoringStatus = val
		case "pid":
			currentStatus.PID = val
		case "parent pid":
			currentStatus.ParentPID = val
		case "uptime":
			if d, err := parseUptime(val); err == nil {
				currentStatus.Uptime = d
			}
		case "children":
			currentStatus.Children = atoi(val)
		case "memory kilobytes":
			currentStatus.MemoryBytes = atoi(val) * 1024
		case "memory kilobytes total":
			currentStatus.MemoryBytesTotal = atoi(val) * 1024
		case "memory percent":
			currentStatus.MemoryPercent = parsePercent(val)
		case "memory percent total":
			currentStatus.MemoryPercentTotal = parsePercent(val)
		case "cpu percent":
			currentStatus.CPUPercent = parsePercent(val)
		case "cpu percent total":
			currentStatus.CPUPercentTotal = parsePercent(val)
		case "data collected":
			if t, err := time.Parse("Mon Jan 2 15:04:05 2006", val); err == nil {
				currentStatus.DataCollected = t
			}
		}
	}
	// commit the last section
	commit()

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning data: %w", err)
	}
	return statusMap, nil
}

// atoi converts string to int, returns 0 on error
func atoi(s string) int {
	i, _ := strconv.Atoi(strings.TrimSpace(s))
	return i
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
