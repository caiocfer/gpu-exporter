package collector

import (
	"log"

	"github.com/caiocfer/gpu-exporter/internal/metrics"
	"github.com/caiocfer/gpu-exporter/internal/nvml"
	"github.com/prometheus/client_golang/prometheus"
)

type GlobalCollector struct {
	devices []nvml.Device
}

func NewGlobalCollector(devices []nvml.Device) *GlobalCollector {
	return &GlobalCollector{devices: devices}
}

func (c *GlobalCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- metrics.GpuUtilizationPercent
	ch <- metrics.GpuMemoryUsageBytes
}

func (c *GlobalCollector) Collect(ch chan<- prometheus.Metric) {
	for _, dev := range c.devices {
		gpuUtil, _, err := nvml.GetUtilization(dev)
		if err != nil {
			log.Printf("ERROR: %v", err)
			continue
		}
		_, used, err := nvml.GetMemoryInfo(dev)
		if err != nil {
			log.Printf("ERROR: %v", err)
			continue
		}
		ch <- prometheus.MustNewConstMetric(
			metrics.GpuUtilizationPercent, prometheus.GaugeValue,
			float64(gpuUtil),
			formatIndex(dev.Index), dev.UUID,
		)
		ch <- prometheus.MustNewConstMetric(
			metrics.GpuMemoryUsageBytes, prometheus.GaugeValue,
			float64(used),
			formatIndex(dev.Index), dev.UUID,
		)
	}
}
