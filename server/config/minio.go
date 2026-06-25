package config

import (
	"os"

	"sag-wiki/utils"
)

// MinIO 配置
type MinIOConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	BucketName      string
}

// 获取 MinIO 配置
func GetMinIOConfig() *MinIOConfig {
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"

	return &MinIOConfig{
		Endpoint:        utils.GetEnv("MINIO_ENDPOINT", "localhost:9000"),
		AccessKeyID:     utils.GetEnv("MINIO_ACCESS_KEY", "minioadmin"),
		SecretAccessKey: utils.GetEnv("MINIO_SECRET_KEY", "minioadmin"),
		UseSSL:          useSSL,
		BucketName:      utils.GetEnv("MINIO_BUCKET_NAME", "documents"),
	}
}
