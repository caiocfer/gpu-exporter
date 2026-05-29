package collector

import (
	"fmt"
	"time"

	"github.com/caiocfer/gpu-exporter/internal/cache"
	"github.com/caiocfer/gpu-exporter/internal/nvml"
	"github.com/prometheus/client_golang/prometheus"
)

func NewRegistry(cacheTTL time.Duration) (*prometheus.Registry, error) {
	devices, err := nvml.GetDevices()
	if err != nil {
		return nil, fmt.Errorf("nvml devices: %w", err)
	}
	if len(devices) == 0 {
		return nil, fmt.Errorf("no NVIDIA GPUs found")
	}
	c := cache.New(cacheTTL)
	reg := prometheus.NewRegistry()
	reg.MustRegister(NewGlobalCollector(devices))
	reg.MustRegister(NewProcessCollector(devices, c))
	reg.MustRegister(prometheus.NewGoCollector())
	reg.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	return reg, nil
}

func formatIndex(i int) string {
	return fmt.Sprintf("%d", i)
}
