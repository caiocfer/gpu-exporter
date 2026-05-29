package nvml

import (
	"fmt"
	"os"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

type Device struct {
	Handle nvml.Device
	Index  int
	UUID   string
}

type ProcessInfo struct {
	PID    uint32
	MemMB  uint64
	Device Device
}

func Init() error {
	if p := os.Getenv("NVML_LIB_PATH"); p != "" {
		if err := nvml.SetLibraryOptions(nvml.WithLibraryPath(p)); err != nil {
			return fmt.Errorf("nvml.SetLibraryOptions: %w", err)
		}
	}
	ret := nvml.Init()
	if ret != nvml.SUCCESS {
		return fmt.Errorf("nvml.Init: %s", nvml.ErrorString(ret))
	}
	return nil
}

func Shutdown() {
	nvml.Shutdown()
}

func GetDevices() ([]Device, error) {
	count, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		return nil, fmt.Errorf("DeviceGetCount: %s", nvml.ErrorString(ret))
	}
	devices := make([]Device, 0, count)
	for i := 0; i < count; i++ {
		handle, ret := nvml.DeviceGetHandleByIndex(i)
		if ret != nvml.SUCCESS {
			return nil, fmt.Errorf("DeviceGetHandleByIndex(%d): %s", i, nvml.ErrorString(ret))
		}
		uuid, ret := handle.GetUUID()
		if ret != nvml.SUCCESS {
			return nil, fmt.Errorf("GetUUID(%d): %s", i, nvml.ErrorString(ret))
		}
		devices = append(devices, Device{Handle: handle, Index: i, UUID: uuid})
	}
	return devices, nil
}

func getProcessList(dev Device, fn func() ([]nvml.ProcessInfo, nvml.Return)) ([]ProcessInfo, error) {
	procs, ret := fn()
	if ret != nvml.SUCCESS {
		return nil, fmt.Errorf("nvml error device %d: %s", dev.Index, nvml.ErrorString(ret))
	}
	infos := make([]ProcessInfo, 0, len(procs))
	for _, p := range procs {
		infos = append(infos, ProcessInfo{
			PID:    p.Pid,
			MemMB:  p.UsedGpuMemory,
			Device: dev,
		})
	}
	return infos, nil
}

func GetComputeProcesses(dev Device) ([]ProcessInfo, error) {
	return getProcessList(dev, dev.Handle.GetComputeRunningProcesses)
}

func GetGraphicsProcesses(dev Device) ([]ProcessInfo, error) {
	return getProcessList(dev, dev.Handle.GetGraphicsRunningProcesses)
}

func GetMergedProcesses(dev Device) ([]ProcessInfo, error) {
	compute, errC := GetComputeProcesses(dev)
	graphics, errG := GetGraphicsProcesses(dev)
	if errC != nil && errG != nil {
		return nil, fmt.Errorf("compute: %w; graphics: %w", errC, errG)
	}
	byPID := make(map[uint32]ProcessInfo)
	for _, p := range compute {
		byPID[p.PID] = p
	}
	for _, p := range graphics {
		if existing, ok := byPID[p.PID]; ok {
			if p.MemMB > existing.MemMB {
				byPID[p.PID] = p
			}
		} else {
			byPID[p.PID] = p
		}
	}
	merged := make([]ProcessInfo, 0, len(byPID))
	for _, p := range byPID {
		merged = append(merged, p)
	}
	return merged, nil
}

func GetUtilization(dev Device) (uint32, uint32, error) {
	util, ret := dev.Handle.GetUtilizationRates()
	if ret != nvml.SUCCESS {
		return 0, 0, fmt.Errorf("GetUtilizationRates(%d): %s", dev.Index, nvml.ErrorString(ret))
	}
	return util.Gpu, util.Memory, nil
}

func GetTemperature(dev Device) (uint32, error) {
	temp, ret := dev.Handle.GetTemperature(nvml.TEMPERATURE_GPU)
	if ret != nvml.SUCCESS {
		return 0, fmt.Errorf("GetTemperature(%d): %s", dev.Index, nvml.ErrorString(ret))
	}
	return temp, nil
}

func GetPowerUsage(dev Device) (uint32, error) {
	mw, ret := dev.Handle.GetPowerUsage()
	if ret != nvml.SUCCESS {
		return 0, fmt.Errorf("GetPowerUsage(%d): %s", dev.Index, nvml.ErrorString(ret))
	}
	return mw, nil
}

func GetFanSpeed(dev Device) (uint32, error) {
	speed, ret := dev.Handle.GetFanSpeed()
	if ret != nvml.SUCCESS {
		return 0, fmt.Errorf("GetFanSpeed(%d): %s", dev.Index, nvml.ErrorString(ret))
	}
	return speed, nil
}

func GetMemoryInfo(dev Device) (uint64, uint64, error) {
	info, ret := dev.Handle.GetMemoryInfo()
	if ret != nvml.SUCCESS {
		return 0, 0, fmt.Errorf("GetMemoryInfo(%d): %s", dev.Index, nvml.ErrorString(ret))
	}
	return info.Total, info.Used, nil
}
