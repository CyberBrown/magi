package docker

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/cyberbrown/magi/types"
)

var client = &http.Client{
	Transport: &http.Transport{
		DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", "/var/run/docker.sock")
		},
	},
	Timeout: 5 * time.Second,
}

type container struct {
	Names  []string          `json:"Names"`
	State  string            `json:"State"`
	Status string            `json:"Status"`
	Labels map[string]string `json:"Labels"`
}

// Known services with their VRAM requirements (MiB)
var knownServices = map[string]serviceInfo{
	"nemotron":      {Type: "llm", VramRequired: 48000},
	"vllm":          {Type: "llm", VramRequired: 48000},
	"open-webui":    {Type: "ui", VramRequired: 0},
	"claude-runner": {Type: "agent", VramRequired: 0},
	"gemini-runner": {Type: "agent", VramRequired: 0},
	"reauth-ui":     {Type: "ui", VramRequired: 0},
	"ollama":        {Type: "llm", VramRequired: 8000},
}

type serviceInfo struct {
	Type         string
	VramRequired int
}

// Collect lists running Docker containers as ServiceInfo.
func Collect() ([]types.ServiceInfo, error) {
	resp, err := client.Get("http://localhost/containers/json?all=true")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var containers []container
	if err := json.NewDecoder(resp.Body).Decode(&containers); err != nil {
		return nil, err
	}

	var services []types.ServiceInfo
	for _, c := range containers {
		name := ""
		if len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}

		svc := types.ServiceInfo{
			Name:          name,
			ContainerName: name,
			Status:        c.State,
			Enabled:       true,
			Type:          "service",
		}

		if info, ok := knownServices[name]; ok {
			svc.Type = info.Type
			svc.VramRequired = info.VramRequired
		}

		services = append(services, svc)
	}

	return services, nil
}
