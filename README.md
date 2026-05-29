# GPU Process Exporter

NVIDIA GPU metrics exporter for Prometheus with per-process VRAM, CPU, and memory correlation.

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

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--addr` | `:9835` | Listen address |
| `--path` | `/metrics` | Metrics path |
| `--cache-ttl` | `60s` | Process metadata cache TTL |
| `--proc` | `/proc` | Path to proc filesystem |

## Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `gpu_utilization_percent` | Gauge | `gpu_id`, `gpu_uuid` | GPU compute utilization |
| `gpu_memory_usage_bytes` | Gauge | `gpu_id`, `gpu_uuid` | Total VRAM in use |
| `gpu_process_memory_bytes` | Gauge | `gpu_id`, `gpu_uuid`, `pid`, `process_name` | VRAM per process |
| `gpu_process_cpu_seconds_total` | Counter | `gpu_id`, `gpu_uuid`, `pid`, `process_name` | CPU time per process |
| `gpu_process_ram_usage_bytes` | Gauge | `gpu_id`, `gpu_uuid`, `pid`, `process_name` | System RAM per process |

## How It Works

1. NVML enumerates GPUs and their running processes (compute + graphics).
2. `/proc/<pid>/comm`, `stat`, and `status` are read for process name, CPU time, and RSS.
3. Process metadata (name) is cached with configurable TTL to reduce filesystem reads.
4. A merged view of compute and graphics GPU processes is exposed via a single `/metrics` endpoint.
