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

	ai_db "sag-wiki/app/ai/models/db"
	ai_repo "sag-wiki/app/ai/repository"
	system_db "sag-wiki/app/system/models/db"
	system_repo "sag-wiki/app/system/repository"
	wiki_db "sag-wiki/app/wiki/models/db"
	wiki_repo "sag-wiki/app/wiki/repository"
)

// 数据库服务（只负责连接管理）
type DatabaseService struct {
	db *bun.DB

	// Wiki Repositories
	Documents *wiki_repo.DocumentRepository
	Folders   wiki_repo.FolderRepository

	// AI Repositories
	AgentSessions *ai_repo.AgentSessionRepository
	AgentMessages *ai_repo.AgentMessageRepository
	RagSessions   *ai_repo.RagSessionRepository
	RagMessages   *ai_repo.RagMessageRepository
	ModelConfigs  ai_repo.ModelConfigRepository

	// System Repositories
	Users       *system_repo.UserRepository
	Roles       system_repo.RoleRepository
	Departments system_repo.DepartmentRepository
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
	db.RegisterModel((*wiki_db.Document)(nil))
	db.RegisterModel((*wiki_db.Folder)(nil))
	db.RegisterModel((*wiki_db.FolderPermission)(nil))
	db.RegisterModel((*ai_db.AgentSession)(nil))
	db.RegisterModel((*ai_db.AgentMessage)(nil))
	db.RegisterModel((*ai_db.RagSession)(nil))
	db.RegisterModel((*ai_db.RagMessage)(nil))
	db.RegisterModel((*system_db.User)(nil))
	db.RegisterModel((*system_db.Role)(nil))
	db.RegisterModel((*system_db.Department)(nil))
	db.RegisterModel((*system_db.UserRole)(nil))
	db.RegisterModel((*system_db.UserDepartment)(nil))

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

		// Wiki Repositories
		Documents: wiki_repo.NewDocumentRepository(db),
		Folders:   wiki_repo.NewFolderRepository(db),

		// AI Repositories
		AgentSessions: ai_repo.NewAgentSessionRepository(db),
		AgentMessages: ai_repo.NewAgentMessageRepository(db),
		RagSessions:   ai_repo.NewRagSessionRepository(db),
		RagMessages:   ai_repo.NewRagMessageRepository(db),
		ModelConfigs:  ai_repo.NewModelConfigRepository(db),

		// System Repositories
		Users:       system_repo.NewUserRepository(db),
		Roles:       system_repo.NewRoleRepository(db),
		Departments: system_repo.NewDepartmentRepository(db),
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
