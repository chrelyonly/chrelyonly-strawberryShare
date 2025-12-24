package main

import (
	"bytes"
	"chrelyonly-localsend-go/model"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// Sender 负责发送文件
type Sender struct {
	client *http.Client
	info   model.RegisterDto // 自己的信息
}

func NewSender(alias, fingerprint, deviceModel string, port int) *Sender {
	return &Sender{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		info: model.RegisterDto{
			Alias:       alias,
			Version:     ProtocolVersion,
			DeviceModel: deviceModel,
			DeviceType:  model.DeviceTypeDesktop,
			Fingerprint: fingerprint,
			Port:        port,
			Protocol:    model.ProtocolTypeHttp,
			Download:    false,
		},
	}
}

// SendFile 发送文件给目标设备
func (s *Sender) SendFile(targetIP string, targetPort int, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	fileStat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %v", err)
	}

	fileId := uuid.New().String()
	fileName := filepath.Base(filePath)
	fileSize := fileStat.Size()

	// 1. Prepare Upload
	files := make(map[string]model.FileDto)
	files[fileId] = model.FileDto{
		Id:       fileId,
		FileName: fileName,
		Size:     fileSize,
		FileType: "application/octet-stream", // 简化
	}

	reqDto := model.PrepareUploadRequestDto{
		Info:  s.info,
		Files: files,
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	reqBody, _ := json.Marshal(reqDto)
	targetUrl := fmt.Sprintf("https://%s:%d/api/localsend/v2/prepare-upload", targetIP, targetPort)

	fmt.Printf("[发送端] 正在发送准备上传请求至 %s\n", targetUrl)
	resp, err := client.Post(targetUrl, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("准备上传失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// 读取错误信息
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("准备上传请求被拒绝 (状态码 %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var prepareResp model.PrepareUploadResponseDto
	if err := json.NewDecoder(resp.Body).Decode(&prepareResp); err != nil {
		return fmt.Errorf("解析响应失败: %v", err)
	}

	token, ok := prepareResp.Files[fileId]
	if !ok {
		return fmt.Errorf("服务器未返回文件 Token")
	}

	// 2. Upload File
	uploadUrl := fmt.Sprintf("https://%s:%d/api/localsend/v2/upload?sessionId=%s&fileId=%s&token=%s",
		targetIP, targetPort, prepareResp.SessionId, fileId, token)

	fmt.Printf("[发送端] 正在上传文件至 %s\n", uploadUrl)

	// 由于是二进制流上传，直接把 file 作为 Body
	// 注意：LocalSend v2 upload 接口直接接收 binary stream，不需要 multipart
	uploadReq, err := http.NewRequest("POST", uploadUrl, file)
	if err != nil {
		return fmt.Errorf("创建上传请求失败: %v", err)
	}
	uploadReq.Header.Set("Content-Type", "application/octet-stream")
	uploadReq.Header.Set("Content-Length", fmt.Sprintf("%d", fileSize))

	uploadResp, err := client.Do(uploadReq)
	if err != nil {
		return fmt.Errorf("上传失败: %v", err)
	}
	defer uploadResp.Body.Close()

	if uploadResp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(uploadResp.Body)
		return fmt.Errorf("上传请求被拒绝 (状态码 %d): %s", uploadResp.StatusCode, string(bodyBytes))
	}

	fmt.Printf("[发送端] 文件发送成功!\n")
	return nil
}
