package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hibiken/asynq"

	ai_service "sag-wiki/app/ai/service"
	wiki_db "sag-wiki/app/wiki/models/db"
	wiki_repo "sag-wiki/app/wiki/repository"
	"sag-wiki/infrastructure/database"
	qdrantdao "sag-wiki/infrastructure/qdrant"
	"sag-wiki/infrastructure/storage"
)

const (
	TypeDocumentProcess = "document:process"
)

// 文档处理任务的载荷
type DocumentProcessPayload struct {
	DocumentID string `json:"document_id"`
	FilePath   string `json:"file_path"`
	Filename   string `json:"filename"`
	FolderID   string `json:"folder_id"`
}

// 任务队列服务
type TaskQueue struct {
	client    *asynq.Client
	inspector *asynq.Inspector
}

// 创建任务队列客户端
func NewTaskQueue(redisAddr, redisPassword string, redisDB int) *TaskQueue {
	redisOpt := asynq.RedisClientOpt{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	}

	client := asynq.NewClient(redisOpt)
	inspector := asynq.NewInspector(redisOpt)

	return &TaskQueue{
		client:    client,
		inspector: inspector,
	}
}

// 将文档处理任务加入队列
func (tq *TaskQueue) EnqueueDocumentProcess(ctx context.Context, docID string, filePath, filename, folderID string) error {
	payload := DocumentProcessPayload{
		DocumentID: docID,
		FilePath:   filePath,
		Filename:   filename,
		FolderID:   folderID,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化任务载荷失败: %w", err)
	}

	task := asynq.NewTask(TypeDocumentProcess, payloadBytes)

	// 入队任务
	info, err := tq.client.Enqueue(task, asynq.MaxRetry(3))
	if err != nil {
		return fmt.Errorf("任务入队失败: %w", err)
	}

	log.Printf("✅ 任务已入队: ID=%s, Queue=%s", info.ID, info.Queue)
	return nil
}

// 关闭客户端
func (tq *TaskQueue) Close() error {
	tq.inspector.Close()
	return tq.client.Close()
}

// 队列统计信息
type QueueStats struct {
	QueueName string `json:"queue_name"`
	Pending   int    `json:"pending"`   // 等待处理
	Active    int    `json:"active"`    // 正在处理
	Scheduled int    `json:"scheduled"` // 计划执行
	Retry     int    `json:"retry"`     // 重试中
	Archived  int    `json:"archived"`  // 已归档（失败）
	Completed int    `json:"completed"` // 已完成
}

// 任务详细信息
type TaskInfo struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Payload     map[string]interface{} `json:"payload"`
	Queue       string                 `json:"queue"`
	MaxRetry    int                    `json:"max_retry"`
	Retried     int                    `json:"retried"`
	LastError   string                 `json:"last_error,omitempty"`
	State       string                 `json:"state"`
	NextProcess string                 `json:"next_process,omitempty"`
}

// 获取队列统计信息
func (tq *TaskQueue) GetQueueStats(ctx context.Context, queueName string) (*QueueStats, error) {
	log.Printf("📊 [TaskQueue.GetQueueStats] 开始获取队列信息, queue=%s", queueName)

	if tq.inspector == nil {
		log.Printf("❌ [TaskQueue.GetQueueStats] inspector 为 nil")
		return nil, fmt.Errorf("inspector 未初始化")
	}

	info, err := tq.inspector.GetQueueInfo(queueName)
	if err != nil {
		// 如果队列不存在，返回空统计信息而不是错误
		if err.Error() == "asynq: queue not found" || err.Error() == "NOT_FOUND: queue \""+queueName+"\" does not exist" {
			log.Printf("⚠️  [TaskQueue.GetQueueStats] 队列不存在，返回空统计: %s", queueName)
			return &QueueStats{
				QueueName: queueName,
				Pending:   0,
				Active:    0,
				Scheduled: 0,
				Retry:     0,
				Archived:  0,
				Completed: 0,
			}, nil
		}
		log.Printf("❌ [TaskQueue.GetQueueStats] 获取队列信息失败: %v", err)
		return nil, fmt.Errorf("获取队列信息失败: %w", err)
	}

	stats := &QueueStats{
		QueueName: queueName,
		Pending:   info.Pending,
		Active:    info.Active,
		Scheduled: info.Scheduled,
		Retry:     info.Retry,
		Archived:  info.Archived,
		Completed: info.Processed, // asynq 使用 Processed 表示已完成
	}

	log.Printf("✅ [TaskQueue.GetQueueStats] 获取队列信息成功: pending=%d, active=%d, scheduled=%d, retry=%d, archived=%d, completed=%d",
		stats.Pending, stats.Active, stats.Scheduled, stats.Retry, stats.Archived, stats.Completed)

	return stats, nil
}

// 列出等待中的任务
func (tq *TaskQueue) ListPendingTasks(ctx context.Context, queueName string, pageSize int, page int) ([]*TaskInfo, error) {
	log.Printf("📋 [TaskQueue.ListPendingTasks] queue=%s, pageSize=%d, page=%d", queueName, pageSize, page)

	if tq.inspector == nil {
		log.Printf("❌ [TaskQueue.ListPendingTasks] inspector 为 nil")
		return nil, fmt.Errorf("inspector 未初始化")
	}

	tasks, err := tq.inspector.ListPendingTasks(queueName, asynq.PageSize(pageSize), asynq.Page(page))
	if err != nil {
		// 如果队列不存在，返回空列表而不是错误
		if err.Error() == "asynq: queue not found" {
			log.Printf("⚠️  [TaskQueue.ListPendingTasks] 队列不存在，返回空列表: %s", queueName)
			return []*TaskInfo{}, nil
		}
		log.Printf("❌ [TaskQueue.ListPendingTasks] 获取等待任务失败: %v", err)
		return nil, fmt.Errorf("获取等待任务失败: %w", err)
	}

	result := tq.convertTaskInfos(tasks)
	log.Printf("✅ [TaskQueue.ListPendingTasks] 获取到 %d 个任务", len(result))
	return result, nil
}

// 列出正在处理的任务
func (tq *TaskQueue) ListActiveTasks(ctx context.Context, queueName string) ([]*TaskInfo, error) {
	tasks, err := tq.inspector.ListActiveTasks(queueName)
	if err != nil {
		// 如果队列不存在，返回空列表而不是错误
		if err.Error() == "asynq: queue not found" {
			return []*TaskInfo{}, nil
		}
		return nil, fmt.Errorf("获取活动任务失败: %w", err)
	}

	return tq.convertTaskInfos(tasks), nil
}

// 列出计划执行的任务
func (tq *TaskQueue) ListScheduledTasks(ctx context.Context, queueName string, pageSize int, page int) ([]*TaskInfo, error) {
	tasks, err := tq.inspector.ListScheduledTasks(queueName, asynq.PageSize(pageSize), asynq.Page(page))
	if err != nil {
		// 如果队列不存在，返回空列表而不是错误
		if err.Error() == "asynq: queue not found" {
			return []*TaskInfo{}, nil
		}
		return nil, fmt.Errorf("获取计划任务失败: %w", err)
	}

	return tq.convertTaskInfos(tasks), nil
}

// 列出重试中的任务
func (tq *TaskQueue) ListRetryTasks(ctx context.Context, queueName string, pageSize int, page int) ([]*TaskInfo, error) {
	tasks, err := tq.inspector.ListRetryTasks(queueName, asynq.PageSize(pageSize), asynq.Page(page))
	if err != nil {
		// 如果队列不存在，返回空列表而不是错误
		if err.Error() == "asynq: queue not found" {
			return []*TaskInfo{}, nil
		}
		return nil, fmt.Errorf("获取重试任务失败: %w", err)
	}

	return tq.convertTaskInfos(tasks), nil
}

// 列出已归档（失败）的任务
func (tq *TaskQueue) ListArchivedTasks(ctx context.Context, queueName string, pageSize int, page int) ([]*TaskInfo, error) {
	tasks, err := tq.inspector.ListArchivedTasks(queueName, asynq.PageSize(pageSize), asynq.Page(page))
	if err != nil {
		// 如果队列不存在，返回空列表而不是错误
		if err.Error() == "asynq: queue not found" {
			return []*TaskInfo{}, nil
		}
		return nil, fmt.Errorf("获取归档任务失败: %w", err)
	}

	return tq.convertTaskInfos(tasks), nil
}

// convertTaskInfos 转换任务信息
func (tq *TaskQueue) convertTaskInfos(tasks []*asynq.TaskInfo) []*TaskInfo {
	result := make([]*TaskInfo, 0, len(tasks))
	for _, t := range tasks {
		var payload map[string]interface{}
		json.Unmarshal(t.Payload, &payload)

		state := "unknown"
		switch t.State {
		case asynq.TaskStatePending:
			state = "pending"
		case asynq.TaskStateActive:
			state = "active"
		case asynq.TaskStateScheduled:
			state = "scheduled"
		case asynq.TaskStateRetry:
			state = "retry"
		case asynq.TaskStateArchived:
			state = "archived"
		case asynq.TaskStateCompleted:
			state = "completed"
		}

		info := &TaskInfo{
			ID:        t.ID,
			Type:      t.Type,
			Payload:   payload,
			Queue:     t.Queue,
			MaxRetry:  t.MaxRetry,
			Retried:   t.Retried,
			LastError: t.LastErr,
			State:     state,
		}

		if !t.NextProcessAt.IsZero() {
			info.NextProcess = t.NextProcessAt.Format("2006-01-02 15:04:05")
		}

		result = append(result, info)
	}
	return result
}

// 任务处理器
type TaskWorker struct {
	server            *asynq.Server
	mux               *asynq.ServeMux
	dbService         *database.DatabaseService
	documentProcessor *ai_service.DocumentProcessor
}

// 创建任务处理器
func NewTaskWorker(
	redisAddr, redisPassword string,
	redisDB int,
	dbService *database.DatabaseService,
	embeddingService *ai_service.EmbeddingService,
	chunkDAO *qdrantdao.ChunkDAO,
	minioService *storage.MinIOService,
	documentRepo *wiki_repo.DocumentRepository,
) *TaskWorker {
	server := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     redisAddr,
			Password: redisPassword,
			DB:       redisDB,
		},
		asynq.Config{
			Concurrency: 10, // 并发处理10个任务
			Queues: map[string]int{
				"default": 1,
			},
		},
	)

	mux := asynq.NewServeMux()

	worker := &TaskWorker{
		server:    server,
		mux:       mux,
		dbService: dbService,
		documentProcessor: ai_service.NewDocumentProcessor(
			embeddingService,
			chunkDAO,
			minioService,
			documentRepo,
		),
	}

	// 注册任务处理函数
	mux.HandleFunc(TypeDocumentProcess, worker.HandleDocumentProcess)

	return worker
}

// 处理文档处理任务
func (tw *TaskWorker) HandleDocumentProcess(ctx context.Context, task *asynq.Task) error {
	var payload DocumentProcessPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("解析任务载荷失败: %w", err)
	}

	log.Printf("🔄 开始处理文档任务: %s", payload.DocumentID)

	// 调用 document processor
	result := tw.documentProcessor.ReprocessDocument(
		ctx,
		payload.DocumentID,
		payload.FilePath,
		payload.Filename,
		payload.FolderID,
	)

	if result.Error != nil {
		// 更新状态为失败
		errMsg := result.Error.Error()
		tw.documentProcessor.DocumentRepo().UpdateStatus(
			ctx,
			payload.DocumentID,
			wiki_db.DocumentStatusFailed,
			0,
			&errMsg,
		)
		return result.Error
	}

	log.Printf("✅ 文档处理成功: %s, chunks: %d", payload.DocumentID, result.ChunkCount)
	return nil
}

// 启动任务处理器
func (tw *TaskWorker) Start() error {
	log.Println("🚀 任务处理器启动中...")
	return tw.server.Run(tw.mux)
}

// 优雅关闭
func (tw *TaskWorker) Shutdown() {
	log.Println("⏹️  任务处理器关闭中...")
	tw.server.Shutdown()
}
