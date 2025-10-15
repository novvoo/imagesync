package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// ======================
// 配置结构
// ======================

type Config struct {
	DB struct {
		Host     string
		Port     int
		User     string
		Password string
		Database string
	}
	Src struct {
		URLBase  string
		User     string
		Password string
		Type     string // "harbor" 或 "acr"
	}
	Dst struct {
		URLBase  string
		User     string
		Password string
		Type     string // "harbor" 或 "acr"
	}
}

func loadConfig() Config {
	var c Config

	// 优先从 .env 加载（如果存在则加载，不存在则忽略）
	_ = godotenv.Load()

	// 所有项均为必填，缺失将直接退出
	c.DB.Host = getRequiredEnv("DB_HOST")
	c.DB.Port = getRequiredEnvInt("DB_PORT")
	c.DB.User = getRequiredEnv("DB_USER")
	c.DB.Password = getRequiredEnv("DB_PASSWORD")
	c.DB.Database = getRequiredEnv("DB_NAME")

	c.Src.URLBase = strings.TrimSuffix(strings.TrimSpace(getRequiredEnv("SRC_URLBASE")), "/")
	c.Src.User = getRequiredEnv("SRC_USER")
	c.Src.Password = getRequiredEnv("SRC_PASSWORD")
	c.Src.Type = getRequiredEnv("SRC_TYPE")

	c.Dst.URLBase = strings.TrimSuffix(strings.TrimSpace(getRequiredEnv("DST_URLBASE")), "/")
	c.Dst.User = getRequiredEnv("DST_USER")
	c.Dst.Password = getRequiredEnv("DST_PASSWORD")
	c.Dst.Type = getRequiredEnv("DST_TYPE")

	return c
}

func getRequiredEnv(key string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		log.Fatalf("missing required environment variable: %s", key)
	}
	return v
}

func getRequiredEnvInt(key string) int {
	v := getRequiredEnv(key)
	var i int
	if _, err := fmt.Sscanf(v, "%d", &i); err != nil {
		log.Fatalf("invalid integer for %s: %v", key, err)
	}
	return i
}
