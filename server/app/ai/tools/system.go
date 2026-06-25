package tools

import (
	"context"
	"log"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"

	"sag-wiki/app/system/models/db"
	"sag-wiki/app/system/repository"
)

// ==================== User Tools ====================

type SearchUserInput struct {
	Query  string `json:"query" jsonschema_description:"Search keyword for username, email or full name"`
	Offset int    `json:"offset" jsonschema_description:"Pagination offset, default 0"`
	Limit  int    `json:"limit" jsonschema_description:"Pagination limit, default 10"`
}

type SearchUserOutput struct {
	Users []*db.User `json:"users"`
	Total int        `json:"total"`
}

type GetUserInput struct {
	UserID string `json:"user_id" jsonschema_description:"User ID"`
}

type GetUserOutput struct {
	User *db.User `json:"user"`
}

type CreateUserInput struct {
	Username string  `json:"username" jsonschema_description:"Username (required)"`
	Email    string  `json:"email" jsonschema_description:"Email (required)"`
	Password string  `json:"password" jsonschema_description:"Password (required)"`
	FullName *string `json:"full_name" jsonschema_description:"Full name"`
}

type CreateUserOutput struct {
	User *db.User `json:"user"`
}

type UpdateUserInput struct {
	ID       string  `json:"id" jsonschema_description:"User ID (required)"`
	Username *string `json:"username" jsonschema_description:"Username"`
	Email    *string `json:"email" jsonschema_description:"Email"`
	FullName *string `json:"full_name" jsonschema_description:"Full name"`
	Status   *string `json:"status" jsonschema_description:"Status: active or inactive"`
}

type UpdateUserOutput struct {
	User *db.User `json:"user"`
}

type DeleteUserInput struct {
	UserID string `json:"user_id" jsonschema_description:"User ID to delete"`
}

type DeleteUserOutput struct {
	Success bool `json:"success"`
}

// ==================== Department Tools ====================

type ListDepartmentInput struct {
	Offset int `json:"offset" jsonschema_description:"Pagination offset, default 0"`
	Limit  int `json:"limit" jsonschema_description:"Pagination limit, default 10"`
}

type ListDepartmentOutput struct {
	Departments []*db.Department `json:"departments"`
	Total       int               `json:"total"`
}

type GetDepartmentInput struct {
	DepartmentID string `json:"department_id" jsonschema_description:"Department ID"`
}

type GetDepartmentOutput struct {
	Department *db.Department `json:"department"`
}

type CreateDepartmentInput struct {
	Name        string  `json:"name" jsonschema_description:"Department name (required)"`
	Description *string `json:"description" jsonschema_description:"Department description"`
	ParentID    *string `json:"parent_id" jsonschema_description:"Parent department ID for hierarchy"`
}

type CreateDepartmentOutput struct {
	Department *db.Department `json:"department"`
}

type UpdateDepartmentInput struct {
	ID          string  `json:"id" jsonschema_description:"Department ID (required)"`
	Name        *string `json:"name" jsonschema_description:"Department name"`
	Description *string `json:"description" jsonschema_description:"Department description"`
	ParentID    *string `json:"parent_id" jsonschema_description:"Parent department ID"`
}

type UpdateDepartmentOutput struct {
	Department *db.Department `json:"department"`
}

type DeleteDepartmentInput struct {
	DepartmentID string `json:"department_id" jsonschema_description:"Department ID to delete"`
}

type DeleteDepartmentOutput struct {
	Success bool `json:"success"`
}

// ==================== Role Tools ====================

type ListRoleInput struct {
	Offset int `json:"offset" jsonschema_description:"Pagination offset, default 0"`
	Limit  int `json:"limit" jsonschema_description:"Pagination limit, default 10"`
}

type ListRoleOutput struct {
	Roles []*db.Role `json:"roles"`
	Total int        `json:"total"`
}

type GetRoleInput struct {
	RoleID string `json:"role_id" jsonschema_description:"Role ID"`
}

type GetRoleOutput struct {
	Role *db.Role `json:"role"`
}

type CreateRoleInput struct {
	Name        string  `json:"name" jsonschema_description:"Role name (required, must be unique)"`
	Description *string `json:"description" jsonschema_description:"Role description"`
}

type CreateRoleOutput struct {
	Role *db.Role `json:"role"`
}

type UpdateRoleInput struct {
	ID          string  `json:"id" jsonschema_description:"Role ID (required)"`
	Name        *string `json:"name" jsonschema_description:"Role name (must be unique)"`
	Description *string `json:"description" jsonschema_description:"Role description"`
}

type UpdateRoleOutput struct {
	Role *db.Role `json:"role"`
}

type DeleteRoleInput struct {
	RoleID string `json:"role_id" jsonschema_description:"Role ID to delete"`
}

type DeleteRoleOutput struct {
	Success bool `json:"success"`
}

// ==================== Tool Factories ====================

// NewSystemTools creates all system management tools
func NewSystemTools(userRepo *repository.UserRepository, deptRepo repository.DepartmentRepository, roleRepo repository.RoleRepository) []tool.BaseTool {
	return []tool.BaseTool{
		// User tools
		newSearchUserTool(userRepo),
		newGetUserTool(userRepo),
		newCreateUserTool(userRepo),
		newUpdateUserTool(userRepo),
		newDeleteUserTool(userRepo),
		// Department tools
		newListDepartmentTool(deptRepo),
		newGetDepartmentTool(deptRepo),
		newCreateDepartmentTool(deptRepo),
		newUpdateDepartmentTool(deptRepo),
		newDeleteDepartmentTool(deptRepo),
		// Role tools
		newListRoleTool(roleRepo),
		newGetRoleTool(roleRepo),
		newCreateRoleTool(roleRepo),
		newUpdateRoleTool(roleRepo),
		newDeleteRoleTool(roleRepo),
	}
}

// User tools
func newSearchUserTool(repo *repository.UserRepository) tool.InvokableTool {
	t, err := utils.InferTool("search_user", "Search users by keyword (username, email or full name)",
		func(ctx context.Context, input *SearchUserInput) (output *SearchUserOutput, err error) {
			offset := input.Offset
			limit := input.Limit
			if limit <= 0 {
				limit = 10
			}
			users, total, err := repo.FindPage(ctx, offset, limit, input.Query)
			return &SearchUserOutput{Users: users, Total: total}, err
		})
	if err != nil {
		log.Fatal(err)
	}
	return t
}

func newGetUserTool(repo *repository.UserRepository) tool.InvokableTool {
	t, err := utils.InferTool("get_user", "Get user details by user ID",
		func(ctx context.Context, input *GetUserInput) (output *GetUserOutput, err error) {
			user, err := repo.FindOne(ctx, input.UserID)
			return &GetUserOutput{User: user}, err
		})
	if err != nil {
		log.Fatal(err)
	}
	return t
}

func newCreateUserTool(repo *repository.UserRepository) tool.InvokableTool {
	t, err := utils.InferTool("create_user", "Create a new user",
		func(ctx context.Context, input *CreateUserInput) (output *CreateUserOutput, err error) {
			user := &db.User{
				Username: input.Username,
				Email:    input.Email,
				Password: input.Password,
				FullName: input.FullName,
			}
			err = repo.Create(ctx, user)
			return &CreateUserOutput{User: user}, err
		})
	if err != nil {
		log.Fatal(err)
	}
	return t
}

func newUpdateUserTool(repo *repository.UserRepository) tool.InvokableTool {
	t, err := utils.InferTool("update_user", "Update an existing user",
		func(ctx context.Context, input *UpdateUserInput) (output *UpdateUserOutput, err error) {
			user, err := repo.FindOne(ctx, input.ID)
			if err != nil {
				return nil, err
			}
			if input.Username != nil {
				user.Username = *input.Username
			}
			if input.Email != nil {
				user.Email = *input.Email
			}
			if input.FullName != nil {
				user.FullName = input.FullName
			}
			if input.Status != nil {
				user.Status = *input.Status
			}
			err = repo.Update(ctx, user)
			return &UpdateUserOutput{User: user}, err
		})
	if err != nil {
		log.Fatal(err)
	}
	return t
}

func newDeleteUserTool(repo *repository.UserRepository) tool.InvokableTool {
	t, err := utils.InferTool("delete_user", "Delete a user by ID",
		func(ctx context.Context, input *DeleteUserInput) (output *DeleteUserOutput, err error) {
			err = repo.Delete(ctx, input.UserID)
			return &DeleteUserOutput{Success: err == nil}, err
		})
	if err != nil {
		log.Fatal(err)
	}
	return t
}

// Department tools
func newListDepartmentTool(repo repository.DepartmentRepository) tool.InvokableTool {
	t, err := utils.InferTool("list_departments", "List departments with pagination",
		func(ctx context.Context, input *ListDepartmentInput) (output *ListDepartmentOutput, err error) {
			offset := input.Offset
			limit := input.Limit
			if limit <= 0 {
				limit = 10
			}
			depts, total, err := repo.FindPage(ctx, offset, limit)
			return &ListDepartmentOutput{Departments: depts, Total: total}, err
		})
	if err != nil {
		log.Fatal(err)
	}
	return t
}

func newGetDepartmentTool(repo repository.DepartmentRepository) tool.InvokableTool {
	t, err := utils.InferTool("get_department", "Get department details by ID",
		func(ctx context.Context, input *GetDepartmentInput) (output *GetDepartmentOutput, err error) {
			dept, err := repo.FindOne(ctx, input.DepartmentID)
			return &GetDepartmentOutput{Department: dept}, err
		})
	if err != nil {
		log.Fatal(err)
	}
	return t
}

func newCreateDepartmentTool(repo repository.DepartmentRepository) tool.InvokableTool {
	t, err := utils.InferTool("create_department", "Create a new department",
		func(ctx context.Context, input *CreateDepartmentInput) (output *CreateDepartmentOutput, err error) {
			dept := &db.Department{
				Name:        input.Name,
				Description: input.Description,
				ParentID:    input.ParentID,
			}
			err = repo.Create(ctx, dept)
			return &CreateDepartmentOutput{Department: dept}, err
		})
	if err != nil {
		log.Fatal(err)
	}
	return t
}

func newUpdateDepartmentTool(repo repository.DepartmentRepository) tool.InvokableTool {
	t, err := utils.InferTool("update_department", "Update an existing department",
		func(ctx context.Context, input *UpdateDepartmentInput) (output *UpdateDepartmentOutput, err error) {
			dept, err := repo.FindOne(ctx, input.ID)
			if err != nil {
				return nil, err
			}
			if input.Name != nil {
				dept.Name = *input.Name
			}
			if input.Description != nil {
				dept.Description = input.Description
			}
			if input.ParentID != nil {
				dept.ParentID = input.ParentID
			}
			err = repo.Update(ctx, dept)
			return &UpdateDepartmentOutput{Department: dept}, err
		})
	if err != nil {
		log.Fatal(err)
	}
	return t
}

func newDeleteDepartmentTool(repo repository.DepartmentRepository) tool.InvokableTool {
	t, err := utils.InferTool("delete_department", "Delete a department by ID",
		func(ctx context.Context, input *DeleteDepartmentInput) (output *DeleteDepartmentOutput, err error) {
			err = repo.Delete(ctx, input.DepartmentID)
			return &DeleteDepartmentOutput{Success: err == nil}, err
		})
	if err != nil {
		log.Fatal(err)
	}
	return t
}

// Role tools
func newListRoleTool(repo repository.RoleRepository) tool.InvokableTool {
	t, err := utils.InferTool("list_roles", "List roles with pagination",
		func(ctx context.Context, input *ListRoleInput) (output *ListRoleOutput, err error) {
			offset := input.Offset
			limit := input.Limit
			if limit <= 0 {
				limit = 10
			}
			roles, total, err := repo.FindPage(ctx, offset, limit)
			return &ListRoleOutput{Roles: roles, Total: total}, err
		})
	if err != nil {
		log.Fatal(err)
	}
	return t
}

func newGetRoleTool(repo repository.RoleRepository) tool.InvokableTool {
	t, err := utils.InferTool("get_role", "Get role details by ID",
		func(ctx context.Context, input *GetRoleInput) (output *GetRoleOutput, err error) {
			role, err := repo.FindOne(ctx, input.RoleID)
			return &GetRoleOutput{Role: role}, err
		})
	if err != nil {
		log.Fatal(err)
	}
	return t
}

func newCreateRoleTool(repo repository.RoleRepository) tool.InvokableTool {
	t, err := utils.InferTool("create_role", "Create a new role",
		func(ctx context.Context, input *CreateRoleInput) (output *CreateRoleOutput, err error) {
			role := &db.Role{
				Name:        input.Name,
				Description: input.Description,
			}
			err = repo.Create(ctx, role)
			return &CreateRoleOutput{Role: role}, err
		})
	if err != nil {
		log.Fatal(err)
	}
	return t
}

func newUpdateRoleTool(repo repository.RoleRepository) tool.InvokableTool {
	t, err := utils.InferTool("update_role", "Update an existing role",
		func(ctx context.Context, input *UpdateRoleInput) (output *UpdateRoleOutput, err error) {
			role, err := repo.FindOne(ctx, input.ID)
			if err != nil {
				return nil, err
			}
			if input.Name != nil {
				role.Name = *input.Name
			}
			if input.Description != nil {
				role.Description = input.Description
			}
			err = repo.Update(ctx, role)
			return &UpdateRoleOutput{Role: role}, err
		})
	if err != nil {
		log.Fatal(err)
	}
	return t
}

func newDeleteRoleTool(repo repository.RoleRepository) tool.InvokableTool {
	t, err := utils.InferTool("delete_role", "Delete a role by ID",
		func(ctx context.Context, input *DeleteRoleInput) (output *DeleteRoleOutput, err error) {
			err = repo.Delete(ctx, input.RoleID)
			return &DeleteRoleOutput{Success: err == nil}, err
		})
	if err != nil {
		log.Fatal(err)
	}
	return t
}
