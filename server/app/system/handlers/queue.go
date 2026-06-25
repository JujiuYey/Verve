package handlers

import (
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"sag-wiki/common/response"
	"sag-wiki/infrastructure/queue"
)

// 队列监控处理器
type QueueHandler struct {
	taskQueue *queue.TaskQueue
}

// 创建队列监控处理器
func NewQueueHandler(taskQueue *queue.TaskQueue) *QueueHandler {
	return &QueueHandler{
		taskQueue: taskQueue,
	}
}

// 获取队列统计信息
func (h *QueueHandler) GetStats(c *fiber.Ctx) error {
	queueName := c.Query("queue", "default")
	log.Printf("📊 [QueueHandler.GetStats] 开始获取队列统计, queue=%s", queueName)

	if h.taskQueue == nil {
		log.Printf("❌ [QueueHandler.GetStats] taskQueue 为 nil")
		return response.InternalServerCtx(c, "队列服务未初始化")
	}

	stats, err := h.taskQueue.GetQueueStats(c.Context(), queueName)
	if err != nil {
		log.Printf("❌ [QueueHandler.GetStats] 获取队列统计失败: %v", err)
		return response.InternalServerCtx(c, "获取队列统计失败: "+err.Error())
	}

	log.Printf("✅ [QueueHandler.GetStats] 获取队列统计成功: %+v", stats)
	return c.JSON(stats)
}

// 获取任务列表
func (h *QueueHandler) GetTasks(c *fiber.Ctx) error {
	queueName := c.Query("queue", "default")
	state := c.Query("state", "pending") // pending, active, scheduled, retry, archived
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))
	page, _ := strconv.Atoi(c.Query("page", "0"))

	log.Printf("📋 [QueueHandler.GetTasks] 开始获取任务列表, queue=%s, state=%s, pageSize=%d, page=%d",
		queueName, state, pageSize, page)

	if h.taskQueue == nil {
		log.Printf("❌ [QueueHandler.GetTasks] taskQueue 为 nil")
		return response.InternalServerCtx(c, "队列服务未初始化")
	}

	var tasks []*queue.TaskInfo
	var err error

	switch state {
	case "pending":
		log.Printf("🔍 [QueueHandler.GetTasks] 查询 pending 任务")
		tasks, err = h.taskQueue.ListPendingTasks(c.Context(), queueName, pageSize, page)
	case "active":
		log.Printf("🔍 [QueueHandler.GetTasks] 查询 active 任务")
		tasks, err = h.taskQueue.ListActiveTasks(c.Context(), queueName)
	case "scheduled":
		log.Printf("🔍 [QueueHandler.GetTasks] 查询 scheduled 任务")
		tasks, err = h.taskQueue.ListScheduledTasks(c.Context(), queueName, pageSize, page)
	case "retry":
		log.Printf("🔍 [QueueHandler.GetTasks] 查询 retry 任务")
		tasks, err = h.taskQueue.ListRetryTasks(c.Context(), queueName, pageSize, page)
	case "archived":
		log.Printf("🔍 [QueueHandler.GetTasks] 查询 archived 任务")
		tasks, err = h.taskQueue.ListArchivedTasks(c.Context(), queueName, pageSize, page)
	default:
		log.Printf("⚠️ [QueueHandler.GetTasks] 未知的任务状态: %s", state)
		return response.BadRequestCtx(c, "未知的任务状态")
	}

	if err != nil {
		log.Printf("❌ [QueueHandler.GetTasks] 获取任务列表失败: %v", err)
		return response.InternalServerCtx(c, "获取任务列表失败: "+err.Error())
	}

	log.Printf("✅ [QueueHandler.GetTasks] 获取任务列表成功，共 %d 个任务", len(tasks))
	return c.JSON(tasks)
}
