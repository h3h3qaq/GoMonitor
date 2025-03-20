package main

import (
	"fmt"
	"time"

	"GoMonitor/proto"
	"github.com/google/uuid"
)

// CommandManager 处理命令管理相关功能
type CommandManager struct {
	server      *Server
	pendingCmds map[string]map[string]*proto.Command // client_id -> command_id -> command
	cmdResults  map[string]*proto.CommandResult      // command_id -> result
}

// NewCommandManager 创建命令管理器
func NewCommandManager(server *Server) *CommandManager {
	return &CommandManager{
		server:      server,
		pendingCmds: make(map[string]map[string]*proto.Command),
		cmdResults:  make(map[string]*proto.CommandResult),
	}
}

// InitClientCommands 初始化客户端的命令映射
func (cm *CommandManager) InitClientCommands(clientID string) {
	cm.pendingCmds[clientID] = make(map[string]*proto.Command)
}

// SaveCommandResult 保存命令执行结果
func (cm *CommandManager) SaveCommandResult(clientID string, cmdID string, result *proto.CommandResult) error {
	clientCmds, exists := cm.pendingCmds[clientID]
	if !exists || clientCmds == nil {
		return fmt.Errorf("客户端没有待处理的命令")
	}

	_, cmdExists := clientCmds[cmdID]
	if !cmdExists {
		return fmt.Errorf("未知的命令ID: %s", cmdID)
	}

	cm.cmdResults[cmdID] = result

	delete(clientCmds, cmdID)

	return nil
}

// GetPendingCommands 获取客户端的待处理命令
func (cm *CommandManager) GetPendingCommands(clientID string) []*proto.Command {
	clientCmds, exists := cm.pendingCmds[clientID]
	if !exists || clientCmds == nil {
		return []*proto.Command{}
	}

	cmds := make([]*proto.Command, 0, len(clientCmds))
	for _, cmd := range clientCmds {
		cmds = append(cmds, cmd)
	}

	return cmds
}

// GetCommandResult 获取命令执行结果
func (cm *CommandManager) GetCommandResult(cmdID string) (*proto.CommandResult, error) {
	result, exists := cm.cmdResults[cmdID]
	if !exists {
		return nil, fmt.Errorf("未找到命令结果: %s", cmdID)
	}
	return result, nil
}

// CreateCommand 创建并保存一个新命令
func (cm *CommandManager) CreateCommand(clientID string, cmdType string, content string, timeout int32) (string, error) {
	if err := cm.server.clientManager.ValidateClient(clientID); err != nil {
		return "", err
	}

	cmdID := uuid.New().String()

	cmd := &proto.Command{
		CommandId:      cmdID,
		CommandType:    cmdType,
		Content:        content,
		TimeoutSeconds: timeout,
		IssuedAt:       time.Now().Unix(),
	}

	clientCmds, exists := cm.pendingCmds[clientID]
	if !exists || clientCmds == nil {
		cm.pendingCmds[clientID] = make(map[string]*proto.Command)
		clientCmds = cm.pendingCmds[clientID]
	}
	clientCmds[cmdID] = cmd

	return cmdID, nil
}

// GetCommand 获取特定命令
func (cm *CommandManager) GetCommand(clientID string, cmdID string) (*proto.Command, error) {
	clientCmds, exists := cm.pendingCmds[clientID]
	if !exists || clientCmds == nil {
		return nil, fmt.Errorf("客户端没有待处理的命令")
	}

	cmd, exists := clientCmds[cmdID]
	if !exists {
		return nil, fmt.Errorf("未知的命令ID: %s", cmdID)
	}

	return cmd, nil
}

// RemoveClientCommands 删除客户端的所有命令
func (cm *CommandManager) RemoveClientCommands(clientID string) {
	delete(cm.pendingCmds, clientID)
}

// ListCommandResults 列出所有命令结果
func (cm *CommandManager) ListCommandResults() map[string]*proto.CommandResult {
	return cm.cmdResults
}
