package types

// GPU types match Bridge's GpuStatus interface exactly
type GpuProcess struct {
	PID        int    `json:"pid"`
	Name       string `json:"name"`
	MemoryUsed int    `json:"memory_used"` // MiB
}

type GpuStatus struct {
	Name              string       `json:"name"`
	UtilizationGPU    int          `json:"utilization_gpu"`    // percent
	UtilizationMemory int          `json:"utilization_memory"` // percent
	MemoryTotal       int          `json:"memory_total"`       // MiB
	MemoryUsed        int          `json:"memory_used"`        // MiB
	MemoryFree        int          `json:"memory_free"`        // MiB
	Temperature       int          `json:"temperature"`        // Celsius
	Processes         []GpuProcess `json:"processes"`
}

type VramStatus struct {
	Total int `json:"total"` // MiB
	Used  int `json:"used"`  // MiB
	Free  int `json:"free"`  // MiB
}

type CpuStatus struct {
	LoadAvg1   float64   `json:"load_avg_1"`
	LoadAvg5   float64   `json:"load_avg_5"`
	LoadAvg15  float64   `json:"load_avg_15"`
	Cores      int       `json:"cores"`
	ModelName  string    `json:"model_name"`
	UsageTotal float64   `json:"usage_total"` // percent
	PerCore    []float64 `json:"per_core"`    // percent per core
}

type MemoryStatus struct {
	Total     int `json:"total"`     // MB
	Used      int `json:"used"`      // MB
	Free      int `json:"free"`      // MB
	Available int `json:"available"` // MB
	Cached    int `json:"cached"`    // MB
	SwapTotal int `json:"swap_total"`
	SwapUsed  int `json:"swap_used"`
	SwapFree  int `json:"swap_free"`
}

type DiskMount struct {
	MountPoint   string  `json:"mount_point"`
	Filesystem   string  `json:"filesystem"`
	Total        uint64  `json:"total"`        // bytes
	Used         uint64  `json:"used"`         // bytes
	Free         uint64  `json:"free"`         // bytes
	UsagePercent float64 `json:"usage_percent"`
}

type DiskStatus struct {
	Mounts []DiskMount `json:"mounts"`
}

type ServiceInfo struct {
	Name          string `json:"name"`
	Type          string `json:"type"`
	Enabled       bool   `json:"enabled"`
	Status        string `json:"status"`
	ContainerName string `json:"container_name"`
	VramRequired  int    `json:"vram_required"`
}

type HealthResponse struct {
	Status   string `json:"status"`
	Version  string `json:"version"`
	Hostname string `json:"hostname"`
	Uptime   int64  `json:"uptime"` // seconds
}

type SystemResponse struct {
	Health   HealthResponse `json:"health"`
	GPU      *GpuStatus     `json:"gpu,omitempty"`
	Vram     *VramStatus    `json:"vram,omitempty"`
	CPU      *CpuStatus     `json:"cpu,omitempty"`
	Memory   *MemoryStatus  `json:"memory,omitempty"`
	Disk     *DiskStatus    `json:"disk,omitempty"`
	Services []ServiceInfo  `json:"services,omitempty"`
}
