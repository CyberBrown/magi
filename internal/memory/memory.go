package memory

import (
	"os"
	"strconv"
	"strings"

	"github.com/cyberbrown/magi/types"
)

// Collect parses /proc/meminfo into MemoryStatus (values in MB).
func Collect() (*types.MemoryStatus, error) {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return nil, err
	}

	m := parseMeminfo(string(data))

	total := m["MemTotal"]
	free := m["MemFree"]
	available := m["MemAvailable"]
	cached := m["Cached"] + m["Buffers"]
	used := total - free - cached

	return &types.MemoryStatus{
		Total:     int(total / 1024), // kB → MB
		Used:      int(used / 1024),
		Free:      int(free / 1024),
		Available: int(available / 1024),
		Cached:    int(cached / 1024),
		SwapTotal: int(m["SwapTotal"] / 1024),
		SwapUsed:  int((m["SwapTotal"] - m["SwapFree"]) / 1024),
		SwapFree:  int(m["SwapFree"] / 1024),
	}, nil
}

func parseMeminfo(data string) map[string]int64 {
	m := make(map[string]int64)
	for _, line := range strings.Split(data, "\n") {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		valStr := strings.TrimSpace(parts[1])
		valStr = strings.TrimSuffix(valStr, " kB")
		valStr = strings.TrimSpace(valStr)
		val, _ := strconv.ParseInt(valStr, 10, 64)
		m[key] = val
	}
	return m
}
