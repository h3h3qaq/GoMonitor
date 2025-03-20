package main

import (
	"flag"
	"log"
	"time"
)

var (
	serverAddr = flag.String("server", "localhost:50051", "服务器地址")
	interval   = flag.Int("interval", 60, "收集系统信息的间隔（秒）")
)

func main() {
	flag.Parse()

	log.Printf("客户端启动，服务器地址: %s", *serverAddr)

	c, err := NewClient(*serverAddr)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}
	defer c.Close()

	for {
		err = c.Register()
		if err == nil {
			break
		}

		log.Printf("注册失败: %v，5秒后重试", err)
		time.Sleep(5 * time.Second)
	}

	c.StartReceivingCommands()

	ticker := time.NewTicker(time.Duration(*interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := c.SendSystemInfo()
			if err != nil {
				log.Printf("发送系统信息失败: %v", err)
			} else {
				log.Printf("系统信息发送成功")
			}
		}
	}
}
