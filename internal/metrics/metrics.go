package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	GpuUtilizationPercent = prometheus.NewDesc(
		"gpu_utilization_percent",
		"Percent of GPU compute units currently in use.",
		[]string{"gpu_id", "gpu_uuid", "gpu_name"},
		nil,
	)
	GpuMemoryUsageBytes = prometheus.NewDesc(
		"gpu_memory_usage_bytes",
		"Total VRAM in use on the device.",
		[]string{"gpu_id", "gpu_uuid", "gpu_name"},
		nil,
	)
	GpuMemoryTotalBytes = prometheus.NewDesc(
		"gpu_memory_total_bytes",
		"Total VRAM available on the device.",
		[]string{"gpu_id", "gpu_uuid", "gpu_name"},
		nil,
	)
	GpuTemperatureCelsius = prometheus.NewDesc(
		"gpu_temperature_celsius",
		"Current GPU temperature in Celsius.",
		[]string{"gpu_id", "gpu_uuid", "gpu_name"},
		nil,
	)
	GpuPowerUsageMilliwatts = prometheus.NewDesc(
		"gpu_power_usage_milliwatts",
		"Current GPU power usage in milliwatts.",
		[]string{"gpu_id", "gpu_uuid", "gpu_name"},
		nil,
	)
	GpuFanSpeedPercent = prometheus.NewDesc(
		"gpu_fan_speed_percent",
		"Current GPU fan speed in percent.",
		[]string{"gpu_id", "gpu_uuid", "gpu_name"},
		nil,
	)
	GpuProcessMemoryBytes = prometheus.NewDesc(
		"gpu_process_memory_bytes",
		"VRAM usage per specific PID.",
		[]string{"gpu_id", "gpu_uuid", "gpu_name", "pid", "process_name"},
		nil,
	)
	GpuProcessCPUSecondsTotal = prometheus.NewDesc(
		"gpu_process_cpu_seconds_total",
		"CPU time consumed by a process using the GPU.",
		[]string{"gpu_id", "gpu_uuid", "gpu_name", "pid", "process_name"},
		nil,
	)
	GpuProcessRAMUsageBytes = prometheus.NewDesc(
		"gpu_process_ram_usage_bytes",
		"System RAM used by a GPU-active process.",
		[]string{"gpu_id", "gpu_uuid", "gpu_name", "pid", "process_name"},
		nil,
	)
)
