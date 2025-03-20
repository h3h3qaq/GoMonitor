package main

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"GoMonitor/client/collectors"
	"GoMonitor/proto"
)

// CommandExecutor 负责执行从服务器接收的命令
type CommandExecutor struct {
	client *Client
}

// NewCommandExecutor 创建新的命令执行器
func NewCommandExecutor(client *Client) *CommandExecutor {
	return &CommandExecutor{
		client: client,
	}
}

// ExecuteCommand 执行接收到的命令
func (ce *CommandExecutor) ExecuteCommand(cmd *proto.Command) {
	startTime := time.Now()
	result := &proto.CommandResult{
		ClientId:    ce.client.clientID,
		CommandId:   cmd.CommandId,
		Success:     false,
		Output:      "",
		Error:       "",
		CompletedAt: time.Now().Unix(),
	}

	defer func() {
		elapsedTime := time.Since(startTime)
		result.ExecutionTimeMs = elapsedTime.Milliseconds()
		result.CompletedAt = time.Now().Unix()

		ce.client.mu.Lock()
		ce.client.cmdResults[cmd.CommandId] = result
		ce.client.mu.Unlock()

		ce.client.ReportCommandResult(result)
	}()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(cmd.TimeoutSeconds)*time.Second,
	)
	defer cancel()

	commandHandlers := map[string]func(context.Context, *proto.Command, *proto.CommandResult){
		"shell":        ce.ExecuteShellCommand,
		"collect_info": ce.ExecuteCollectInfoCommand,
		"update":       ce.ExecuteUpdateCommand,
	}

	if handler, exists := commandHandlers[cmd.CommandType]; exists {
		handler(ctx, cmd, result)
	} else {
		result.Success = false
		result.Error = fmt.Sprintf("未知的命令类型: %s", cmd.CommandType)
	}
}

// ExecuteShellCommand 执行Shell命令
func (ce *CommandExecutor) ExecuteShellCommand(ctx context.Context, cmd *proto.Command, result *proto.CommandResult) {
	var execCmd *exec.Cmd

	// 根据不同操作系统选择Shell
	if runtime.GOOS == "windows" {
		execCmd = exec.CommandContext(ctx, "cmd", "/C", cmd.Content)
	} else {
		execCmd = exec.CommandContext(ctx, "sh", "-c", cmd.Content)
	}

	output, err := execCmd.CombinedOutput()
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		result.Output = string(output)
		return
	}

	result.Success = true
	result.Output = string(output)
}

// ExecuteCollectInfoCommand 执行信息收集命令
func (ce *CommandExecutor) ExecuteCollectInfoCommand(ctx context.Context, cmd *proto.Command, result *proto.CommandResult) {
	infoType := cmd.Content
	var output string
	var err error

	infoCollectors := map[string]func() (string, error){
		"cpu": func() (string, error) {
			cpuInfo, err := collectors.CollectCPUInfo()
			if err != nil {
				return "", err
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("CPU使用率: %.2f%%\n核心数: %d\n",
				cpuInfo.CpuUsagePercent, cpuInfo.CpuCores))

			for i, coreUsage := range cpuInfo.CoreUsagePercents {
				sb.WriteString(fmt.Sprintf("核心 %d 使用率: %.2f%%\n", i, coreUsage))
			}

			return sb.String(), nil
		},

		"memory": func() (string, error) {
			memInfo, err := collectors.CollectMemoryInfo()
			if err != nil {
				return "", err
			}

			return fmt.Sprintf(
				"内存使用率: %.2f%%\n总内存: %d MB\n已用内存: %d MB\n可用内存: %d MB\n",
				memInfo.MemoryUsagePercent,
				memInfo.TotalMemory/(1024*1024),
				memInfo.UsedMemory/(1024*1024),
				memInfo.FreeMemory/(1024*1024)), nil
		},

		"disk": func() (string, error) {
			diskInfo, err := collectors.CollectDiskInfo()
			if err != nil {
				return "", err
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("磁盘分区数: %d\n", len(diskInfo.Partitions)))

			for i, partition := range diskInfo.Partitions {
				sb.WriteString(fmt.Sprintf(
					"分区 %d: %s, 总空间: %d GB, 已用: %.2f%%\n",
					i+1,
					partition.MountPoint,
					partition.TotalSpace/(1024*1024*1024),
					partition.UsagePercent))
			}

			return sb.String(), nil
		},

		"network": func() (string, error) {
			netInfo, err := collectors.CollectNetworkInfo()
			if err != nil {
				return "", err
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("网络接口数: %d\n", len(netInfo.Interfaces)))

			for name, iface := range netInfo.Interfaces {
				sb.WriteString(fmt.Sprintf(
					"接口: %s, IP: %s, MAC: %s, 状态: %v\n",
					name, iface.IpAddress, iface.MacAddress, iface.IsUp))
			}

			return sb.String(), nil
		},

		"all": func() (string, error) {
			sysInfo, err := ce.client.CollectSystemInfo()
			if err != nil {
				return "", err
			}

			return fmt.Sprintf(
				"CPU使用率: %.2f%%\n内存使用率: %.2f%%\n磁盘分区数: %d\n网络接口数: %d\n",
				sysInfo.CpuInfo.CpuUsagePercent,
				sysInfo.MemoryInfo.MemoryUsagePercent,
				len(sysInfo.DiskInfo.Partitions),
				len(sysInfo.NetworkInfo.Interfaces)), nil
		},

		"process": func() (string, error) {
			// 这里可以添加进程列表收集功能
			return "进程列表收集功能尚未实现", nil
		},
	}

	if collector, exists := infoCollectors[infoType]; exists {
		output, err = collector()
	} else {
		err = fmt.Errorf("未知的信息类型: %s", infoType)
	}

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return
	}

	result.Success = true
	result.Output = output
}

// ExecuteUpdateCommand 执行更新命令（模拟）
func (ce *CommandExecutor) ExecuteUpdateCommand(ctx context.Context, cmd *proto.Command, result *proto.CommandResult) {
	time.Sleep(2 * time.Second)

	result.Success = true
	result.Output = "更新成功模拟\n版本: " + cmd.Content
}
