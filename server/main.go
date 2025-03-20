package main

import (
	"fmt"
	"log"
	"net"

	"GoMonitor/proto"
	"GoMonitor/server/cli"
	"google.golang.org/grpc"
)

func main() {
	port := 50025
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("监听端口失败: %v", err)
	}

	s := grpc.NewServer()
	serverImpl := NewServer()
	proto.RegisterSystemInfoServiceServer(s, serverImpl)

	log.Printf("服务器启动，监听端口: %d", port)

	go cli.RunCommandLine(serverImpl)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
