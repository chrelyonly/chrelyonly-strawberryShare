package main

import (
	"chrelyonly-localsend-go/model"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

// MulticastService 负责设备的发现逻辑
// 它包含两部分功能：
// 1. Listener: 监听 UDP 多播端口，发现其他设备上线。
// 2. Announcer: 定期或主动发送 UDP 多播，告知其他设备自己在线。
type MulticastService struct {
	alias       string
	fingerprint string
	deviceModel string
	deviceType  model.DeviceType
	port        int // 本机 HTTP 服务端口，告知对方通过此端口连接我
}

// NewMulticastService 创建发现服务实例
func NewMulticastService(alias, fingerprint, deviceModel string, port int) *MulticastService {
	return &MulticastService{
		alias:       alias,
		fingerprint: fingerprint,
		deviceModel: deviceModel,
		deviceType:  model.DeviceTypeDesktop, // 这里硬编码为 Desktop，可根据实际运行环境修改
		port:        port,
	}
}

// StartListener 启动 UDP 多播监听
// 这是一个阻塞方法，建议在 goroutine 中运行
func (s *MulticastService) StartListener() {
	// 解析多播地址
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", DefaultMulticastGroup, s.port))
	if err != nil {
		fmt.Printf("[发现服务] 解析 UDP 地址失败: %v\n", err)
		return
	}

	// 监听多播 UDP
	// 注意：在某些操作系统上，绑定多播端口可能需要特殊权限或配置
	conn, err := net.ListenMulticastUDP("udp", nil, addr)
	if err != nil {
		fmt.Printf("[发现服务] 监听 UDP 多播失败: %v\n", err)
		return
	}
	defer conn.Close()

	// 设置较大的读取缓冲区，避免丢包
	err = conn.SetReadBuffer(UDPSocketBufferSize)
	if err != nil {
		return
	}

	fmt.Printf("[发现服务] 正在监听多播 %s:%d\n", DefaultMulticastGroup, s.port)

	buf := make([]byte, UDPBufferSize) // 最大 UDP 包大小
	for {
		// 读取数据包
		n, src, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Printf("[发现服务] 读取 UDP 数据失败: %v\n", err)
			continue
		}

		// 解析 JSON 数据
		var dto model.MulticastDto
		if err := json.Unmarshal(buf[:n], &dto); err != nil {
			fmt.Printf("[发现服务] 解析多播消息失败: %v\n", err)
			continue
		}

		// 过滤掉自己发送的消息
		// 通过指纹 (Fingerprint) 判断
		if dto.Fingerprint == s.fingerprint {
			continue
		}

		fmt.Printf("[发现服务] 发现设备: %s (%s) 位于 %s:%d\n", dto.Alias, dto.DeviceModel, src.IP, dto.Port)

		// 逻辑扩展点：
		// LocalSend 的标准行为是：
		// 如果收到 Announcement=true (对方刚上线)，且我们也在线 (Server 运行中)，
		// 我们应该回复一个 Announcement，让对方也立即发现我们。
		// 这样可以实现快速的双向发现。
		if dto.Announcement || dto.Announce {
			// 简单策略：收到宣告，我也发一次宣告（注意避免广播风暴，实际可加去重或限流）
			// 这里仅打印日志，实际应用中可以调用 s.SendAnnouncement()
			go s.SendAnnouncement()
		}
	}
}

// SendAnnouncement 发送一次 UDP 广播，宣告自己在线
// 包含自己的 IP、端口、别名等信息
func (s *MulticastService) SendAnnouncement() {
	// 目标地址：多播组 IP + 端口
	// 注意：这里的端口必须与接收端监听的端口一致 (53317)
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", DefaultMulticastGroup, DefaultPort))
	if err != nil {
		fmt.Printf("[发现服务] 解析 UDP 地址失败: %v\n", err)
		return
	}

	// 创建 UDP 连接
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Printf("[发现服务] 连接 UDP 失败: %v\n", err)
		return
	}
	defer conn.Close()

	// 构建数据包
	dto := model.MulticastDto{
		Alias:        s.alias,
		Version:      ProtocolVersion, // "2.1"
		DeviceModel:  s.deviceModel,
		DeviceType:   s.deviceType,
		Fingerprint:  s.fingerprint,
		Port:         s.port, // 告知对方我的 HTTP 服务端口
		Protocol:     model.ProtocolTypeHttp,
		Download:     false,
		Announcement: true, // v1 标志
		Announce:     true, // v2 标志
	}

	data, err := json.Marshal(dto)
	if err != nil {
		fmt.Printf("序列化多播数据失败: %v\n", err)
		return
	}

	// 发送数据
	_, err = conn.Write(data)
	if err != nil {
		fmt.Printf("发送宣告消息失败: %v\n", err)
		return
	}
}

// StartAnnouncer 启动定期广播
// 用于保活或应对网络波动，确保新加入的设备能发现自己
func (s *MulticastService) StartAnnouncer(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// 立即发送一次
	s.SendAnnouncement()

	for range ticker.C {
		s.SendAnnouncement()
	}
}
