package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	ConfigFileName = "config.yaml"
)

type Config struct {
	Network  NetworkConfig  `yaml:"network"`
	Device   DeviceConfig   `yaml:"device"`
	Transfer TransferConfig `yaml:"transfer"`
}

type NetworkConfig struct {
	IsHttps        bool   `yaml:"is_https"`
	Port           int    `yaml:"port"`
	MulticastGroup string `yaml:"multicast_group"`
	ConnectTimeout int    `yaml:"connect_timeout_seconds"`
}

type DeviceConfig struct {
	Alias       string `yaml:"alias"`
	DeviceModel string `yaml:"device_model"`
}

type TransferConfig struct {
	DownloadDir string `yaml:"download_dir"`
}

var (
	cfg *Config
)

func Load() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	configPath := getConfigPath()

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		cfg = getDefaultConfig()
		if err := Save(cfg); err != nil {
			return nil, fmt.Errorf("创建默认配置文件失败: %w", err)
		}
		fmt.Printf("[配置] 配置文件不存在，已创建默认配置: %s\n", configPath)
	} else {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("读取配置文件失败: %w", err)
		}

		cfg = &Config{}
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("解析配置文件失败: %w", err)
		}
		fmt.Printf("[配置] 已加载配置文件: %s\n", configPath)
	}

	return cfg, nil
}

func Save(config *Config) error {
	configPath := getConfigPath()
	configDir := filepath.Dir(configPath)

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

func getConfigPath() string {
	return filepath.Join(getCurrentDir(), ConfigFileName)
}

func getCurrentDir() string {
	if dir, err := os.Getwd(); err == nil {
		return dir
	}
	return "."
}

func getDefaultConfig() *Config {
	return &Config{
		Network: NetworkConfig{
			IsHttps:        true,
			Port:           53317,
			MulticastGroup: "224.0.0.167",
			ConnectTimeout: 60,
		},
		Device: DeviceConfig{
			Alias:       "局域网共享传输",
			DeviceModel: runtime.GOOS,
		},
		Transfer: TransferConfig{
			DownloadDir: "downloads",
		},
	}
}

func Get() *Config {
	if cfg == nil {
		cfg, _ = Load()
	}
	return cfg
}

func GetConnectTimeout() time.Duration {
	return time.Duration(Get().Network.ConnectTimeout) * time.Second
}
