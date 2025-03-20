package collectors

import (
	"GoMonitor/proto"
	"github.com/shirou/gopsutil/mem"
)

// CollectMemoryInfo 收集内存信息
func CollectMemoryInfo() (*proto.MemoryInfo, error) {
	memStat, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	swapStat, err := mem.SwapMemory()
	if err != nil {
		swapStat = &mem.SwapMemoryStat{
			Total: 0,
			Used:  0,
			Free:  0,
		}
	}

	return &proto.MemoryInfo{
		TotalMemory:        int64(memStat.Total),
		UsedMemory:         int64(memStat.Used),
		FreeMemory:         int64(memStat.Free),
		MemoryUsagePercent: memStat.UsedPercent,
		SwapTotal:          int64(swapStat.Total),
		SwapUsed:           int64(swapStat.Used),
		SwapFree:           int64(swapStat.Free),
	}, nil
}
