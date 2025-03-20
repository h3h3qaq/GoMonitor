package cli

import (
	"GoMonitor/pkg/models"
	"GoMonitor/proto"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"

	"GoMonitor/pkg/utils"
	"github.com/manifoldco/promptui"
)

// ServerInterface 定义服务器接口，便于CLI与服务器交互
type ServerInterface interface {
	ListClients() []*models.ClientInfo
	GetClientInfo(clientID string) (*models.ClientInfo, error)
	SendCommandToClient(clientID string, cmdType string, content string, timeout int32) (string, error)
	GetCommandResult(cmdID string) (*proto.CommandResult, error)
}

// ClientInfo 定义CLI需要的客户端信息结构
type ClientInfo struct {
	ID         string
	Hostname   string
	IPAddress  string
	MACAddress string
	OSInfo     string
	LastSeen   time.Time
	Info       *SystemInfo
}

// SystemInfo 定义CLI需要的系统信息结构
type SystemInfo struct {
	CpuInfo     *CPUInfo
	MemoryInfo  *MemoryInfo
	DiskInfo    *DiskInfo
	NetworkInfo *NetworkInfo
}

// CPUInfo 定义CLI需要的CPU信息结构
type CPUInfo struct {
	CpuUsagePercent float64
	CpuCores        int32
}

// MemoryInfo 定义CLI需要的内存信息结构
type MemoryInfo struct {
	TotalMemory        int64
	UsedMemory         int64
	FreeMemory         int64
	MemoryUsagePercent float64
}

// DiskInfo 定义CLI需要的磁盘信息结构
type DiskInfo struct {
	Partitions []*DiskPartition
}

// DiskPartition 定义CLI需要的磁盘分区信息结构
type DiskPartition struct {
	MountPoint   string
	UsagePercent float64
}

// NetworkInfo 定义CLI需要的网络信息结构
type NetworkInfo struct {
	Interfaces map[string]*NetworkInterface
}

// NetworkInterface 定义CLI需要的网络接口信息结构
type NetworkInterface struct {
	Name       string
	IpAddress  string
	MacAddress string
	IsUp       bool
}

// CommandResult 定义CLI需要的命令结果结构
type CommandResult struct {
	ClientId        string
	CommandId       string
	Success         bool
	Output          string
	Error           string
	ExecutionTimeMs int64
	CompletedAt     int64
}

// RunCommandLine 运行交互式命令行界面
func RunCommandLine(s ServerInterface) {
	for {
		prompt := promptui.Select{
			Label: "服务器管理界面",
			Items: []string{
				"列出所有客户端",
				"获取客户端信息",
				"向客户端发送命令",
				"查看命令执行结果",
				"退出",
			},
			HideSelected: false,
			Size:         10,
		}

		idx, _, err := prompt.Run()
		if err != nil {
			fmt.Printf("菜单选择错误: %v\n", err)
			continue
		}

		switch idx {
		case 0:
			handleListClients(s)
		case 1:
			handleGetClientInfo(s)
		case 2:
			handleSendCommand(s)
		case 3:
			handleViewCommandResult(s)
		case 4:
			fmt.Println("退出程序")
			os.Exit(0)
		}
	}
}

// handleListClients 处理列出所有客户端的功能
func handleListClients(s ServerInterface) {
	clients := s.ListClients()
	if len(clients) == 0 {
		fmt.Println("目前没有已连接的客户端")
		return
	}

	fmt.Printf("\n共有 %d 个客户端:\n", len(clients))
	clientsDisplay := make([]string, len(clients))
	for i, client := range clients {
		clientsDisplay[i] = fmt.Sprintf("ID: %s | 主机名: %s | IP: %s",
			client.ID, client.Hostname, client.IPAddress)
	}

	selectPrompt := promptui.Select{
		Label: "选择客户端查看详情 (ESC返回)",
		Items: clientsDisplay,
		Size:  10,
	}

	_, _, err := selectPrompt.Run()
	if err != nil {
		if err == promptui.ErrInterrupt || err == promptui.ErrEOF {
			return
		}
		fmt.Printf("选择错误: %v\n", err)
	}
}

// handleGetClientInfo 处理获取客户端信息的功能
func handleGetClientInfo(s ServerInterface) {
	clients := s.ListClients()
	if len(clients) == 0 {
		fmt.Println("目前没有已连接的客户端")
		return
	}

	clientIDs := make([]string, len(clients))
	for i, client := range clients {
		clientIDs[i] = fmt.Sprintf("%s (%s)", client.ID, client.Hostname)
	}

	selectPrompt := promptui.Select{
		Label: "选择客户端",
		Items: clientIDs,
		Size:  10,
	}

	idx, _, err := selectPrompt.Run()
	if err != nil {
		fmt.Printf("选择错误: %v\n", err)
		return
	}

	clientID := clients[idx].ID
	client, err := s.GetClientInfo(clientID)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	fmt.Printf("\n===== 客户端信息 =====\n")
	fmt.Printf("ID: %s\n", client.ID)
	fmt.Printf("主机名: %s\n", client.Hostname)
	fmt.Printf("IP地址: %s\n", client.IPAddress)
	fmt.Printf("MAC地址: %s\n", client.MACAddress)
	fmt.Printf("操作系统: %s\n", client.OSInfo)
	fmt.Printf("最后活跃时间: %s\n", client.LastSeen.Format(time.RFC3339))

	if client.Info != nil {
		fmt.Printf("\n===== 系统信息 =====\n")
		fmt.Printf("CPU使用率: %.2f%%\n", client.Info.CpuInfo.CpuUsagePercent)
		fmt.Printf("CPU核心数: %d\n", client.Info.CpuInfo.CpuCores)
		fmt.Printf("内存使用率: %.2f%%\n", client.Info.MemoryInfo.MemoryUsagePercent)
		fmt.Printf("总内存: %s\n", utils.FormatBytes(client.Info.MemoryInfo.TotalMemory))
		fmt.Printf("已用内存: %s\n", utils.FormatBytes(client.Info.MemoryInfo.UsedMemory))
		fmt.Printf("磁盘分区数: %d\n", len(client.Info.DiskInfo.Partitions))

		for i, partition := range client.Info.DiskInfo.Partitions {
			fmt.Printf("  分区 %d: %s, 使用率: %.2f%%\n",
				i+1, partition.MountPoint, partition.UsagePercent)
		}
	}

	fmt.Println("\n按Enter键继续...")
	fmt.Scanln()
}

// handleSendCommand 处理向客户端发送命令的功能
func handleSendCommand(s ServerInterface) {
	clients := s.ListClients()
	if len(clients) == 0 {
		fmt.Println("目前没有已连接的客户端")
		return
	}

	clientIDs := make([]string, len(clients))
	for i, client := range clients {
		clientIDs[i] = fmt.Sprintf("%s (%s)", client.ID, client.Hostname)
	}

	selectPrompt := promptui.Select{
		Label: "选择客户端",
		Items: clientIDs,
		Size:  10,
	}

	idx, _, err := selectPrompt.Run()
	if err != nil {
		fmt.Printf("选择错误: %v\n", err)
		return
	}

	clientID := clients[idx].ID

	cmdTypes := []string{"shell", "collect_info", "update"}
	cmdTypePrompt := promptui.Select{
		Label: "选择命令类型",
		Items: cmdTypes,
		Size:  10,
	}

	cmdTypeIdx, _, err := cmdTypePrompt.Run()
	if err != nil {
		fmt.Printf("选择错误: %v\n", err)
		return
	}
	cmdType := cmdTypes[cmdTypeIdx]

	contentPrompt := promptui.Prompt{
		Label:     "输入命令内容",
		Default:   getDefaultContentForCommand(cmdType),
		AllowEdit: true,
	}

	content, err := contentPrompt.Run()
	if err != nil {
		fmt.Printf("输入错误: %v\n", err)
		return
	}

	validate := func(input string) error {
		_, err := strconv.Atoi(input)
		if err != nil {
			return errors.New("请输入有效的数字")
		}
		return nil
	}

	timeoutPrompt := promptui.Prompt{
		Label:    "输入超时时间(秒)",
		Default:  "30",
		Validate: validate,
	}

	timeoutStr, err := timeoutPrompt.Run()
	if err != nil {
		fmt.Printf("输入错误: %v\n", err)
		return
	}

	timeout, _ := strconv.Atoi(timeoutStr)

	cmdID, err := s.SendCommandToClient(clientID, cmdType, content, int32(timeout))
	if err != nil {
		fmt.Printf("发送命令失败: %v\n", err)
	} else {
		fmt.Printf("命令已发送，ID: %s\n", cmdID)
	}

	fmt.Println("\n按Enter键继续...")
	fmt.Scanln()
}

// handleViewCommandResult 处理查看命令执行结果的功能
func handleViewCommandResult(s ServerInterface) {
	cmdIDPrompt := promptui.Prompt{
		Label: "请输入命令ID",
	}

	cmdID, err := cmdIDPrompt.Run()
	if err != nil {
		fmt.Printf("输入错误: %v\n", err)
		return
	}

	result, err := s.GetCommandResult(cmdID)
	if err != nil {
		fmt.Printf("获取结果失败: %v\n", err)
		return
	}

	fmt.Printf("\n===== 命令执行结果 =====\n")
	fmt.Printf("命令ID: %s\n", result.CommandId)
	fmt.Printf("客户端ID: %s\n", result.ClientId)
	fmt.Printf("执行状态: %v\n", result.Success)
	fmt.Printf("执行时间: %d ms\n", result.ExecutionTimeMs)
	fmt.Printf("完成时间: %s\n",
		time.Unix(result.CompletedAt, 0).Format(time.RFC3339))

	if result.Success {
		fmt.Printf("\n===== 输出 =====\n%s\n", result.Output)
	} else {
		fmt.Printf("\n===== 错误 =====\n%s\n", result.Error)
	}

	fmt.Println("\n按Enter键继续...")
	fmt.Scanln()
}

// getDefaultContentForCommand 根据命令类型获取默认内容
func getDefaultContentForCommand(cmdType string) string {
	switch cmdType {
	case "shell":
		if runtime.GOOS == "windows" {
			return "dir"
		}
		return "whoami"
	case "collect_info":
		return "all"
	case "update":
		return "latest"
	default:
		return ""
	}
}
