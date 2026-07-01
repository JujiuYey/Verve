package repository

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	learning_db "verve/app/learning/models/db"
)

// 小目标数据访问层
type ObjectiveRepository struct {
	db *bun.DB
}

func NewObjectiveRepository(database *bun.DB) *ObjectiveRepository {
	return &ObjectiveRepository{db: database}
}

func (r *ObjectiveRepository) GetDB() *bun.DB { return r.db }

func (r *ObjectiveRepository) FindOne(ctx context.Context, id string) (*learning_db.LearningObjective, error) {
	obj := new(learning_db.LearningObjective)
	if err := r.db.NewSelect().Model(obj).Where("id = ?", id).Scan(ctx); err != nil {
		return nil, err
	}
	return obj, nil
}

// 按 Wiki 文件夹列出学习小节(按顺序)
func (r *ObjectiveRepository) FindByFolder(ctx context.Context, folderID string) ([]*learning_db.LearningObjective, error) {
	var objectives []*learning_db.LearningObjective
	err := r.db.NewSelect().Model(&objectives).
		Where("source_folder_id = ?", folderID).
		Order("order_index ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return objectives, nil
}

// 按用户列出最近学习小节,用于学习调度入口在未选文件夹时找上下文。
func (r *ObjectiveRepository) FindRecentByUser(ctx context.Context, userID string, limit int) ([]*learning_db.LearningObjective, error) {
	var objectives []*learning_db.LearningObjective
	query := r.db.NewSelect().Model(&objectives).
		Where("user_id = ?", userID).
		Order("updated_at DESC", "order_index ASC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Scan(ctx); err != nil {
		return nil, err
	}
	return objectives, nil
}

// 统计某 Wiki 文件夹下已完成 / 总数(用于进度)
func (r *ObjectiveRepository) CountByFolder(ctx context.Context, folderID string) (completed, total int, err error) {
	total, err = r.db.NewSelect().Model((*learning_db.LearningObjective)(nil)).
		Where("source_folder_id = ?", folderID).Count(ctx)
	if err != nil {
		return 0, 0, err
	}
	completed, err = r.db.NewSelect().Model((*learning_db.LearningObjective)(nil)).
		Where("source_folder_id = ?", folderID).Where("status = ?", "completed").Count(ctx)
	if err != nil {
		return 0, 0, err
	}
	return completed, total, nil
}

// 批量创建学习小节
func (r *ObjectiveRepository) BulkCreate(ctx context.Context, objectives []*learning_db.LearningObjective) error {
	if len(objectives) == 0 {
		return nil
	}
	for _, o := range objectives {
		o.ID = strings.ReplaceAll(uuid.New().String(), "-", "")
	}
	_, err := r.db.NewInsert().Model(&objectives).Exec(ctx)
	return err
}

func (r *ObjectiveRepository) Update(ctx context.Context, obj *learning_db.LearningObjective) error {
	_, err := r.db.NewUpdate().Model(obj).Where("id = ?", obj.ID).Exec(ctx)
	return err
}
