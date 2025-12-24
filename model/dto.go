package model

// ProtocolType 定义协议类型
// LocalSend 支持 HTTP 和 HTTPS，v2 协议中默认为 HTTPS，但在局域网内 HTTP 更为常见和高效
type ProtocolType string

const (
	ProtocolTypeHttp  ProtocolType = "http"
	ProtocolTypeHttps ProtocolType = "https"
)

// DeviceType 定义设备类型
// 用于在发现阶段告知对方自己的设备类型，以便显示对应的图标
type DeviceType string

const (
	DeviceTypeMobile   DeviceType = "mobile"
	DeviceTypeDesktop  DeviceType = "desktop"
	DeviceTypeWeb      DeviceType = "web"
	DeviceTypeHeadless DeviceType = "headless"
	DeviceTypeServer   DeviceType = "server"
)

// MulticastDto 对应 common/lib/model/dto/multicast_dto.dart
// 这是 UDP 广播的数据包结构。
// 当设备上线时，会向 224.0.0.167:53317 发送此 JSON 数据。
// 其他设备收到后，解析此数据，得知新设备上线，并发起 HTTP 请求获取详细信息或直接记录。
type MulticastDto struct {
	Alias        string       `json:"alias"`                  // 设备别名，用户可设置
	Version      string       `json:"version,omitempty"`      // 协议版本，例如 "2.1"
	DeviceModel  string       `json:"deviceModel,omitempty"`  // 设备具体型号，如 "iPhone 15", "Windows PC"
	DeviceType   DeviceType   `json:"deviceType,omitempty"`   // 设备类型，用于 UI 显示
	Fingerprint  string       `json:"fingerprint"`            // 设备唯一标识（证书指纹或 UUID），用于去重
	Port         int          `json:"port,omitempty"`         // HTTP 服务监听端口，v2 协议支持自定义端口
	Protocol     ProtocolType `json:"protocol,omitempty"`     // 协议类型 (http/https)
	Download     bool         `json:"download,omitempty"`     // 是否支持下载模式（v2特性）
	Announcement bool         `json:"announcement,omitempty"` // v1 字段：是否为上线宣告
	Announce     bool         `json:"announce,omitempty"`     // v2 字段：是否为上线宣告
}

// InfoDto 对应 common/lib/model/dto/info_dto.dart
// 响应 GET /api/localsend/v2/info
// 用于在单播（直接 IP 访问）时返回设备基本信息
type InfoDto struct {
	Alias       string     `json:"alias"`
	Version     string     `json:"version,omitempty"`
	DeviceModel string     `json:"deviceModel,omitempty"`
	DeviceType  DeviceType `json:"deviceType,omitempty"`
	Fingerprint string     `json:"fingerprint,omitempty"`
	Download    bool       `json:"download,omitempty"`
}

// RegisterDto 对应 common/lib/model/dto/register_dto.dart
// 请求 POST /api/localsend/v2/register
// 当通过 UDP 发现设备后，或者手动输入 IP 后，会发送此请求进行握手
type RegisterDto struct {
	Alias       string       `json:"alias"`
	Version     string       `json:"version,omitempty"`
	DeviceModel string       `json:"deviceModel,omitempty"`
	DeviceType  DeviceType   `json:"deviceType,omitempty"`
	Fingerprint string       `json:"fingerprint"`
	Port        int          `json:"port,omitempty"`
	Protocol    ProtocolType `json:"protocol,omitempty"`
	Download    bool         `json:"download,omitempty"`
}

// FileDto 对应 common/lib/model/dto/file_dto.dart
// 描述单个文件的元数据
type FileDto struct {
	Id       string `json:"id"`                 // 文件唯一 ID，本次传输会话内唯一
	FileName string `json:"fileName"`           // 文件名
	Size     int64  `json:"size"`               // 文件大小（字节）
	FileType string `json:"fileType"`           // MIME 类型
	Hash     string `json:"hash,omitempty"`     // 文件哈希（可选，用于完整性校验）
	Preview  string `json:"preview,omitempty"`  // 文本预览或缩略图（可选）
	Metadata any    `json:"metadata,omitempty"` // 额外元数据（如修改时间）
	Legacy   bool   `json:"legacy,omitempty"`   // 是否为旧版协议
}

// PrepareUploadRequestDto 对应 common/lib/model/dto/prepare_upload_request_dto.dart
// 发送文件前的握手请求：POST /api/localsend/v2/prepare-upload
// 发送方告知接收方即将发送哪些文件
type PrepareUploadRequestDto struct {
	Info  RegisterDto        `json:"info"`  // 发送方设备信息
	Files map[string]FileDto `json:"files"` // 待发送文件列表，key 为 fileId
}

// PrepareUploadResponseDto 对应 common/lib/model/dto/prepare_upload_response_dto.dart
// 接收方同意接收后的响应
// 包含会话 ID 和每个文件的传输 Token
type PrepareUploadResponseDto struct {
	SessionId string            `json:"sessionId"` // 本次传输会话 ID
	Files     map[string]string `json:"files"`     // key: fileId, value: token (用于上传时的鉴权)
}
