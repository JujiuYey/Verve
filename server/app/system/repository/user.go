package repository

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	system_db "sag-wiki/app/system/models/db"
)

// 用户数据访问层
type UserRepository struct {
	db *bun.DB
}

// 创建用户仓库
func NewUserRepository(database *bun.DB) *UserRepository {
	return &UserRepository{db: database}
}

// 获取数据库连接
func (r *UserRepository) GetDB() *bun.DB {
	return r.db
}

// 根据用户名查找用户
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*system_db.User, error) {
	user := new(system_db.User)
	err := r.db.NewSelect().
		Model(user).
		Where("username = ?", username).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// 根据 ID 查找用户
func (r *UserRepository) FindOne(ctx context.Context, id string) (*system_db.User, error) {
	user := new(system_db.User)
	err := r.db.NewSelect().
		Model(user).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// 分页获取用户列表
func (r *UserRepository) FindPage(ctx context.Context, offset, limit int, search string) ([]*system_db.User, int, error) {
	var users []*system_db.User

	query := r.db.NewSelect().
		Model(&users).
		Order("created_at DESC")

	if search != "" {
		query = query.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.WhereOr("username LIKE ?", "%"+search+"%").
				WhereOr("email LIKE ?", "%"+search+"%").
				WhereOr("full_name LIKE ?", "%"+search+"%")
		})
	}

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	err = query.Offset(offset).Limit(limit).Scan(ctx)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// 创建用户
func (r *UserRepository) Create(ctx context.Context, user *system_db.User) error {
	user.ID = strings.ReplaceAll(uuid.New().String(), "-", "")

	_, err := r.db.NewInsert().Model(user).Exec(ctx)
	return err
}

// 更新用户
func (r *UserRepository) Update(ctx context.Context, user *system_db.User) error {
	_, err := r.db.NewUpdate().
		Model(user).
		Where("id = ?", user.ID).
		Exec(ctx)
	return err
}

// 删除用户
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().
		Model((*system_db.User)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

// 搜索用户
func (r *UserRepository) SearchUsers(ctx context.Context, query string) ([]*system_db.User, error) {
	var users []*system_db.User

	err := r.db.NewSelect().
		Model(&users).
		WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.WhereOr("username LIKE ?", "%"+query+"%").
				WhereOr("email LIKE ?", "%"+query+"%").
				WhereOr("full_name LIKE ?", "%"+query+"%")
		}).
		Order("created_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return users, nil
}
