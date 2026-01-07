package main

import (
	"chrelyonly-localsend-go/config"
	"chrelyonly-localsend-go/model"
	"time"
)

const (
	ProtocolVersion = "2.1"

	UDPBufferSize = 65535

	UDPSocketBufferSize = 1024 * 1024
)

var (
	IsHttps               bool
	DefaultPort           int
	DefaultMulticastGroup string
	DefaultAlias          string
	DefaultDeviceModel    string
	ConnectTimeout        time.Duration
	DefaultDownloadDir    string

	ProtocolTypeHttpStatus                    = ProtocolTypeHttp
	ProtocolTypeHttp       model.ProtocolType = "http"
	ProtocolTypeHttps      model.ProtocolType = "https"
)

func init() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	IsHttps = cfg.Network.IsHttps
	DefaultPort = cfg.Network.Port
	DefaultMulticastGroup = cfg.Network.MulticastGroup
	DefaultAlias = cfg.Device.Alias
	DefaultDeviceModel = cfg.Device.DeviceModel
	ConnectTimeout = config.GetConnectTimeout()
	DefaultDownloadDir = cfg.Transfer.DownloadDir

	if IsHttps {
		ProtocolTypeHttpStatus = ProtocolTypeHttps
	} else {
		ProtocolTypeHttpStatus = ProtocolTypeHttp
	}
}
