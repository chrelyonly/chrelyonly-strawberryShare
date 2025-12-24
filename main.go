package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

// main 是程序的入口点
// 支持两种模式：
// 1. server (默认): 启动接收端，监听 UDP 广播和 HTTP 文件上传请求
// 2. sender: 启动发送端，向指定 IP 发送文件
func main() {
	// --- 1. 解析命令行参数 ---
	port := flag.Int("port", DefaultPort, "监听端口 (默认: 53317)")
	alias := flag.String("alias", "Go-LocalSend", "设备别名")
	mode := flag.String("mode", "server", "运行模式: server (接收) 或 sender (发送)")
	targetIP := flag.String("target", "", "目标 IP 地址 (发送模式必填)")
	fileToSend := flag.String("file", "", "待发送文件路径 (发送模式必填)")
	flag.Parse()

	// --- 2. 初始化设备标识 ---
	// 实际应用中，指纹应持久化存储，以保持设备身份一致性
	fingerprint := uuid.New().String()
	deviceModel := "Go-Client" // 设备型号

	fmt.Println("------------------------------------------------")
	fmt.Printf("Go-LocalSend 协议版本 v%s\n", ProtocolVersion)
	fmt.Printf("别名:        %s\n", *alias)
	fmt.Printf("指纹:        %s\n", fingerprint)
	fmt.Printf("端口:        %d\n", *port)
	fmt.Printf("模式:        %s\n", *mode)
	fmt.Println("------------------------------------------------")

	// --- 3. 初始化 UDP 发现服务 ---
	// 无论发送端还是接收端，都需要监听多播，以便发现其他设备
	discovery := NewMulticastService(*alias, fingerprint, deviceModel, *port)

	// 异步启动 UDP 监听器
	go discovery.StartListener()

	// --- 4. 根据模式执行逻辑 ---
	if *mode == "server" {
		// === 接收端逻辑 ===

		// 启动定期广播宣告 (Announcer)
		// 这样其他设备打开 App 时能立即发现我
		go discovery.StartAnnouncer(2 * time.Second)

		// 启动 HTTP 服务器
		// 阻塞运行，处理所有入站请求 (Info, Register, Upload)
		server := NewFileServer(*port, *alias, fingerprint, deviceModel)
		server.Start()

	} else if *mode == "sender" {
		// === 发送端逻辑 ===

		if *targetIP == "" || *fileToSend == "" {
			log.Fatal("错误: 发送模式需要指定 -target 和 -file 参数")
		}

		// 发送一次广播宣告（可选）
		// 让局域网内其他设备知道我上线了
		discovery.SendAnnouncement()

		// 初始化发送器
		sender := NewSender(*alias, fingerprint, deviceModel, *port)

		// 执行发送流程
		// 注意：这里假设对方监听默认端口 (53317)。
		// 完善的实现应该是先通过 Discovery 发现对方的 Port，再建立连接。
		err := sender.SendFile(*targetIP, DefaultPort, *fileToSend)
		if err != nil {
			log.Fatalf("发送失败: %v", err)
		}

		fmt.Println("完成。")
	} else {
		log.Fatal("无效模式。请使用 'server' 或 'sender'")
	}
}
