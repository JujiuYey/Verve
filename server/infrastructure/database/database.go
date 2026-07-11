package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"

	learning_db "verve/app/learning/models/db"
	learning_repo "verve/app/learning/repository"
	rag_db "verve/app/rag/models/db"
	rag_repo "verve/app/rag/repository"
	system_db "verve/app/system/models/db"
	system_repo "verve/app/system/repository"
	wiki_db "verve/app/wiki/models/db"
	wiki_repo "verve/app/wiki/repository"
)

// 数据库服务（只负责连接管理）
type DatabaseService struct {
	db *bun.DB
	// System Repositories
	Users        *system_repo.UserRepository
	ModelConfigs system_repo.ModelConfigRepository

	// Learning Repositories
	Sessions *learning_repo.SessionRepository
	Messages *learning_repo.MessageRepository
	Reviews  *learning_repo.ReviewRepository
	Journals *learning_repo.JournalRepository
	Memories *learning_repo.MemoryRepository

	// Wiki Repositories
	Folders   wiki_repo.FolderRepository
	Documents *wiki_repo.DocumentRepository

	// RAG Repositories
	RAGChunks *rag_repo.ChunkRepository
	RAGJobs   *rag_repo.IndexJobRepository
}

// 创建数据库服务
func NewDatabaseService(dsn string) (*DatabaseService, error) {
	// 创建 PostgreSQL 连接
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))

	// 配置连接池
	sqldb.SetMaxOpenConns(25)                  // 最大打开连接数
	sqldb.SetMaxIdleConns(10)                  // 最大空闲连接数（增加以减少重连）
	sqldb.SetConnMaxLifetime(30 * time.Minute) // 连接最大生命周期（增加到30分钟）
	sqldb.SetConnMaxIdleTime(5 * time.Minute)  // 空闲连接最大生命周期（必须小于 MaxLifetime）

	// 创建 Bun DB 实例
	db := bun.NewDB(sqldb, pgdialect.New())

	// 注册模型
	db.RegisterModel((*system_db.User)(nil))
	db.RegisterModel((*system_db.SysModelPlatform)(nil))
	db.RegisterModel((*system_db.SysModel)(nil))
	db.RegisterModel((*system_db.AgentModelConfig)(nil))
	db.RegisterModel((*learning_db.LearningSession)(nil))
	db.RegisterModel((*learning_db.LearningMessage)(nil))
	db.RegisterModel((*learning_db.LearningExplanationReview)(nil))
	db.RegisterModel((*learning_db.LearningJournal)(nil))
	db.RegisterModel((*learning_db.LearningMemoryEvent)(nil))
	db.RegisterModel((*learning_db.LearningMemoryItem)(nil))
	db.RegisterModel((*learning_db.LearningMemorySummary)(nil))
	db.RegisterModel((*rag_db.WikiChunk)(nil))
	db.RegisterModel((*rag_db.IndexJob)(nil))
	db.RegisterModel((*wiki_db.Folder)(nil))
	db.RegisterModel((*wiki_db.Document)(nil))

	// 添加查询钩子（开发环境下打印 SQL）
	db.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithVerbose(true),
		bundebug.FromEnv("BUNDEBUG"),
	))

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("数据库连接失败: %w", err)
	}

	// 初始化所有 repositories
	return &DatabaseService{
		db: db,

		// System Repositories
		Users:        system_repo.NewUserRepository(db),
		ModelConfigs: system_repo.NewModelConfigRepository(db),

		// Learning Repositories
		Sessions: learning_repo.NewSessionRepository(db),
		Messages: learning_repo.NewMessageRepository(db),
		Reviews:  learning_repo.NewReviewRepository(db),
		Journals: learning_repo.NewJournalRepository(db),
		Memories: learning_repo.NewMemoryRepository(db),

		// Wiki Repositories
		Folders:   wiki_repo.NewFolderRepository(db),
		Documents: wiki_repo.NewDocumentRepository(db),

		// RAG Repositories
		RAGChunks: rag_repo.NewChunkRepository(db),
		RAGJobs:   rag_repo.NewIndexJobRepository(db),
	}, nil
}

// 关闭数据库连接
func (s *DatabaseService) Close() error {
	return s.db.Close()
}

// 获取 Bun DB 实例（供其他服务使用）
func (s *DatabaseService) GetDB() *bun.DB {
	return s.db
}
