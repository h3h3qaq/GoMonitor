package models

import (
	"time"

	"GoMonitor/proto"
)

// ClientInfo 存储客户端的详细信息
type ClientInfo struct {
	ID         string
	Hostname   string
	IPAddress  string
	MACAddress string
	OSInfo     string
	LastSeen   time.Time
	Info       *proto.SystemInfo
}
