package main

import (
	"fmt"
	"time"

	"GoMonitor/pkg/models"
	"GoMonitor/proto"
	"github.com/google/uuid"
)

// ClientManager 处理客户端管理相关功能
type ClientManager struct {
	server *Server
}

// NewClientManager 创建客户端管理器
func NewClientManager(server *Server) *ClientManager {
	return &ClientManager{
		server: server,
	}
}

// RegisterClient 注册一个新客户端
func (cm *ClientManager) RegisterClient(req *proto.RegisterRequest) (string, error) {
	clientID := uuid.New().String()

	cm.server.clients[clientID] = &models.ClientInfo{
		ID:         clientID,
		Hostname:   req.Hostname,
		IPAddress:  req.IpAddress,
		MACAddress: req.MacAddress,
		OSInfo:     req.OsInfo,
		LastSeen:   time.Now(),
	}

	cm.server.cmdManager.InitClientCommands(clientID)

	return clientID, nil
}

// UpdateClientInfo 更新客户端的系统信息
func (cm *ClientManager) UpdateClientInfo(clientID string, info *proto.SystemInfo) error {
	client, exists := cm.server.clients[clientID]
	if !exists {
		return fmt.Errorf("未知的客户端ID: %s", clientID)
	}

	client.Info = info
	client.LastSeen = time.Now()

	return nil
}

// ValidateClient 检查客户端是否存在且有效
func (cm *ClientManager) ValidateClient(clientID string) error {
	_, exists := cm.server.clients[clientID]
	if !exists {
		return fmt.Errorf("未知的客户端ID: %s", clientID)
	}
	return nil
}

// GetClientInfo 获取客户端信息
func (cm *ClientManager) GetClientInfo(clientID string) (*models.ClientInfo, error) {
	client, exists := cm.server.clients[clientID]
	if !exists {
		return nil, fmt.Errorf("未知的客户端ID: %s", clientID)
	}
	return client, nil
}

// ListClients 获取所有客户端信息
func (cm *ClientManager) ListClients() []*models.ClientInfo {
	clients := make([]*models.ClientInfo, 0, len(cm.server.clients))
	for _, client := range cm.server.clients {
		clients = append(clients, client)
	}
	return clients
}

// RemoveClient 移除客户端
func (cm *ClientManager) RemoveClient(clientID string) error {
	_, exists := cm.server.clients[clientID]
	if !exists {
		return fmt.Errorf("未知的客户端ID: %s", clientID)
	}

	delete(cm.server.clients, clientID)

	delete(cm.server.clientStreams, clientID)

	cm.server.cmdManager.RemoveClientCommands(clientID)

	return nil
}

// GetClientByHostname 根据主机名查找客户端
func (cm *ClientManager) GetClientByHostname(hostname string) (*models.ClientInfo, error) {
	for _, client := range cm.server.clients {
		if client.Hostname == hostname {
			return client, nil
		}
	}
	return nil, fmt.Errorf("找不到主机名为 %s 的客户端", hostname)
}

// GetClientByIP 根据IP地址查找客户端
func (cm *ClientManager) GetClientByIP(ipAddress string) (*models.ClientInfo, error) {
	for _, client := range cm.server.clients {
		if client.IPAddress == ipAddress {
			return client, nil
		}
	}
	return nil, fmt.Errorf("找不到IP地址为 %s 的客户端", ipAddress)
}
