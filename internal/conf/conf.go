// Package conf 读取 configs/config.yaml（后台配置不复杂，直接用 yaml.v3）。
package conf

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 是全部配置项。
type Config struct {
	Server struct {
		Addr          string `yaml:"addr"`
		TimeoutSec    int    `yaml:"timeout_sec"`
		JWTSecret     string `yaml:"jwt_secret"`
		TokenTTLHours int    `yaml:"token_ttl_hours"`
	} `yaml:"server"`
	Database struct {
		Path string `yaml:"path"`
	} `yaml:"database"`
	Seed struct {
		Enabled bool `yaml:"enabled"`
	} `yaml:"seed"`
}

// Load 读配置文件，缺项用默认值补齐。
func Load(path string) (*Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件 %s 失败: %w", path, err)
	}
	var c Config
	if err := yaml.Unmarshal(raw, &c); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}
	if c.Server.Addr == "" {
		c.Server.Addr = ":8000"
	}
	if c.Server.TimeoutSec <= 0 {
		c.Server.TimeoutSec = 30
	}
	if c.Server.TokenTTLHours <= 0 {
		c.Server.TokenTTLHours = 72
	}
	if c.Server.JWTSecret == "" {
		c.Server.JWTSecret = "questioning-dev-secret"
	}
	if c.Database.Path == "" {
		c.Database.Path = "data/questioning.db"
	}
	return &c, nil
}
