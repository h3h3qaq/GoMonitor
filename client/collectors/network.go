package collectors

import (
	"strings"

	"GoMonitor/pkg/utils"
	"GoMonitor/proto"
	psnet "github.com/shirou/gopsutil/net"
)

// CollectNetworkInfo 收集网络信息
func CollectNetworkInfo() (*proto.NetworkInfo, error) {
	interfaces, err := psnet.Interfaces()
	if err != nil {
		return nil, err
	}

	ioStats, err := psnet.IOCounters(true)
	if err != nil {
		ioStats = []psnet.IOCountersStat{}
	}

	netInfo := &proto.NetworkInfo{
		Interfaces:      make(map[string]*proto.NetworkInterface),
		BytesSent:       0,
		BytesReceived:   0,
		PacketsSent:     0,
		PacketsReceived: 0,
	}

	for _, io := range ioStats {
		netInfo.BytesSent += int64(io.BytesSent)
		netInfo.BytesReceived += int64(io.BytesRecv)
		netInfo.PacketsSent += int64(io.PacketsSent)
		netInfo.PacketsReceived += int64(io.PacketsRecv)
	}

	for _, iface := range interfaces {
		if !utils.ShouldIncludeInterface(iface.Name) {
			continue
		}

		isUp := false
		for _, flag := range iface.Flags {
			if strings.ToLower(flag) == "up" {
				isUp = true
				break
			}
		}

		netIface := &proto.NetworkInterface{
			Name:          iface.Name,
			IpAddress:     getInterfaceIPAddress(iface),
			MacAddress:    iface.HardwareAddr,
			BytesSent:     0,
			BytesReceived: 0,
			IsUp:          isUp,
		}

		for _, io := range ioStats {
			if io.Name == iface.Name {
				netIface.BytesSent = int64(io.BytesSent)
				netIface.BytesReceived = int64(io.BytesRecv)
				break
			}
		}

		netInfo.Interfaces[iface.Name] = netIface
	}

	return netInfo, nil
}

func getInterfaceIPAddress(iface psnet.InterfaceStat) string {
	for _, addr := range iface.Addrs {
		// 从CIDR格式中提取IP地址
		ip := strings.Split(addr.Addr, "/")[0]
		// 简单检查是否为IPv4地址
		if strings.Count(ip, ".") == 3 {
			return ip
		}
	}
	return ""
}
