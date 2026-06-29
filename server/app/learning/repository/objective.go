package repository

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	learning_db "sag-wiki/app/learning/models/db"
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

// 按路线列出小目标(按顺序)
func (r *ObjectiveRepository) FindByPath(ctx context.Context, pathID string) ([]*learning_db.LearningObjective, error) {
	var objectives []*learning_db.LearningObjective
	err := r.db.NewSelect().Model(&objectives).
		Where("path_id = ?", pathID).
		Order("order_index ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return objectives, nil
}

// 统计某路线下已完成 / 总数(用于进度)
func (r *ObjectiveRepository) CountByPath(ctx context.Context, pathID string) (completed, total int, err error) {
	total, err = r.db.NewSelect().Model((*learning_db.LearningObjective)(nil)).
		Where("path_id = ?", pathID).Count(ctx)
	if err != nil {
		return 0, 0, err
	}
	completed, err = r.db.NewSelect().Model((*learning_db.LearningObjective)(nil)).
		Where("path_id = ?", pathID).Where("status = ?", "completed").Count(ctx)
	if err != nil {
		return 0, 0, err
	}
	return completed, total, nil
}

// 批量创建(Planner 生成路线时)
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
