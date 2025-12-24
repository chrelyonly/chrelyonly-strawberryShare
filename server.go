package main

import (
	"chrelyonly-localsend-go/model"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// FileServer 实现 LocalSend 的 HTTP 协议服务端
// 负责处理设备信息查询、握手、接收文件等请求
type FileServer struct {
	port        int
	alias       string
	fingerprint string
	deviceModel string

	// sessions 存储当前的传输会话状态
	// key: sessionId
	sessions map[string]*Session
}

// Session 代表一次传输会话
type Session struct {
	Id     string
	Files  map[string]model.FileDto // 待接收的文件信息
	Tokens map[string]string        // 每个文件的上传鉴权 Token
}

func NewFileServer(port int, alias, fingerprint, deviceModel string) *FileServer {
	return &FileServer{
		port:        port,
		alias:       alias,
		fingerprint: fingerprint,
		deviceModel: deviceModel,
		sessions:    make(map[string]*Session),
	}
}

// Start 启动 HTTP 服务器
func (s *FileServer) Start() {
	mux := http.NewServeMux()

	// 注册 v2 协议路由
	// 1. 获取设备信息 (用于单播发现)
	mux.HandleFunc("/api/localsend/v2/info", s.handleInfo)
	// 2. 注册/握手 (发现设备后建立连接)
	mux.HandleFunc("/api/localsend/v2/register", s.handleRegister)
	// 3. 准备上传 (发送方请求发送文件)
	mux.HandleFunc("/api/localsend/v2/prepare-upload", s.handlePrepareUpload)
	// 4. 实际上传 (二进制流传输)
	mux.HandleFunc("/api/localsend/v2/upload", s.handleUpload)
	// 5. 取消传输
	mux.HandleFunc("/api/localsend/v2/cancel", s.handleCancel)

	addr := fmt.Sprintf("0.0.0.0:%d", s.port)
	fmt.Printf("[服务端] HTTP 服务器正在监听 %s\n", addr)

	// 启动监听
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Printf("[服务端] 错误: %v\n", err)
	}
}

// handleInfo GET /api/localsend/v2/info
// 返回本机基本信息，用于其他设备通过 IP 直接访问时的探测
func (s *FileServer) handleInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	dto := model.InfoDto{
		Alias:       s.alias,
		Version:     ProtocolVersion,
		DeviceModel: s.deviceModel,
		DeviceType:  model.DeviceTypeDesktop,
		Fingerprint: s.fingerprint,
		Download:    false,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto)
}

// handleRegister POST /api/localsend/v2/register
// 其他设备发现本机后，可能会发送此请求进行握手，或者在准备发送文件前进行握手
func (s *FileServer) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	var req model.RegisterDto
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// 实际业务中，这里可以将对方设备加入到"最近设备"列表或缓存中
	fmt.Printf("[Server] 心跳ping: %s (%s)\n", req.Alias, req.Fingerprint)
	dto := model.InfoDto{
		Alias:       s.alias,
		Version:     ProtocolVersion,
		DeviceModel: s.deviceModel,
		DeviceType:  model.DeviceTypeDesktop,
		Fingerprint: s.fingerprint,
		Download:    false,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto)
}

// handlePrepareUpload POST /api/localsend/v2/prepare-upload
// 接收文件传输请求。发送方会发送包含文件列表的 JSON。
// 接收方（本机）需要在此决定是否接受请求（自动接受或弹窗询问用户）。
func (s *FileServer) handlePrepareUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	var req model.PrepareUploadRequestDto
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Printf("[服务端] 收到来自 %s 的文件传输请求: %d 个文件\n", req.Info.Alias, len(req.Files))
	for _, f := range req.Files {
		fmt.Printf("  - %s (%d 字节)\n", f.FileName, f.Size)
	}

	// --- 关键点：这里通常需要用户交互确认 ---
	// 为了演示，默认自动接受所有请求
	// 实际项目中这里应该阻塞，直到用户点击"接受"或"拒绝"
	// ------------------------------------

	// 生成会话 ID
	sessionId := uuid.New().String()
	session := &Session{
		Id:     sessionId,
		Files:  req.Files,
		Tokens: make(map[string]string),
	}

	// 为每个文件生成传输 Token，用于后续 upload 接口鉴权
	filesResp := make(map[string]string)
	for fileId := range req.Files {
		token := uuid.New().String()
		session.Tokens[fileId] = token
		filesResp[fileId] = token
	}

	// 保存会话状态
	s.sessions[sessionId] = session

	// 返回响应，包含 SessionId 和 Tokens
	resp := model.PrepareUploadResponseDto{
		SessionId: sessionId,
		Files:     filesResp,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleUpload POST /api/localsend/v2/upload
// 实际接收文件数据。请求通过 URL 参数携带 sessionId, fileId, token。
// Body 为文件的原始二进制流。
func (s *FileServer) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 1. 验证参数
	sessionId := r.URL.Query().Get("sessionId")
	fileId := r.URL.Query().Get("fileId")
	token := r.URL.Query().Get("token")

	if sessionId == "" || fileId == "" || token == "" {
		http.Error(w, "Missing parameters", http.StatusBadRequest)
		return
	}

	// 2. 验证会话
	session, ok := s.sessions[sessionId]
	if !ok {
		http.Error(w, "Invalid session", http.StatusForbidden)
		return
	}

	// 3. 验证 Token
	expectedToken, ok := session.Tokens[fileId]
	if !ok || expectedToken != token {
		http.Error(w, "Invalid token", http.StatusForbidden)
		return
	}

	// 4. 获取文件元数据
	fileInfo, ok := session.Files[fileId]
	if !ok {
		http.Error(w, "Invalid fileId", http.StatusBadRequest)
		return
	}

	// 5. 准备保存路径
	// 默认保存到当前目录下的 downloads 文件夹
	downloadDir := DefaultDownloadDir
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		http.Error(w, "Failed to create download dir", http.StatusInternalServerError)
		return
	}

	// 安全处理文件名，防止路径遍历攻击 (../../etc/passwd)
	safeFileName := filepath.Base(fileInfo.FileName)
	// 这里可以添加逻辑：如果文件已存在，自动重命名 (例如 file (1).txt)
	savePath := filepath.Join(downloadDir, safeFileName)

	outFile, err := os.Create(savePath)
	if err != nil {
		http.Error(w, "Failed to create file", http.StatusInternalServerError)
		return
	}
	defer outFile.Close()

	fmt.Printf("[服务端] 正在接收文件: %s ...\n", safeFileName)

	// 6. 接收并写入数据
	// io.Copy 会高效地将 Request Body 流复制到 File
	written, err := io.Copy(outFile, r.Body)
	if err != nil {
		fmt.Printf("[服务端] 写入文件失败: %v\n", err)
		http.Error(w, "写入文件失败", http.StatusInternalServerError)
		return
	}

	if written != fileInfo.Size {
		fmt.Printf("[服务端] 警告: 实际接收大小 %d 与预期 %d 不符\n", written, fileInfo.Size)
	}

	fmt.Printf("[服务端] 文件接收成功: %s (%d 字节)\n", safeFileName, written)
	w.WriteHeader(http.StatusOK)
}

// handleCancel POST /api/localsend/v2/cancel
// 发送方或接收方取消传输
func (s *FileServer) handleCancel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}
	sessionId := r.URL.Query().Get("sessionId")
	if sessionId != "" {
		delete(s.sessions, sessionId)
		fmt.Printf("[服务端] 会话 %s 已取消\n", sessionId)
	}
	w.WriteHeader(http.StatusOK)
}
