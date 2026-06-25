package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"sag-wiki/config"
)

// 对象存储服务
type MinIOService struct {
	client     *minio.Client
	bucketName string
}

// 创建 MinIO 服务
func NewMinIOService(cfg *config.MinIOConfig) (*MinIOService, error) {
	// 初始化 MinIO 客户端
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("初始化 MinIO 客户端失败: %w", err)
	}

	service := &MinIOService{
		client:     client,
		bucketName: cfg.BucketName,
	}

	// 确保 bucket 存在
	if err := service.ensureBucket(context.Background()); err != nil {
		return nil, err
	}

	log.Printf("✅ MinIO 连接成功，Bucket: %s", cfg.BucketName)
	return service, nil
}

// ensureBucket 确保 bucket 存在，不存在则创建
func (s *MinIOService) ensureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucketName)
	if err != nil {
		return fmt.Errorf("检查 bucket 失败: %w", err)
	}

	if !exists {
		err = s.client.MakeBucket(ctx, s.bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("创建 bucket 失败: %w", err)
		}
		log.Printf("✅ 创建 MinIO Bucket: %s", s.bucketName)
	}

	return nil
}

// 上传文件
func (s *MinIOService) UploadFile(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObject(ctx, s.bucketName, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("上传文件到 MinIO 失败: %w", err)
	}

	log.Printf("✅ 文件已上传到 MinIO: %s", objectName)
	return nil
}

// 获取文件
func (s *MinIOService) GetFile(ctx context.Context, objectName string) (*minio.Object, error) {
	object, err := s.client.GetObject(ctx, s.bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("从 MinIO 获取文件失败: %w", err)
	}

	return object, nil
}

// 删除文件
func (s *MinIOService) DeleteFile(ctx context.Context, objectName string) error {
	err := s.client.RemoveObject(ctx, s.bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("从 MinIO 删除文件失败: %w", err)
	}

	log.Printf("✅ 文件已从 MinIO 删除: %s", objectName)
	return nil
}

// 生成预签名下载 URL（有效期 1 小时）
func (s *MinIOService) GetPresignedURL(ctx context.Context, objectName string) (string, error) {
	url, err := s.client.PresignedGetObject(ctx, s.bucketName, objectName, time.Hour, nil)
	if err != nil {
		return "", fmt.Errorf("生成预签名 URL 失败: %w", err)
	}

	return url.String(), nil
}

// 读取文件内容为字符串
func (s *MinIOService) GetFileContent(ctx context.Context, objectName string) (string, error) {
	object, err := s.client.GetObject(ctx, s.bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return "", fmt.Errorf("从 MinIO 获取文件失败: %w", err)
	}
	defer object.Close()

	data, err := io.ReadAll(object)
	if err != nil {
		return "", fmt.Errorf("读取文件内容失败: %w", err)
	}

	return string(data), nil
}

// 将字符串内容写回 MinIO（覆盖原文件）
func (s *MinIOService) PutFileContent(ctx context.Context, objectName string, content string) error {
	reader := strings.NewReader(content)
	_, err := s.client.PutObject(ctx, s.bucketName, objectName, reader, int64(len(content)), minio.PutObjectOptions{
		ContentType: "text/markdown",
	})
	if err != nil {
		return fmt.Errorf("写入文件内容到 MinIO 失败: %w", err)
	}

	log.Printf("✅ 文件内容已更新: %s", objectName)
	return nil
}

// 检查文件是否存在
func (s *MinIOService) FileExists(ctx context.Context, objectName string) (bool, error) {
	_, err := s.client.StatObject(ctx, s.bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf("检查文件是否存在失败: %w", err)
	}

	return true, nil
}
