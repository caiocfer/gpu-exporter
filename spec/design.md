# Software Design Document: GPU Process Exporter

## 1. Introduction
The GPU Process Exporter is a lightweight, standalone Go application designed to collect and expose detailed NVIDIA GPU metrics, with a specific focus on per-process resource utilization. This tool enables observability into which processes are consuming VRAM and compute resources, filling the gap between raw hardware metrics and system-level process monitoring.

## 2. Goals
### 2.1 Core Objectives
- **Standalone Execution:** Compile to a single binary with no external dependencies (other than NVML libraries).
- **Prometheus Compatibility:** Expose metrics via the `/metrics` endpoint using the OpenMetrics/Prometheus exposition format.
- **NVIDIA Integration:** Utilize official NVIDIA NVML Go bindings for robust data collection.
- **Process Correlation:** Map GPU PIDs to OS-level PIDs via `/proc` to provide context (process name, CPU/RAM usage) alongside GPU VRAM usage.

### 2.2 Deployment
- **Helm chart** at `deploy/helm/gpu-exporter/` for Kubernetes (k3s) DaemonSet deployment.
- **Dockerfile** with multi-stage build (`golang:1.26-bookworm` → `gcr.io/distroless/cc-debian12`).
- **DaemonSet** runs on all GPU nodes with `hostPID: true`, `privileged: true`, and NVIDIA device/library hostPath mounts.

## 3. Technology Stack
- **Language:** Go 1.26+
- **Metrics Library:** `prometheus/client_golang`
- **NVIDIA Interface:** `github.com/NVIDIA/go-nvml/pkg/nvml`
- **System Interface:** Standard library `os` and `/proc` filesystem parsing.

## 4. Architecture Design

### 4.1 Scrape Lifecycle
1. **Init:** `nvml.Init()` → `nvml.GetDevices()` (enumerates GPUs, caches UUID + Name in `Device` struct).
2. **Registry:** `prometheus.NewRegistry()` registers `GlobalCollector` + `ProcessCollector` + Go runtime collectors.
3. **Scrape (per HTTP request to /metrics):**
   - `GlobalCollector.Collect()`:
     - For each device: utilization, memory (total/used), temperature, power, fan speed.
   - `ProcessCollector.Collect()`:
     - For each device: `GetMergedProcesses()` (compute + graphics deduped by PID).
     - For each PID: read `/proc/<pid>/comm` (name), `/proc/<pid>/stat` (CPU ticks), `/proc/<pid>/status` (RSS).
     - Process metadata cached with TTL (default 60s) to reduce `/proc` reads.
4. **Serve:** Standard `http.ServeMux` on `:9835/metrics`.

### 4.2 CLI Flags

| Flag | Default | Description |
| :--- | :--- | :--- |
| `--addr` | `:9835` | Listen address |
| `--path` | `/metrics` | Metrics HTTP path |
| `--cache-ttl` | `60s` | Process metadata cache TTL |
| `--proc` | `/proc` | Path to proc filesystem (for containerized use with `--pid=host`) |

### 4.3 Environment Variables

| Variable | Description |
| :--- | :--- |
| `NVML_LIB_PATH` | Explicit path to `libnvidia-ml.so` directory or file (bypasses `LD_LIBRARY_PATH`) |

### 4.4 Cache Strategy
- TTL-based cache for process name and cmdline (stable metadata).
- VRAM usage and CPU ticks fetched per scrape (volatile).
- Cache eviction on TTL expiry or PID reuse detection.

## 5. Metrics Schema (OpenMetrics)

| Metric Name | Type | Labels | Description |
| :--- | :--- | :--- | :--- |
| `gpu_utilization_percent` | Gauge | `gpu_id`, `gpu_uuid`, `gpu_name` | GPU compute unit utilization |
| `gpu_memory_usage_bytes` | Gauge | `gpu_id`, `gpu_uuid`, `gpu_name` | VRAM currently in use |
| `gpu_memory_total_bytes` | Gauge | `gpu_id`, `gpu_uuid`, `gpu_name` | Total VRAM available |
| `gpu_temperature_celsius` | Gauge | `gpu_id`, `gpu_uuid`, `gpu_name` | GPU temperature |
| `gpu_power_usage_milliwatts` | Gauge | `gpu_id`, `gpu_uuid`, `gpu_name` | GPU power draw |
| `gpu_fan_speed_percent` | Gauge | `gpu_id`, `gpu_uuid`, `gpu_name` | GPU fan speed |
| `gpu_process_memory_bytes` | Gauge | `gpu_id`, `gpu_uuid`, `gpu_name`, `pid`, `process_name` | VRAM per PID |
| `gpu_process_cpu_seconds_total` | Counter | `gpu_id`, `gpu_uuid`, `gpu_name`, `pid`, `process_name` | CPU time per GPU process |
| `gpu_process_ram_usage_bytes` | Gauge | `gpu_id`, `gpu_uuid`, `gpu_name`, `pid`, `process_name` | System RAM per GPU process |

## 6. Implementation Details

### 6.1 Process Correlation
```go
func getProcessDetails(pid uint32) (string, uint64, uint64) {
    // /proc/<pid>/comm → name
    // /proc/<pid>/stat → CPU ticks (utime + stime)
    // /proc/<pid>/status → VmRSS in bytes
    return name, cpuUsage, ramUsage
}
```

### 6.2 Labeling Strategy
- `gpu_id`: NVML device index (e.g., `"0"`).
- `gpu_uuid`: NVML device UUID (stable across reboots).
- `gpu_name`: NVML device name (e.g., `"NVIDIA GeForce RTX 4090"`).
- `pid`: OS process ID.
- `process_name`: from `/proc/<pid>/comm` (fallback to cmdline).

### 6.3 NVML Init
- `nvml.SetLibraryOptions(nvml.WithLibraryPath(path))` called before `nvml.Init()` when `NVML_LIB_PATH` is set.
- Enables explicit path resolution in distroless containers lacking `ldconfig`.

## 7. Deployment (Kubernetes)

### 7.1 Helm Chart Structure
```
deploy/helm/gpu-exporter/
├── Chart.yaml
├── templates/
│   ├── daemonset.yaml    # DaemonSet with hostPath mounts + privileged
│   ├── serviceaccount.yaml
│   └── service.yaml      # ClusterIP + prometheus.io scrape annotations
└── values.yaml            # Configurable devices, library paths, resources
```

### 7.2 Required Access
- `hostPID: true` — for `/proc` access to host processes.
- `privileged: true` — for cgroup v2 eBPF device access to `/dev/nvidia*`.
- NVIDIA device mounts: `nvidiactl`, `nvidia0`, `nvidia-uvm`, `nvidia-modeset`, `nvidia-uvm-tools`.
- NVIDIA library mounts: `libnvidia-ml.so`, `libnvidia-ml.so.1` to `/usr/lib/x86_64-linux-gnu/`.
- `LD_LIBRARY_PATH=/usr/lib:/usr/lib/x86_64-linux-gnu`.

### 7.3 Image
- Base: `gcr.io/distroless/cc-debian12` (~54 MB).
- Build: `golang:1.26-bookworm` with `CGO_ENABLED=1`.

## 8. Development Plan

| Phase | Status | Description |
| :--- | :--- | :--- |
| 1 | ✓ Done | Skeleton project and NVML initialization |
| 2 | ✓ Done | Global GPU metrics (utilization, memory, temp, power, fan) |
| 3 | ✓ Done | Per-process VRAM collection via NVML |
| 4 | ✓ Done | /proc filesystem scanning for CPU/RAM correlation |
| 5 | ✓ Done | Prometheus exposition via /metrics |
