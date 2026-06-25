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

// 获取用户的角色列表
func (r *UserRepository) GetUserRoles(ctx context.Context, userID string) ([]string, error) {
	var roles []string
	err := r.db.NewSelect().
		Model((*system_db.UserRole)(nil)).
		Column("r.name").
		Join("JOIN sys_roles AS r ON r.id = sur.role_id").
		Where("sur.user_id = ?", userID).
		Scan(ctx, &roles)
	if err != nil {
		return nil, err
	}
	return roles, nil
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

	// 如果有搜索条件
	if search != "" {
		query = query.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.WhereOr("username LIKE ?", "%"+search+"%").
				WhereOr("email LIKE ?", "%"+search+"%").
				WhereOr("full_name LIKE ?", "%"+search+"%")
		})
	}

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

// 设置默认部门
func (r *UserRepository) SetDefaultDepartment(ctx context.Context, userID string, departmentID string) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// 先将该用户的所有部门的 is_primary 设置为 false
		_, err := tx.NewUpdate().
			Model((*system_db.UserDepartment)(nil)).
			Set("is_primary = ?", false).
			Where("user_id = ?", userID).
			Exec(ctx)
		if err != nil {
			return err
		}

		// 尝试将指定部门的 is_primary 设置为 true
		result, err := tx.NewUpdate().
			Model((*system_db.UserDepartment)(nil)).
			Set("is_primary = ?", true).
			Where("user_id = ?", userID).
			Where("department_id = ?", departmentID).
			Exec(ctx)
		if err != nil {
			return err
		}

		// 检查是否更新了记录
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}

		// 如果没有更新记录，说明该用户和该部门的关联不存在，需要插入新记录
		if rowsAffected == 0 {
			userDepartment := &system_db.UserDepartment{
				ID:           strings.ReplaceAll(uuid.New().String(), "-", ""),
				UserID:       userID,
				DepartmentID: departmentID,
				IsPrimary:    true,
			}
			_, err = tx.NewInsert().Model(userDepartment).Exec(ctx)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// 创建用户部门关联
func (r *UserRepository) CreateUserDepartments(ctx context.Context, userID string, departmentIDs []string, primaryDepartmentID string) error {
	if len(departmentIDs) == 0 {
		return nil
	}

	userDepartments := make([]*system_db.UserDepartment, len(departmentIDs))
	for i, deptID := range departmentIDs {
		userDepartments[i] = &system_db.UserDepartment{
			ID:           strings.ReplaceAll(uuid.New().String(), "-", ""),
			UserID:       userID,
			DepartmentID: deptID,
			IsPrimary:    deptID == primaryDepartmentID,
		}
	}

	_, err := r.db.NewInsert().Model(&userDepartments).Exec(ctx)
	return err
}

// 创建用户角色关联
func (r *UserRepository) CreateUserRoles(ctx context.Context, userID string, roleIDs []string) error {
	if len(roleIDs) == 0 {
		return nil
	}

	userRoles := make([]*system_db.UserRole, len(roleIDs))
	for i, roleID := range roleIDs {
		userRoles[i] = &system_db.UserRole{
			ID:     strings.ReplaceAll(uuid.New().String(), "-", ""),
			UserID: userID,
			RoleID: roleID,
		}
	}

	_, err := r.db.NewInsert().Model(&userRoles).Exec(ctx)
	return err
}

// 删除用户的所有部门关联
func (r *UserRepository) DeleteUserDepartments(ctx context.Context, userID string) error {
	_, err := r.db.NewDelete().
		Model((*system_db.UserDepartment)(nil)).
		Where("user_id = ?", userID).
		Exec(ctx)
	return err
}

// 删除用户的所有角色关联
func (r *UserRepository) DeleteUserRoles(ctx context.Context, userID string) error {
	_, err := r.db.NewDelete().
		Model((*system_db.UserRole)(nil)).
		Where("user_id = ?", userID).
		Exec(ctx)
	return err
}

// 获取用户的部门列表
func (r *UserRepository) GetUserDepartments(ctx context.Context, userID string) ([]*system_db.Department, string, error) {
	var departments []*system_db.Department
	var primaryDepartmentID string

	// 查询用户的部门
	err := r.db.NewSelect().
		Model(&departments).
		Join("JOIN sys_user_departments AS sud ON sud.department_id = sd.id").
		Where("sud.user_id = ?", userID).
		Scan(ctx)
	if err != nil {
		return nil, "", err
	}

	// 查询主部门ID
	err = r.db.NewSelect().
		Model((*system_db.UserDepartment)(nil)).
		Column("department_id").
		Where("user_id = ?", userID).
		Where("is_primary = ?", true).
		Scan(ctx, &primaryDepartmentID)
	if err != nil && err.Error() != "sql: no rows in result set" {
		return nil, "", err
	}

	return departments, primaryDepartmentID, nil
}

// 获取用户的角色列表
func (r *UserRepository) GetUserRolesWithDetails(ctx context.Context, userID string) ([]*system_db.Role, error) {
	var roles []*system_db.Role

	err := r.db.NewSelect().
		Model(&roles).
		Join("JOIN sys_user_roles AS sur ON sur.role_id = sr.id").
		Where("sur.user_id = ?", userID).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return roles, nil
}

// 根据部门ID获取用户列表
func (r *UserRepository) FindByDepartmentID(ctx context.Context, departmentID string) ([]*system_db.User, error) {
	var users []*system_db.User

	err := r.db.NewSelect().
		Model(&users).
		Join("JOIN sys_user_departments AS sud ON sud.user_id = su.id").
		Where("sud.department_id = ?", departmentID).
		Order("su.created_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return users, nil
}

// 搜索用户（返回用户及其部门路径）
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
