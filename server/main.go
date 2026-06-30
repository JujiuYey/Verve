package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"sag-wiki/config"
	"sag-wiki/infrastructure/database"
	"sag-wiki/infrastructure/storage"
	"sag-wiki/router"
)

//go:generate msgp -file=../../models/api/ai/chat.go -o=../../models/api/ai/chat_msgp.go -tests=false
//go:generate msgp -file=../../models/api/ai/rag.go -o=../../models/api/ai/rag_msgp.go -tests=false

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Printf("⚠️  未找到 .env 文件: %v", err)
	}

	// 初始化数据库
	dbConfig := config.GetDatabaseConfig()
	dbService, err := database.NewDatabaseService(dbConfig.GetDSN())
	if err != nil {
		log.Fatalf("❌ 数据库初始化失败: %v", err)
	}
	defer dbService.Close()
	log.Println("✅ 数据库连接成功")

	// 初始化 MinIO
	minioConfig := config.GetMinIOConfig()
	minioService, err := storage.NewMinIOService(minioConfig)
	if err != nil {
		log.Fatalf("❌ MinIO 初始化失败: %v", err)
	}

	// 设置路由
	app := router.SetupRouter(dbService, minioService)

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		// 启动 Fiber 服务器
		log.Println("🚀 Go RAG 后端服务启动在 http://localhost:8080")
		if err := app.Listen(":8080"); err != nil {
			log.Fatal("启动服务器失败:", err)
		}
	}()

	<-quit
	log.Println("⏹️  服务器正在关闭...")
}
