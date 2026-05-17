# MAGI

A lightweight Go agent that collects system metrics and exposes them over a token-authenticated JSON API. Designed for monitoring Linux servers with NVIDIA GPUs and Docker workloads.

Named after the Magi system from *Neon Genesis Evangelion* — three supercomputers that analyze, predict, and manage. This agent does the same for your hardware.

## What It Does

MAGI runs as a single binary background service. On a configurable interval (default 2s) it reads system state from `/proc`, `nvidia-smi`, and the Docker socket, caches the latest snapshot, and serves it as JSON over HTTP.

| Endpoint     | Auth | Description                                    |
|-------------|------|------------------------------------------------|
| `GET /health`  | No  | Service health, version, hostname, uptime      |
| `GET /system`  | Yes | All metrics in one response                    |
| `GET /cpu`     | Yes | Load averages, per-core usage, model name      |
| `GET /memory`  | Yes | RAM and swap usage (total/used/free/cached)     |
| `GET /disk`    | Yes | Filesystem usage for key mount points           |
| `GET /gpu`     | Yes | GPU name, utilization, VRAM, temperature, procs |
| `GET /vram`    | Yes | VRAM summary (total/used/free)                  |
| `GET /services`| Yes | Docker containers with service metadata         |

Authentication is via `Authorization: Bearer <token>` header.

## Architecture

```
cmd/magi/main.go          Entry point — configures routes, starts collector
api/
  handlers.go             HTTP handlers for each endpoint
  middleware.go            Bearer token auth + CORS middleware
types/
  types.go                Shared response structs
internal/
  collector/collector.go  Background goroutine that caches metric snapshots
  cpu/cpu.go              /proc/stat, /proc/loadavg, /proc/cpuinfo
  memory/memory.go        /proc/meminfo
  disk/disk.go            /proc/mounts + syscall.Statfs
  gpu/gpu.go              nvidia-smi CSV parsing (supports unified memory GPUs)
  docker/docker.go        Docker Engine API via unix socket
```

The collector pattern means HTTP handlers never block on metric collection — they just read the latest cached snapshot under an RWMutex.

## Requirements

- Linux (reads from `/proc` filesystem)
- Go 1.22+
- `nvidia-smi` on PATH (for GPU metrics; degrades gracefully if absent)
- Docker socket at `/var/run/docker.sock` (for container listing; degrades gracefully)

## Building

```bash
make build          # Build for current architecture
make build-arm64    # Cross-compile for ARM64 (e.g. Jetson, Raspberry Pi)
make test           # Run tests
make clean          # Remove binaries
```

## Running

Set environment variables and run:

```bash
export MAGI_TOKEN="your-secret-token"
export MAGI_PORT=8780        # optional, defaults to 8780
./magi
```

Or use the example env file:

```bash
cp .env.example .env
# edit .env with your token
```

### systemd

A service unit file (`magi.service`) is included for production deployment:

```bash
sudo cp magi /usr/local/bin/magi
sudo useradd -r -s /bin/false magi
sudo usermod -aG docker magi
sudo cp .env /etc/magi/env
sudo cp magi.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now magi
```

The service file includes security hardening: `NoNewPrivileges`, `ProtectSystem=strict`, `ProtectHome`, `PrivateTmp`, and read-only `/proc`/`/sys`.

### Deploy to Remote

```bash
make deploy    # builds ARM64, SCPs to host "prometheus", restarts service
```

## API Example

```bash
# Health check (no auth)
curl http://localhost:8780/health

# Full system snapshot
curl -H "Authorization: Bearer your-secret-token" http://localhost:8780/system

# GPU only
curl -H "Authorization: Bearer your-secret-token" http://localhost:8780/gpu
```

Sample `/system` response:

```json
{
  "health": {"status": "ok", "version": "dev", "hostname": "mir", "uptime": 3600},
  "gpu": {
    "name": "NVIDIA GeForce RTX 4090",
    "utilization_gpu": 42,
    "utilization_memory": 55,
    "memory_total": 24564,
    "memory_used": 13510,
    "memory_free": 11054,
    "temperature": 71,
    "processes": [{"pid": 1234, "name": "python3", "memory_used": 8192}]
  },
  "cpu": {
    "load_avg_1": 2.15, "load_avg_5": 1.87, "load_avg_15": 1.64,
    "cores": 16, "model_name": "AMD Ryzen 9 5950X",
    "usage_total": 34.5, "per_core": [45.2, 22.1, ...]
  },
  "memory": {"total": 65536, "used": 32768, "free": 8192, "available": 40960, "cached": 24576, "swap_total": 8192, "swap_used": 0, "swap_free": 8192},
  "disk": {"mounts": [{"mount_point": "/", "filesystem": "/dev/nvme0n1p2", "total": 1000204886016, "used": 450971566080, "free": 549232922624, "usage_percent": 45.1}]},
  "services": [{"name": "vllm", "type": "llm", "enabled": true, "status": "running", "container_name": "vllm", "vram_required": 48000}]
}
```

## Docker Service Metadata

The Docker collector has a built-in registry of known services with types and VRAM requirements. This lets dashboards understand what each container does and how much GPU memory it needs:

| Service         | Type   | VRAM Required |
|----------------|--------|---------------|
| nemotron        | llm    | 48000 MiB     |
| vllm            | llm    | 48000 MiB     |
| open-webui      | ui     | 0             |
| claude-runner   | agent  | 0             |
| gemini-runner   | agent  | 0             |
| reauth-ui       | ui     | 0             |
| ollama          | llm    | 8000 MiB      |

## Project Status

**Functional but pre-release.** No git commits yet — all files are staged in the index. Key observations:

- **Built and working:** CPU, memory, disk, GPU (including unified memory fallback), Docker collectors, HTTP API with auth, CORS, background caching collector, systemd unit, Makefile with cross-compile.
- **Types refactor in progress:** `api/types.go` was deleted from the working tree and types were moved to a new `types/types.go` package. The working tree reflects this refactor but `git add` has not been run to update the index.
- **No tests found:** No `_test.go` files exist yet, though `make test` is wired up.
- **No external dependencies:** `go.mod` has zero third-party imports — everything uses the standard library.

## License

Proprietary. All rights reserved.
