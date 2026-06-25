package repository

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	system_db "sag-wiki/app/system/models/db"
	"sag-wiki/common/option"
)

// 部门仓储接口
type DepartmentRepository interface {
	FindOne(ctx context.Context, id string) (*system_db.Department, error)
	FindPage(ctx context.Context, offset, limit int) ([]*system_db.Department, int, error)
	FindList(ctx context.Context) ([]*system_db.Department, error)
	FindTree(ctx context.Context) ([]*system_db.Department, error)
	Create(ctx context.Context, department *system_db.Department) error
	Update(ctx context.Context, department *system_db.Department) error
	Delete(ctx context.Context, id string) error
	FindChildren(ctx context.Context, parentID *string) ([]*system_db.Department, error)
	FindOptions(ctx context.Context) ([]*option.Option, error)
	SearchDepartments(ctx context.Context, query string) ([]*system_db.Department, error)
}

type departmentRepository struct {
	db *bun.DB
}

// 创建部门仓储实例
func NewDepartmentRepository(db *bun.DB) DepartmentRepository {
	return &departmentRepository{db: db}
}

// 根据ID获取部门
func (r *departmentRepository) FindOne(ctx context.Context, id string) (*system_db.Department, error) {
	department := new(system_db.Department)
	err := r.db.NewSelect().
		Model(department).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return department, nil
}

// 获取部门列表
func (r *departmentRepository) FindPage(ctx context.Context, offset, limit int) ([]*system_db.Department, int, error) {
	var departments []*system_db.Department

	query := r.db.NewSelect().
		Model(&departments).
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

	return departments, total, nil
}

func (r *departmentRepository) FindList(ctx context.Context) ([]*system_db.Department, error) {
	var departments []*system_db.Department

	err := r.db.NewSelect().
		Model(&departments).
		Order("created_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return departments, nil
}

// 创建部门
func (r *departmentRepository) Create(ctx context.Context, department *system_db.Department) error {
	department.ID = strings.ReplaceAll(uuid.New().String(), "-", "")

	_, err := r.db.NewInsert().Model(department).Exec(ctx)
	return err
}

// 更新部门
func (r *departmentRepository) Update(ctx context.Context, department *system_db.Department) error {
	_, err := r.db.NewUpdate().
		Model(department).
		Where("id = ?", department.ID).
		Exec(ctx)
	return err
}

// 删除部门
func (r *departmentRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().
		Model((*system_db.Department)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

// 根据父部门ID获取子部门列表
func (r *departmentRepository) FindChildren(ctx context.Context, parentID *string) ([]*system_db.Department, error) {
	var departments []*system_db.Department

	query := r.db.NewSelect().
		Model(&departments)

	if parentID == nil {
		query = query.Where("parent_id IS NULL")
	} else {
		query = query.Where("parent_id = ?", *parentID)
	}

	err := query.Order("created_at DESC").Scan(ctx)
	if err != nil {
		return nil, err
	}

	return departments, nil
}

// 获取部门选项
func (r *departmentRepository) FindOptions(ctx context.Context) ([]*option.Option, error) {
	var options []*option.Option

	err := r.db.NewSelect().
		Model((*system_db.Department)(nil)).
		Column("id", "name", "description").
		Order("created_at DESC").
		Scan(ctx, &options)
	if err != nil {
		return nil, err
	}

	return options, nil
}

// 获取树形结构的部门列表
func (r *departmentRepository) FindTree(ctx context.Context) ([]*system_db.Department, error) {
	var departments []*system_db.Department

	// 获取所有部门
	err := r.db.NewSelect().
		Model(&departments).
		Order("created_at ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	// 构建树形结构
	return buildDepartmentTree(departments, nil), nil
}

// buildDepartmentTree 递归构建部门树
func buildDepartmentTree(departments []*system_db.Department, parentID *string) []*system_db.Department {
	var tree []*system_db.Department

	for _, dept := range departments {
		// 匹配父节点
		if (parentID == nil && dept.ParentID == nil) ||
			(parentID != nil && dept.ParentID != nil && *dept.ParentID == *parentID) {
			// 递归查找子节点
			dept.Children = buildDepartmentTree(departments, &dept.ID)
			tree = append(tree, dept)
		}
	}

	return tree
}

// 搜索部门
func (r *departmentRepository) SearchDepartments(ctx context.Context, query string) ([]*system_db.Department, error) {
	var departments []*system_db.Department

	err := r.db.NewSelect().
		Model(&departments).
		Where("name LIKE ?", "%"+query+"%").
		Order("created_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return departments, nil
}
