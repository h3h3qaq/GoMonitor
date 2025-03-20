package utils

import (
	"fmt"
	"runtime"

	"github.com/shirou/gopsutil/host"
)

// GetOSInfo 获取操作系统信息
func GetOSInfo() string {
	hostInfo, err := host.Info()
	if err != nil {
		return fmt.Sprintf("unknown (%s)", runtime.GOOS)
	}

	return fmt.Sprintf("%s %s (%s)", hostInfo.Platform, hostInfo.PlatformVersion, runtime.GOOS)
}

// GetUptime 获取系统运行时间（秒）
func GetUptime() int64 {
	hostInfo, err := host.Info()
	if err != nil {
		return 0
	}
	return int64(hostInfo.Uptime)
}

// FormatBytes 格式化字节数为更易读的形式
func FormatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
		TB = 1024 * GB
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/float64(TB))
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
