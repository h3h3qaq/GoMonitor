package main

import (
	"GoMonitor/pkg/models"
	"GoMonitor/proto"
	"context"
	"log"
	"sync"
)

// Server 是gRPC服务的主要实现
type Server struct {
	proto.UnimplementedSystemInfoServiceServer
	mu            sync.Mutex
	clients       map[string]*models.ClientInfo
	clientStreams map[string]proto.SystemInfoService_ReceiveCommandsServer
	clientManager *ClientManager
	cmdManager    *CommandManager
}

// NewServer 创建一个新的服务器实例
func NewServer() *Server {
	server := &Server{
		clients:       make(map[string]*models.ClientInfo),
		clientStreams: make(map[string]proto.SystemInfoService_ReceiveCommandsServer),
	}

	server.clientManager = NewClientManager(server)
	server.cmdManager = NewCommandManager(server)

	return server
}

// Register 处理客户端注册请求
func (s *Server) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	clientID, err := s.clientManager.RegisterClient(req)
	if err != nil {
		return &proto.RegisterResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	log.Printf("客户端注册: ID=%s, 主机名=%s, IP=%s", clientID, req.Hostname, req.IpAddress)

	return &proto.RegisterResponse{
		ClientId: clientID,
		Success:  true,
		Message:  "注册成功",
	}, nil
}

// SendSystemInfo 处理客户端发送的系统信息
func (s *Server) SendSystemInfo(ctx context.Context, req *proto.SystemInfoRequest) (*proto.SystemInfoResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	clientID := req.ClientId
	err := s.clientManager.UpdateClientInfo(clientID, req.SystemInfo)
	if err != nil {
		return &proto.SystemInfoResponse{
			Received: false,
			Message:  err.Error(),
		}, err
	}

	cpuUsage := req.SystemInfo.CpuInfo.CpuUsagePercent
	memUsage := req.SystemInfo.MemoryInfo.MemoryUsagePercent
	log.Printf("收到客户端 %s 的系统信息: CPU使用率=%.2f%%, 内存使用率=%.2f%%",
		clientID, cpuUsage, memUsage)

	return &proto.SystemInfoResponse{
		Received: true,
		Message:  "信息已接收",
	}, nil
}

// ReceiveCommands 建立命令流
func (s *Server) ReceiveCommands(req *proto.CommandRequest, stream proto.SystemInfoService_ReceiveCommandsServer) error {
	clientID := req.ClientId

	if err := s.clientManager.ValidateClient(clientID); err != nil {
		return err
	}

	s.mu.Lock()
	s.clientStreams[clientID] = stream

	pendingCommands := s.cmdManager.GetPendingCommands(clientID)
	s.mu.Unlock()

	log.Printf("客户端 %s 已连接接收命令流", clientID)

	for _, cmd := range pendingCommands {
		if err := stream.Send(cmd); err != nil {
			log.Printf("向客户端 %s 发送命令失败: %v", clientID, err)
		}
	}

	<-stream.Context().Done()

	s.mu.Lock()
	delete(s.clientStreams, clientID)
	s.mu.Unlock()

	log.Printf("客户端 %s 的命令流已断开", clientID)
	return nil
}

// ReportCommandResult 处理命令执行结果
func (s *Server) ReportCommandResult(ctx context.Context, result *proto.CommandResult) (*proto.CommandResultResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	clientID := result.ClientId
	cmdID := result.CommandId

	if err := s.clientManager.ValidateClient(clientID); err != nil {
		return &proto.CommandResultResponse{
			Received: false,
			Message:  err.Error(),
		}, err
	}

	if err := s.cmdManager.SaveCommandResult(clientID, cmdID, result); err != nil {
		return &proto.CommandResultResponse{
			Received: false,
			Message:  err.Error(),
		}, err
	}

	log.Printf("收到客户端 %s 的命令 %s 执行结果: 成功=%v",
		clientID, cmdID, result.Success)

	if !result.Success {
		log.Printf("命令执行失败，错误: %s", result.Error)
	}

	return &proto.CommandResultResponse{
		Received: true,
		Message:  "结果已接收",
	}, nil
}

// 获取客户端列表
func (s *Server) ListClients() []*models.ClientInfo {
	return s.clientManager.ListClients()
}

// 获取客户端信息
func (s *Server) GetClientInfo(clientID string) (*models.ClientInfo, error) {
	return s.clientManager.GetClientInfo(clientID)
}

// 向客户端发送命令
func (s *Server) SendCommandToClient(clientID string, cmdType string, content string, timeout int32) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cmdID, err := s.cmdManager.CreateCommand(clientID, cmdType, content, timeout)
	if err != nil {
		return "", err
	}

	if stream, ok := s.clientStreams[clientID]; ok {
		cmd, err := s.cmdManager.GetCommand(clientID, cmdID)
		if err != nil {
			return cmdID, nil // 返回命令ID，但不报告错误
		}

		if err := stream.Send(cmd); err != nil {
			log.Printf("向客户端 %s 发送命令失败: %v", clientID, err)
			// 命令已创建，稍后会重试发送，所以不返回错误
		}
	}

	log.Printf("向客户端 %s 创建命令: ID=%s, 类型=%s", clientID, cmdID, cmdType)
	return cmdID, nil
}

// 获取命令执行结果
func (s *Server) GetCommandResult(cmdID string) (*proto.CommandResult, error) {
	return s.cmdManager.GetCommandResult(cmdID)
}
