package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"

	rag_queue "verve/app/rag/queue"
	rag_service "verve/app/rag/service"
	"verve/config"
	"verve/infrastructure/database"
	"verve/infrastructure/storage"
	"verve/infrastructure/vector"
	"verve/router"
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

	vectorStore := vector.NewQdrantStore(config.GetQdrantConfig())
	embedder := rag_service.NewOpenAICompatibleEmbedder(dbService.ModelConfigs)
	indexer := rag_service.NewIndexer(dbService.RAGChunks, dbService.RAGJobs, dbService.Folders, dbService.Documents, minioService, embedder, vectorStore)

	redisConfig := config.GetRedisConfig()
	redisOpt := asynq.RedisClientOpt{
		Addr:     redisConfig.Addr,
		Password: redisConfig.Password,
		DB:       redisConfig.DB,
	}
	asynqClient := asynq.NewClient(redisOpt)
	defer asynqClient.Close()
	ragEnqueuer := rag_queue.NewEnqueuer(
		asynqClient,
		dbService.RAGBatches,
		dbService.RAGJobs,
		dbService.Folders,
		dbService.Documents,
		indexer,
	)
	ragProcessor := rag_queue.NewProcessor(indexer, dbService.RAGBatches)
	ragWorker := asynq.NewServer(redisOpt, asynq.Config{
		Concurrency: 1,
		Queues:      map[string]int{rag_queue.QueueRAG: 1},
		RetryDelayFunc: func(n int, err error, task *asynq.Task) time.Duration {
			return asynq.DefaultRetryDelayFunc(n, err, task)
		},
	})
	ragMux := asynq.NewServeMux()
	ragMux.HandleFunc(rag_queue.TypeIndexDocument, ragProcessor.HandleIndexDocument)
	go func() {
		if err := ragWorker.Run(ragMux); err != nil {
			log.Printf("⚠️  RAG 队列 worker 已停止: %v", err)
		}
	}()

	// 设置路由
	app := router.SetupRouter(dbService, minioService, vectorStore, ragEnqueuer)

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
	ragWorker.Shutdown()
}
