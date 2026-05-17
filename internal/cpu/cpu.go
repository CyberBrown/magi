package cpu

import (
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/cyberbrown/magi/types"
)

type snapshot struct {
	total []uint64
	idle  []uint64
}

var prev *snapshot

// Collect reads /proc/stat and /proc/loadavg to build CpuStatus.
func Collect() (*types.CpuStatus, error) {
	stat, err := os.ReadFile("/proc/stat")
	if err != nil {
		return nil, err
	}

	cur := parseStat(string(stat))
	loadavg, _ := readLoadAvg()
	model := readModelName()

	s := &types.CpuStatus{
		LoadAvg1:  loadavg[0],
		LoadAvg5:  loadavg[1],
		LoadAvg15: loadavg[2],
		Cores:     runtime.NumCPU(),
		ModelName: model,
	}

	if prev != nil && len(cur.total) == len(prev.total) {
		s.UsageTotal = calcUsage(prev.total[0], prev.idle[0], cur.total[0], cur.idle[0])
		perCore := make([]float64, len(cur.total)-1)
		for i := 1; i < len(cur.total); i++ {
			perCore[i-1] = calcUsage(prev.total[i], prev.idle[i], cur.total[i], cur.idle[i])
		}
		s.PerCore = perCore
	}

	prev = cur
	return s, nil
}

func parseStat(data string) *snapshot {
	var totals []uint64
	var idles []uint64
	for _, line := range strings.Split(data, "\n") {
		if !strings.HasPrefix(line, "cpu") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		var total, idle uint64
		for i, f := range fields[1:] {
			v, _ := strconv.ParseUint(f, 10, 64)
			total += v
			if i == 3 { // idle is 4th field
				idle = v
			}
		}
		totals = append(totals, total)
		idles = append(idles, idle)
	}
	return &snapshot{total: totals, idle: idles}
}

func calcUsage(prevTotal, prevIdle, curTotal, curIdle uint64) float64 {
	totalDelta := curTotal - prevTotal
	idleDelta := curIdle - prevIdle
	if totalDelta == 0 {
		return 0
	}
	return float64(totalDelta-idleDelta) / float64(totalDelta) * 100
}

func readLoadAvg() ([3]float64, error) {
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return [3]float64{}, err
	}
	fields := strings.Fields(string(data))
	var result [3]float64
	for i := 0; i < 3 && i < len(fields); i++ {
		result[i], _ = strconv.ParseFloat(fields[i], 64)
	}
	return result, nil
}

func readModelName() string {
	data, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return "unknown"
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "model name") || strings.HasPrefix(line, "Model") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return "unknown"
}

// Init takes an initial CPU snapshot so the first Collect() has a baseline.
func Init() {
	stat, err := os.ReadFile("/proc/stat")
	if err != nil {
		return
	}
	prev = parseStat(string(stat))
	time.Sleep(100 * time.Millisecond)
}
