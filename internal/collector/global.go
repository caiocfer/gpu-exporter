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
	ch <- metrics.GpuMemoryTotalBytes
	ch <- metrics.GpuTemperatureCelsius
	ch <- metrics.GpuPowerUsageMilliwatts
	ch <- metrics.GpuFanSpeedPercent
}

func (c *GlobalCollector) Collect(ch chan<- prometheus.Metric) {
	for _, dev := range c.devices {
		labels := []string{formatIndex(dev.Index), dev.UUID, dev.Name}
		gpuUtil, memUtil, err := nvml.GetUtilization(dev)
		if err != nil {
			log.Printf("ERROR: %v", err)
		} else {
			ch <- prometheus.MustNewConstMetric(
				metrics.GpuUtilizationPercent, prometheus.GaugeValue,
				float64(gpuUtil), labels...,
			)
			_ = memUtil
		}
		total, used, err := nvml.GetMemoryInfo(dev)
		if err != nil {
			log.Printf("ERROR: %v", err)
		} else {
			ch <- prometheus.MustNewConstMetric(
				metrics.GpuMemoryUsageBytes, prometheus.GaugeValue,
				float64(used), labels...,
			)
			ch <- prometheus.MustNewConstMetric(
				metrics.GpuMemoryTotalBytes, prometheus.GaugeValue,
				float64(total), labels...,
			)
		}
		temp, err := nvml.GetTemperature(dev)
		if err != nil {
			log.Printf("ERROR: %v", err)
		} else {
			ch <- prometheus.MustNewConstMetric(
				metrics.GpuTemperatureCelsius, prometheus.GaugeValue,
				float64(temp), labels...,
			)
		}
		pw, err := nvml.GetPowerUsage(dev)
		if err != nil {
			log.Printf("ERROR: %v", err)
		} else {
			ch <- prometheus.MustNewConstMetric(
				metrics.GpuPowerUsageMilliwatts, prometheus.GaugeValue,
				float64(pw), labels...,
			)
		}
		fan, err := nvml.GetFanSpeed(dev)
		if err != nil {
			log.Printf("ERROR: %v", err)
		} else {
			ch <- prometheus.MustNewConstMetric(
				metrics.GpuFanSpeedPercent, prometheus.GaugeValue,
				float64(fan), labels...,
			)
		}
	}
}
