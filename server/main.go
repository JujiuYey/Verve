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

	// 初始化 Redis 和任务队列
	// redisConfig := config.GetRedisConfig()
	// taskQueue := queue.NewTaskQueue(redisConfig.Addr, redisConfig.Password, redisConfig.DB)
	// defer taskQueue.Close()
	// log.Println("✅ 任务队列初始化成功")

	// 初始化 Qdrant ChunkDAO
	// chunkDAO, err := qdrantdao.NewChunkDAO()
	// if err != nil {
	// 	log.Fatalf("❌ Qdrant ChunkDAO 初始化失败: %v", err)
	// }
	// log.Println("✅ Qdrant ChunkDAO 初始化成功")

	// 初始化 RetrievalService (RAG 检索服务)
	// folderPermissionRepo := repository.NewFolderPermissionRepository(dbService.GetDB())
	// folderExpander := repository.NewFolderExpander(dbService.Folders, folderPermissionRepo)
	// retrievalService := ai_service.NewRetrievalService(
	// 	ai_service.NewEmbeddingService(dbService.ModelConfigs),
	// 	chunkDAO,
	// 	folderExpander,
	// 	*dbService.Documents,
	// )
	// log.Println("✅ RetrievalService 初始化成功")

	// 启动 Task Worker
	// taskWorker := queue.NewTaskWorker(
	// 	redisConfig.Addr,
	// 	redisConfig.Password,
	// 	redisConfig.DB,
	// 	dbService,
	// 	ai_service.NewEmbeddingService(dbService.ModelConfigs),
	// 	chunkDAO,
	// 	minioService,
	// 	dbService.Documents,
	// )
	// go taskWorker.Start()
	// log.Println("✅ Task Worker 启动成功")

	// 设置路由
	app := router.SetupRouter(dbService, minioService, nil)

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
