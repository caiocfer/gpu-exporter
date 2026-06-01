# GPU Process Exporter

NVIDIA GPU metrics exporter for Prometheus with per-process VRAM, CPU, and memory correlation.

See [`spec/design.md`](spec/design.md) for full architecture, deployment, and development plan.

## Quick Start

### Native
```bash
go build -o gpu-exporter ./cmd/gpu-exporter/
./gpu-exporter
curl http://localhost:9835/metrics
```

### Docker
```bash
docker build -t gpu-exporter .
docker run --rm --gpus all -p 9835:9835 --pid=host gpu-exporter
```

`--pid=host` is required for per-process metrics — the exporter reads host `/proc` to correlate GPU PIDs with process names and resource usage.

### Kubernetes (Helm)
```bash
helm upgrade --install gpu-exporter deploy/helm/gpu-exporter/
```

Requires NVIDIA drivers on host nodes and `runtimeClassName: nvidia` if configured (set via `--set nvidia.runtimeClassName=nvidia`).

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--addr` | `:9835` | Listen address |
| `--path` | `/metrics` | Metrics path |
| `--cache-ttl` | `60s` | Process metadata cache TTL |
| `--proc` | `/proc` | Path to proc filesystem |

## Environment

| Variable | Description |
| :--- | :--- |
| `NVML_LIB_PATH` | Explicit path to `libnvidia-ml.so` (bypasses `LD_LIBRARY_PATH`) |

## Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `gpu_utilization_percent` | Gauge | `gpu_id`, `gpu_uuid`, `gpu_name` | GPU compute utilization |
| `gpu_memory_usage_bytes` | Gauge | `gpu_id`, `gpu_uuid`, `gpu_name` | VRAM currently in use |
| `gpu_memory_total_bytes` | Gauge | `gpu_id`, `gpu_uuid`, `gpu_name` | Total VRAM available |
| `gpu_temperature_celsius` | Gauge | `gpu_id`, `gpu_uuid`, `gpu_name` | GPU temperature |
| `gpu_power_usage_milliwatts` | Gauge | `gpu_id`, `gpu_uuid`, `gpu_name` | GPU power draw |
| `gpu_fan_speed_percent` | Gauge | `gpu_id`, `gpu_uuid`, `gpu_name` | GPU fan speed |
| `gpu_process_memory_bytes` | Gauge | `gpu_id`, `gpu_uuid`, `gpu_name`, `pid`, `process_name` | VRAM per PID |
| `gpu_process_cpu_seconds_total` | Counter | `gpu_id`, `gpu_uuid`, `gpu_name`, `pid`, `process_name` | CPU time per process |
| `gpu_process_ram_usage_bytes` | Gauge | `gpu_id`, `gpu_uuid`, `gpu_name`, `pid`, `process_name` | System RAM per process |

## How It Works

1. NVML enumerates GPUs and their running processes (compute + graphics).
2. `/proc/<pid>/comm`, `stat`, and `status` are read for process name, CPU time, and RSS.
3. Process metadata (name) is cached with configurable TTL to reduce filesystem reads.
4. A merged view of compute and graphics GPU processes is exposed via a single `/metrics` endpoint.
