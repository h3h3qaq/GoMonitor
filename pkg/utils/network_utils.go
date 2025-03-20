package utils

import (
	"net"
	"strings"
)

// GetLocalIPv4 获取本地IPv4地址
func GetLocalIPv4() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "unknown"
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	return "unknown"
}

// GetLocalMACAddress 获取本地MAC地址
func GetLocalMACAddress() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "unknown"
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp != 0 && iface.Flags&net.FlagLoopback == 0 {
			if len(iface.HardwareAddr) > 0 {
				return iface.HardwareAddr.String()
			}
		}
	}

	return "unknown"
}

// GetInterfaceIPAddress 获取指定网络接口的IP地址
func GetInterfaceIPAddress(iface net.Interface) string {
	addrs, err := iface.Addrs()
	if err != nil {
		return ""
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	return ""
}

// ShouldIncludeInterface 判断是否应包含某网络接口
func ShouldIncludeInterface(name string) bool {
	// 过滤掉一些虚拟接口
	excludePrefixes := []string{
		"lo", "veth", "docker", "br-", "vmnet", "vbox", "virbr",
	}

	for _, prefix := range excludePrefixes {
		if strings.HasPrefix(name, prefix) {
			return false
		}
	}

	return true
}
