package collectors

import (
	"runtime"
	"time"

	"GoMonitor/proto"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
)

func CollectCPUInfo() (*proto.CPUInfo, error) {
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return nil, err
	}

	corePercents, err := cpu.Percent(time.Second, true)
	if err != nil {
		corePercents = []float64{}
	}

	counts, err := cpu.Counts(true)
	if err != nil {
		counts = runtime.NumCPU()
	}

	var loadAvg1, loadAvg5, loadAvg15 int64

	loadAvgStat, err := load.Avg()
	if err == nil {
		loadAvg1 = int64(loadAvgStat.Load1 * 100)
		loadAvg5 = int64(loadAvgStat.Load5 * 100)
		loadAvg15 = int64(loadAvgStat.Load15 * 100)
	}

	return &proto.CPUInfo{
		CpuUsagePercent:   cpuPercent[0],
		CpuCores:          int32(counts),
		CoreUsagePercents: corePercents,
		LoadAverage_1M:    loadAvg1,
		LoadAverage_5M:    loadAvg5,
		LoadAverage_15M:   loadAvg15,
	}, nil
}
