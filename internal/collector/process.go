package collector

import (
	"log"
	"strconv"

	"github.com/caiocfer/gpu-exporter/internal/cache"
	"github.com/caiocfer/gpu-exporter/internal/metrics"
	"github.com/caiocfer/gpu-exporter/internal/nvml"
	"github.com/caiocfer/gpu-exporter/internal/proc"
	"github.com/prometheus/client_golang/prometheus"
)

type ProcessCollector struct {
	devices []nvml.Device
	cache   *cache.Cache
}

func NewProcessCollector(devices []nvml.Device, c *cache.Cache) *ProcessCollector {
	return &ProcessCollector{devices: devices, cache: c}
}

func (c *ProcessCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- metrics.GpuProcessMemoryBytes
	ch <- metrics.GpuProcessCPUSecondsTotal
	ch <- metrics.GpuProcessRAMUsageBytes
}

func (c *ProcessCollector) Collect(ch chan<- prometheus.Metric) {
	for _, dev := range c.devices {
		procs, err := nvml.GetMergedProcesses(dev)
		if err != nil {
			log.Printf("ERROR: %v", err)
			continue
		}
		for _, p := range procs {
			pidStr := strconv.FormatUint(uint64(p.PID), 10)
			entry, ok := c.cache.Get(p.PID)
			var name string
			if ok {
				name = entry.Name
			} else {
				name = proc.ReadComm(p.PID)
				cmdline := proc.ReadCmdline(p.PID)
				if name == "" {
					name = cmdline
				}
				c.cache.Set(p.PID, name, cmdline)
			}
			labels := []string{formatIndex(dev.Index), dev.UUID, dev.Name, pidStr, name}
			ch <- prometheus.MustNewConstMetric(
				metrics.GpuProcessMemoryBytes, prometheus.GaugeValue,
				float64(p.MemMB), labels...,
			)
			cpu, err := proc.ReadStat(p.PID)
			if err == nil {
				ch <- prometheus.MustNewConstMetric(
					metrics.GpuProcessCPUSecondsTotal, prometheus.CounterValue,
					float64(cpu)/100.0, labels...,
				)
			}
			rss := proc.ReadRSS(p.PID)
			ch <- prometheus.MustNewConstMetric(
				metrics.GpuProcessRAMUsageBytes, prometheus.GaugeValue,
				float64(rss), labels...,
			)
		}
	}
}
