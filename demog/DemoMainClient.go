package main

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

type MulticastDto struct {
	Alias        string `json:"alias"`
	DeviceModel  string `json:"deviceModel"`
	Fingerprint  string `json:"fingerprint"`
	Port         int    `json:"port"`
	Announcement bool   `json:"announcement"`
}

func main() {
	// 多播地址
	group := "224.0.0.167"
	port := 53317
	addr, _ := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", group, port))

	conn, _ := net.DialUDP("udp", nil, addr)
	defer conn.Close()

	dto := MulticastDto{
		Alias:        "模拟设备A",
		DeviceModel:  "GoSim-1",
		Fingerprint:  "fake-fingerprint-123",
		Port:         53317,
		Announcement: true,
	}

	data, _ := json.Marshal(dto)

	fmt.Printf("发送模拟多播: %s\n", string(data))
	for i := 0; i < 5; i++ {
		conn.Write(data)
		time.Sleep(2 * time.Second)
	}
}
