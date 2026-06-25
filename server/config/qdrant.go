package config

import "sag-wiki/utils"

// Qdrant 配置
type QdrantConfig struct {
	URL string
}

// 从环境变量获取 Qdrant 配置
func GetQdrantConfig() *QdrantConfig {
	return &QdrantConfig{
		URL: utils.GetEnv("QDRANT_URL", "http://localhost:6333"),
	}
}
