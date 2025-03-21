## GoMonitor 项目结构

这是笔者在读书时候为了方便打比赛写的一个小工具，最近使用 LLM 将其翻新了一下

```
整个项目架构如下：
GoMonitor
├── client
│         ├── client.go
│         ├── collectors
│         │         ├── cpu.go
│         │         ├── disk.go
│         │         ├── memory.go
│         │         └── network.go
│         ├── command_executor.go
│         └── main.go
├── go.mod
├── pkg
│         ├── models
│         │         └── client_info.go
│         └── utils
│             ├── network_utils.go
│             └── system_info.go
├── proto
│         ├── system.pb.go
│         ├── system.proto
│         └── system_grpc.pb.go
└── server
    ├── cli
    │         └── command_line.go
    ├── client_manager.go
    ├── command_manager.go
    ├── main.go
    └── server.go
```

## 使用说明：

1.分别将 server 和 client 进行编译

```bash
$ cd server
$ go build -o gomonitor_server
```

```bash
$ cd client
$ go build -o gomonitor_client
```

2.启动服务端

```bash
$ ./gomonitor_server
```

3.启动客户端

```bash
$ ./gomonitor_client --server=localhost:50025 --interval=60
```

