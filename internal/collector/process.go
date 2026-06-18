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
	devices       []nvml.Device
	cache         *cache.Cache
	lastTimestamps map[int]uint64
}

func NewProcessCollector(devices []nvml.Device, c *cache.Cache) *ProcessCollector {
	return &ProcessCollector{devices: devices, cache: c, lastTimestamps: make(map[int]uint64)}
}

func (c *ProcessCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- metrics.GpuProcessMemoryBytes
	ch <- metrics.GpuProcessCPUSecondsTotal
	ch <- metrics.GpuProcessRAMUsageBytes
	ch <- metrics.GpuProcessSmUtilizationPercent
}

func (c *ProcessCollector) Collect(ch chan<- prometheus.Metric) {
	for _, dev := range c.devices {
		procs, err := nvml.GetMergedProcesses(dev)
		if err != nil {
			log.Printf("ERROR: %v", err)
			continue
		}

		smByPID, newTS, err := nvml.GetProcessUtilization(dev, c.lastTimestamps[dev.Index])
		if err != nil {
			log.Printf("ERROR: GetProcessUtilization(%d): %v", dev.Index, err)
		} else {
			if newTS > c.lastTimestamps[dev.Index] {
				c.lastTimestamps[dev.Index] = newTS
			}
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
			if smByPID != nil {
				if smUtil, ok := smByPID[p.PID]; ok {
					ch <- prometheus.MustNewConstMetric(
						metrics.GpuProcessSmUtilizationPercent, prometheus.GaugeValue,
						float64(smUtil), labels...,
					)
				}
			}
		}
	}
}
