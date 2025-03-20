package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"GoMonitor/client/collectors"
	"GoMonitor/pkg/utils"
	"GoMonitor/proto"
	"google.golang.org/grpc"
)

// Client 表示客户端实例
type Client struct {
	clientID    string
	serverConn  *grpc.ClientConn
	client      proto.SystemInfoServiceClient
	mu          sync.Mutex
	cmdResults  map[string]*proto.CommandResult
	cmdExecutor *CommandExecutor
}

// NewClient 创建新的客户端实例
func NewClient(serverAddr string) (*Client, error) {
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("无法连接到服务器: %v", err)
	}

	client := &Client{
		serverConn: conn,
		client:     proto.NewSystemInfoServiceClient(conn),
		cmdResults: make(map[string]*proto.CommandResult),
	}

	client.cmdExecutor = NewCommandExecutor(client)

	return client, nil
}

// Close 关闭客户端连接
func (c *Client) Close() {
	if c.serverConn != nil {
		c.serverConn.Close()
	}
}

// Register 注册客户端到服务器
func (c *Client) Register() error {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	ipAddr := utils.GetLocalIPv4()
	macAddr := utils.GetLocalMACAddress()
	osInfo := utils.GetOSInfo()

	req := &proto.RegisterRequest{
		Hostname:   hostname,
		IpAddress:  ipAddr,
		MacAddress: macAddr,
		OsInfo:     osInfo,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.client.Register(ctx, req)
	if err != nil {
		return fmt.Errorf("注册失败: %v", err)
	}

	if !resp.Success {
		return fmt.Errorf("注册被拒绝: %s", resp.Message)
	}

	c.clientID = resp.ClientId
	log.Printf("客户端注册成功，ID: %s", c.clientID)
	return nil
}

// CollectSystemInfo 收集当前系统信息
func (c *Client) CollectSystemInfo() (*proto.SystemInfo, error) {
	sysInfo := &proto.SystemInfo{}

	cpuInfo, err := collectors.CollectCPUInfo()
	if err != nil {
		return nil, fmt.Errorf("收集CPU信息失败: %v", err)
	}
	sysInfo.CpuInfo = cpuInfo

	memInfo, err := collectors.CollectMemoryInfo()
	if err != nil {
		return nil, fmt.Errorf("收集内存信息失败: %v", err)
	}
	sysInfo.MemoryInfo = memInfo

	diskInfo, err := collectors.CollectDiskInfo()
	if err != nil {
		return nil, fmt.Errorf("收集磁盘信息失败: %v", err)
	}
	sysInfo.DiskInfo = diskInfo

	netInfo, err := collectors.CollectNetworkInfo()
	if err != nil {
		return nil, fmt.Errorf("收集网络信息失败: %v", err)
	}
	sysInfo.NetworkInfo = netInfo

	// 添加自定义指标
	sysInfo.CustomMetrics = make(map[string]string)
	sysInfo.CustomMetrics["uptime"] = fmt.Sprintf("%d", utils.GetUptime())

	return sysInfo, nil
}

// SendSystemInfo 发送系统信息到服务器
func (c *Client) SendSystemInfo() error {
	if c.clientID == "" {
		return fmt.Errorf("客户端未注册")
	}

	sysInfo, err := c.CollectSystemInfo()
	if err != nil {
		return err
	}

	req := &proto.SystemInfoRequest{
		ClientId:   c.clientID,
		SystemInfo: sysInfo,
		Timestamp:  time.Now().Unix(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.client.SendSystemInfo(ctx, req)
	if err != nil {
		return fmt.Errorf("发送系统信息失败: %v", err)
	}

	if !resp.Received {
		return fmt.Errorf("服务器拒绝接收系统信息: %s", resp.Message)
	}

	return nil
}

// StartReceivingCommands 启动命令接收流
func (c *Client) StartReceivingCommands() {
	go func() {
		for {
			if c.clientID == "" {
				time.Sleep(5 * time.Second)
				continue
			}

			req := &proto.CommandRequest{
				ClientId: c.clientID,
			}

			ctx := context.Background()
			stream, err := c.client.ReceiveCommands(ctx, req)
			if err != nil {
				log.Printf("开始接收命令失败: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}

			log.Printf("已连接到命令流")

			for {
				cmd, err := stream.Recv()
				if err != nil {
					log.Printf("命令流中断: %v", err)
					break
				}

				log.Printf("收到新命令: ID=%s, 类型=%s", cmd.CommandId, cmd.CommandType)
				go c.cmdExecutor.ExecuteCommand(cmd)
			}

			time.Sleep(5 * time.Second)
		}
	}()
}

// ReportCommandResult 向服务器报告命令执行结果
func (c *Client) ReportCommandResult(result *proto.CommandResult) {
	for i := 0; i < 3; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		resp, err := c.client.ReportCommandResult(ctx, result)
		if err == nil && resp.Received {
			log.Printf("命令结果报告成功: %s", result.CommandId)
			return
		}

		log.Printf("命令结果报告失败 (尝试 %d/3): %v", i+1, err)
		time.Sleep(2 * time.Second)
	}

	log.Printf("命令结果报告最终失败: %s", result.CommandId)
}
