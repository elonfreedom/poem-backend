package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	WebAuthn WebAuthnConfig
}

type ServerConfig struct {
	Port int
	Mode string // debug, release
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type JWTConfig struct {
	Secret     string
	ExpireHour int
}

// WebAuthnConfig WebAuthn/Passkey 配置
type WebAuthnConfig struct {
	RPDisplayName string // 显示名称
	RPID          string // 域名
	RPOrigin      string // 来源地址
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnvAsInt("SERVER_PORT", 8080),
			Mode: getEnv("SERVER_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "poem"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "your-secret-key"),
			ExpireHour: getEnvAsInt("JWT_EXPIRE_HOUR", 72),
		},
		WebAuthn: WebAuthnConfig{
			RPDisplayName: getEnv("RP_DISPLAY_NAME", "晓诗"),
			RPID:          getEnv("RP_ID", "localhost"),
			RPOrigin:      getEnv("RP_ORIGIN", "http://localhost:3000"),
		},
	}
}

// ConnString 返回数据库连接字符串
func (d DatabaseConfig) ConnString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode)
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}
