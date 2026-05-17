package api

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/cyberbrown/magi/internal/collector"
	"github.com/cyberbrown/magi/types"
)

var (
	Version   = "dev"
	StartTime = time.Now()
)

type Handlers struct {
	col *collector.Collector
}

func NewHandlers(col *collector.Collector) *Handlers {
	return &Handlers{col: col}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	hostname, _ := os.Hostname()
	writeJSON(w, types.HealthResponse{
		Status:   "ok",
		Version:  Version,
		Hostname: hostname,
		Uptime:   int64(time.Since(StartTime).Seconds()),
	})
}

func (h *Handlers) GPU(w http.ResponseWriter, r *http.Request) {
	g := h.col.GPU()
	if g == nil {
		http.Error(w, `{"error":"no gpu data"}`, http.StatusServiceUnavailable)
		return
	}
	writeJSON(w, g)
}

func (h *Handlers) Vram(w http.ResponseWriter, r *http.Request) {
	v := h.col.Vram()
	if v == nil {
		http.Error(w, `{"error":"no vram data"}`, http.StatusServiceUnavailable)
		return
	}
	writeJSON(w, v)
}

func (h *Handlers) CPU(w http.ResponseWriter, r *http.Request) {
	c := h.col.CPU()
	if c == nil {
		http.Error(w, `{"error":"no cpu data"}`, http.StatusServiceUnavailable)
		return
	}
	writeJSON(w, c)
}

func (h *Handlers) Memory(w http.ResponseWriter, r *http.Request) {
	m := h.col.Memory()
	if m == nil {
		http.Error(w, `{"error":"no memory data"}`, http.StatusServiceUnavailable)
		return
	}
	writeJSON(w, m)
}

func (h *Handlers) Disk(w http.ResponseWriter, r *http.Request) {
	d := h.col.Disk()
	if d == nil {
		http.Error(w, `{"error":"no disk data"}`, http.StatusServiceUnavailable)
		return
	}
	writeJSON(w, d)
}

func (h *Handlers) Services(w http.ResponseWriter, r *http.Request) {
	svcs := h.col.Services()
	if svcs == nil {
		svcs = []types.ServiceInfo{}
	}
	writeJSON(w, svcs)
}

func (h *Handlers) System(w http.ResponseWriter, r *http.Request) {
	hostname, _ := os.Hostname()
	writeJSON(w, types.SystemResponse{
		Health: types.HealthResponse{
			Status:   "ok",
			Version:  Version,
			Hostname: hostname,
			Uptime:   int64(time.Since(StartTime).Seconds()),
		},
		GPU:      h.col.GPU(),
		Vram:     h.col.Vram(),
		CPU:      h.col.CPU(),
		Memory:   h.col.Memory(),
		Disk:     h.col.Disk(),
		Services: h.col.Services(),
	})
}
