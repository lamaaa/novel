package config

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	DB     DBConfig     `yaml:"db"`
}

type ServerConfig struct {
	Port     string `yaml:"port"`
	MCPToken string `yaml:"mcp_token"`
}

type DBConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
}

func Load() *Config {
	configPath := "config.yaml"
	if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
		configPath = envPath
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Failed to read config file %s: %v", configPath, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}

	// 环境变量覆盖数据库配置（适配 Docker 部署）
	if v := os.Getenv("DB_HOST"); v != "" {
		cfg.DB.Host = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		cfg.DB.Port = v
	}
	if v := os.Getenv("DB_USER"); v != "" {
		cfg.DB.User = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		cfg.DB.Password = v
	}
	if v := os.Getenv("DB_NAME"); v != "" {
		cfg.DB.DBName = v
	}
	if v := os.Getenv("SERVER_PORT"); v != "" {
		cfg.Server.Port = v
	}
	if v := os.Getenv("MCP_TOKEN"); v != "" {
		cfg.Server.MCPToken = v
	}

	return &cfg
}

func (c *DBConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.User, c.Password, c.Host, c.Port, c.DBName)
}
