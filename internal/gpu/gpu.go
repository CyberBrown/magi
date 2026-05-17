package gpu

import (
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/cyberbrown/magi/types"
)

// Collect runs nvidia-smi and parses CSV output into GpuStatus.
func Collect() (*types.GpuStatus, error) {
	out, err := exec.Command("nvidia-smi",
		"--query-gpu=name,utilization.gpu,utilization.memory,memory.total,memory.used,memory.free,temperature.gpu",
		"--format=csv,noheader,nounits",
	).Output()
	if err != nil {
		return nil, err
	}

	line := strings.TrimSpace(string(out))
	parts := strings.Split(line, ", ")
	if len(parts) < 7 {
		return nil, err
	}

	utilGPU, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
	utilMem, _ := strconv.Atoi(strings.TrimSpace(parts[2]))
	memTotal := parseIntOrZero(strings.TrimSpace(parts[3]))
	memUsed := parseIntOrZero(strings.TrimSpace(parts[4]))
	memFree := parseIntOrZero(strings.TrimSpace(parts[5]))
	temp, _ := strconv.Atoi(strings.TrimSpace(parts[6]))

	s := &types.GpuStatus{
		Name:              strings.TrimSpace(parts[0]),
		UtilizationGPU:    utilGPU,
		UtilizationMemory: utilMem,
		MemoryTotal:       memTotal,
		MemoryUsed:        memUsed,
		MemoryFree:        memFree,
		Temperature:       temp,
	}

	s.Processes, _ = collectProcesses()

	// On unified memory architectures (e.g. GB10), nvidia-smi reports [N/A]
	// for FB memory. Derive from process usage and system total memory.
	if s.MemoryTotal == 0 {
		// Sum process GPU memory as "used"
		var procMem int
		for _, p := range s.Processes {
			procMem += p.MemoryUsed
		}
		s.MemoryUsed = procMem
		// Read total from system memory (unified architecture shares all RAM)
		if total := readTotalSystemMemoryMiB(); total > 0 {
			s.MemoryTotal = total
			s.MemoryFree = total - procMem
			if s.MemoryTotal > 0 {
				s.UtilizationMemory = procMem * 100 / s.MemoryTotal
			}
		}
	}

	return s, nil
}

func collectProcesses() ([]types.GpuProcess, error) {
	out, err := exec.Command("nvidia-smi",
		"--query-compute-apps=pid,process_name,used_gpu_memory",
		"--format=csv,noheader,nounits",
	).Output()
	if err != nil {
		return nil, err
	}

	var procs []types.GpuProcess
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		parts := strings.Split(line, ", ")
		if len(parts) < 3 {
			continue
		}
		pid, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
		mem, _ := strconv.Atoi(strings.TrimSpace(parts[2]))
		procs = append(procs, types.GpuProcess{
			PID:        pid,
			Name:       strings.TrimSpace(parts[1]),
			MemoryUsed: mem,
		})
	}
	return procs, nil
}

// parseIntOrZero handles [N/A] values from nvidia-smi on unified memory GPUs.
func parseIntOrZero(s string) int {
	if strings.Contains(s, "N/A") {
		return 0
	}
	v, _ := strconv.Atoi(s)
	return v
}

// readTotalSystemMemoryMiB reads MemTotal from /proc/meminfo.
func readTotalSystemMemoryMiB() int {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				kb, _ := strconv.Atoi(fields[1])
				return kb / 1024 // kB → MiB
			}
		}
	}
	return 0
}
