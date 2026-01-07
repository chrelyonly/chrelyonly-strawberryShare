# strawberryShare

基于 LocalSend 协议的局域网文件传输工具，使用 Go 语言实现。

## 功能特性

- 🔍 **自动设备发现**：通过 UDP 多播自动发现局域网内的其他设备
- 📁 **文件传输**：支持设备间快速文件传输
- 🔒 **HTTPS 支持**：默认启用 HTTPS 加密传输
- 🎯 **双模式运行**：支持接收端（Server）和发送端（Sender）两种模式
- 📱 **设备识别**：支持设备别名、型号和指纹识别
- 🛡️ **会话管理**：基于 Token 的安全传输机制

## 快速开始

### 环境要求

- Go 1.25 或更高版本

### 安装

```bash
# 克隆仓库
git clone https://github.com/yourusername/chrelyonly-strawberryShare.git
cd chrelyonly-strawberryShare

# 下载依赖
go mod download
```

### 运行

#### 接收端模式（Server）

启动接收端，监听文件上传请求：

```bash
go run . -mode server -port 53317 -alias "我的设备"
```

参数说明：
- `-mode server`：运行模式为接收端
- `-port 53317`：监听端口（默认：53317）
- `-alias "我的设备"`：设备别名（默认：局域网共享传输）

#### 发送端模式（Sender）

向指定设备发送文件：

```bash
go run . -mode sender -target 192.168.1.100 -file /path/to/file.txt
```

参数说明：
- `-mode sender`：运行模式为发送端
- `-target 192.168.1.100`：目标设备 IP 地址（必填）
- `-file /path/to/file.txt`：待发送文件路径（必填）

## 项目结构

```
.
├── main.go           # 程序入口
├── server.go         # HTTP 服务器实现
├── discovery.go      # UDP 多播设备发现
├── sender.go         # 文件发送器
├── constants.go      # 常量定义
├── model/
│   └── dto.go        # 数据传输对象
├── downloads/        # 默认文件下载目录
├── server.pem        # HTTPS 证书
└── server.key        # HTTPS 私钥
```

## 协议说明

本实现遵循 LocalSend v2.1 协议规范：

### UDP 多播发现

- **多播地址**：224.0.0.167
- **端口**：53317
- **协议**：JSON 格式的 UDP 数据包

### HTTP/HTTPS 接口

#### 1. 获取设备信息
```
GET /api/localsend/v2/info
```

#### 2. 设备注册/握手
```
POST /api/localsend/v2/register
```

#### 3. 准备上传
```
POST /api/localsend/v2/prepare-upload
```

#### 4. 文件上传
```
POST /api/localsend/v2/upload?sessionId={id}&fileId={id}&token={token}
```

#### 5. 取消传输
```
POST /api/localsend/v2/cancel
```

## 配置

### HTTPS 配置

默认启用 HTTPS，证书文件位于项目根目录：
- `server.pem`：证书文件
- `server.key`：私钥文件

如需切换到 HTTP 模式，修改 `constants.go` 中的配置：

```go
const IsHttps = false
```

### 默认配置

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| 端口 | 53317 | 监听端口 |
| 多播组 | 224.0.0.167 | UDP 多播地址 |
| 下载目录 | ./downloads | 文件保存路径 |
| 设备别名 | 局域网共享传输 | 设备显示名称 |

## 使用示例

### 场景 1：两台设备间传输文件

**设备 A（接收端）：**
```bash
go run . -mode server -alias "设备A"
```

**设备 B（发送端）：**
```bash
go run . -mode sender -target 192.168.1.100 -file document.pdf
```

### 场景 2：自定义端口

```bash
# 接收端使用自定义端口
go run . -mode server -port 8080 -alias "我的电脑"

# 发送端指定目标端口
go run . -mode sender -target 192.168.1.100:8080 -file photo.jpg
```

## 安全说明

- 所有传输默认使用 HTTPS 加密
- 每次传输使用唯一的会话 ID 和 Token
- 文件名经过安全处理，防止路径遍历攻击
- 设备使用 UUID 指纹进行身份验证

## 开发

### 生成 HTTPS 证书

```bash
openssl req -x509 -newkey rsa:4096 -keyout server.key -out server.pem -days 365 -nodes
```

### 依赖管理

```bash
# 添加依赖
go get github.com/package/name

# 更新依赖
go mod tidy
```

## 注意事项

- 确保设备在同一局域网内
- 防火墙需要允许 UDP 53317 端口和 TCP 53317 端口
- 发送端需要知道接收端的 IP 地址
- 接收的文件默认保存在 `downloads` 目录

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！

## 致谢

本项目基于 [LocalSend](https://github.com/localsend/localsend) 协议实现。
