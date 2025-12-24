package main

import (
	"runtime"
	"time"
)

const (
	// DefaultPort LocalSend 默认端口
	DefaultPort = 53317

	// ProtocolVersion 当前实现的协议版本
	ProtocolVersion = "2.1"

	// DefaultMulticastGroup LocalSend 默认多播组地址
	DefaultMulticastGroup = "224.0.0.167"

	// DefaultAlias 默认设备别名
	DefaultAlias = "莓莓分享小助手"

	// DefaultDeviceModel 默认设备型号
	DefaultDeviceModel = runtime.GOOS

	// UDPBufferSize UDP 读取缓冲区大小
	UDPBufferSize = 65535 // Max UDP packet size

	// UDPSocketBufferSize UDP Socket 缓冲区大小
	UDPSocketBufferSize = 1024 * 1024

	// ConnectTimeout 连接超时时间
	ConnectTimeout = 60 * time.Second

	// DefaultDownloadDir 默认下载目录
	DefaultDownloadDir = "downloads"
)
