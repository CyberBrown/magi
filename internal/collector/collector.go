package collector

import (
	"log"
	"sync"
	"time"

	"github.com/cyberbrown/magi/internal/cpu"
	"github.com/cyberbrown/magi/internal/disk"
	"github.com/cyberbrown/magi/internal/docker"
	"github.com/cyberbrown/magi/internal/gpu"
	"github.com/cyberbrown/magi/internal/memory"
	"github.com/cyberbrown/magi/types"
)

// Collector runs background data collection and caches snapshots.
type Collector struct {
	mu       sync.RWMutex
	gpu      *types.GpuStatus
	cpuStat  *types.CpuStatus
	mem      *types.MemoryStatus
	diskStat *types.DiskStatus
	services []types.ServiceInfo
}

func New() *Collector {
	return &Collector{}
}

// Start begins periodic collection. Call cpu.Init() before this.
func (c *Collector) Start(interval time.Duration) {
	c.collect() // initial collection
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			c.collect()
		}
	}()
}

func (c *Collector) collect() {
	g, err := gpu.Collect()
	if err != nil {
		log.Printf("gpu collect: %v", err)
	}

	cp, err := cpu.Collect()
	if err != nil {
		log.Printf("cpu collect: %v", err)
	}

	m, err := memory.Collect()
	if err != nil {
		log.Printf("memory collect: %v", err)
	}

	d, err := disk.Collect()
	if err != nil {
		log.Printf("disk collect: %v", err)
	}

	svcs, err := docker.Collect()
	if err != nil {
		log.Printf("docker collect: %v", err)
	}

	c.mu.Lock()
	if g != nil {
		c.gpu = g
	}
	if cp != nil {
		c.cpuStat = cp
	}
	if m != nil {
		c.mem = m
	}
	if d != nil {
		c.diskStat = d
	}
	if svcs != nil {
		c.services = svcs
	}
	c.mu.Unlock()
}

func (c *Collector) GPU() *types.GpuStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.gpu
}

func (c *Collector) Vram() *types.VramStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.gpu == nil {
		return nil
	}
	return &types.VramStatus{
		Total: c.gpu.MemoryTotal,
		Used:  c.gpu.MemoryUsed,
		Free:  c.gpu.MemoryFree,
	}
}

func (c *Collector) CPU() *types.CpuStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cpuStat
}

func (c *Collector) Memory() *types.MemoryStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.mem
}

func (c *Collector) Disk() *types.DiskStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.diskStat
}

func (c *Collector) Services() []types.ServiceInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.services
}
