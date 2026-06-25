package repository

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	system_db "sag-wiki/app/system/models/db"
)

// 角色仓储接口
type RoleRepository interface {
	FindOne(ctx context.Context, id string) (*system_db.Role, error)
	FindPage(ctx context.Context, offset, limit int) ([]*system_db.Role, int, error)
	FindList(ctx context.Context) ([]*system_db.Role, error)
	Create(ctx context.Context, role *system_db.Role) error
	Update(ctx context.Context, role *system_db.Role) error
	Delete(ctx context.Context, id string) error
}

type roleRepository struct {
	db *bun.DB
}

// 创建角色仓储实例
func NewRoleRepository(db *bun.DB) RoleRepository {
	return &roleRepository{db: db}
}

// 根据ID获取角色
func (r *roleRepository) FindOne(ctx context.Context, id string) (*system_db.Role, error) {
	role := new(system_db.Role)
	err := r.db.NewSelect().
		Model(role).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return role, nil
}

// 获取角色列表(分页)
func (r *roleRepository) FindPage(ctx context.Context, offset, limit int) ([]*system_db.Role, int, error) {
	var roles []*system_db.Role

	query := r.db.NewSelect().
		Model(&roles).
		Order("created_at DESC")

	// 获取总数
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	err = query.Offset(offset).Limit(limit).Scan(ctx)
	if err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}

// 获取角色列表(列表)
func (r *roleRepository) FindList(ctx context.Context) ([]*system_db.Role, error) {
	var roles []*system_db.Role
	err := r.db.NewSelect().
		Model(&roles).
		Order("created_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return roles, nil
}

// 创建角色
func (r *roleRepository) Create(ctx context.Context, role *system_db.Role) error {
	role.ID = strings.ReplaceAll(uuid.New().String(), "-", "")

	_, err := r.db.NewInsert().Model(role).Exec(ctx)
	return err
}

// 更新角色
func (r *roleRepository) Update(ctx context.Context, role *system_db.Role) error {
	_, err := r.db.NewUpdate().
		Model(role).
		Where("id = ?", role.ID).
		Exec(ctx)
	return err
}

// 删除角色
func (r *roleRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().
		Model((*system_db.Role)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	return err
}
