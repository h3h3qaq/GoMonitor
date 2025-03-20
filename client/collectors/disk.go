package collectors

import (
	"GoMonitor/proto"
	"github.com/shirou/gopsutil/disk"
)

// CollectDiskInfo 收集磁盘信息
func CollectDiskInfo() (*proto.DiskInfo, error) {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil, err
	}

	diskInfo := &proto.DiskInfo{
		Partitions: []*proto.DiskPartition{},
		DiskReads:  0,
		DiskWrites: 0,
	}

	for _, partition := range partitions {
		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			continue
		}

		diskPartition := &proto.DiskPartition{
			MountPoint:   partition.Mountpoint,
			Filesystem:   partition.Fstype,
			TotalSpace:   int64(usage.Total),
			UsedSpace:    int64(usage.Used),
			FreeSpace:    int64(usage.Free),
			UsagePercent: usage.UsedPercent,
		}

		diskInfo.Partitions = append(diskInfo.Partitions, diskPartition)
	}

	ioStats, err := disk.IOCounters()
	if err == nil {
		for _, stat := range ioStats {
			diskInfo.DiskReads += int64(stat.ReadCount)
			diskInfo.DiskWrites += int64(stat.WriteCount)
		}
	}

	return diskInfo, nil
}
