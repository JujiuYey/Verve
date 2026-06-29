package handlers

import (
	"github.com/gofiber/fiber/v2"

	learning_db "sag-wiki/app/learning/models/db"
	learning_payload "sag-wiki/app/learning/models/payload"
	learning_service "sag-wiki/app/learning/service"
	"sag-wiki/common/pagination"
	"sag-wiki/common/response"
	"sag-wiki/infrastructure/database"
)

// 学习目标处理器
type GoalHandler struct {
	db      *database.DatabaseService
	planner *learning_service.PlannerService
}

func NewGoalHandler(db *database.DatabaseService) *GoalHandler {
	return &GoalHandler{
		db:      db,
		planner: learning_service.NewPlannerService(db),
	}
}

// 创建学习目标(一句话)+ 生成学习路线
func (h *GoalHandler) Create(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	if userID == "" {
		return response.UnauthorizedCtx(c, "未授权")
	}

	var req learning_payload.CreateGoalRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequestCtx(c, err.Error())
	}
	if req.Title == "" {
		return response.BadRequestCtx(c, "学习目标不能为空")
	}

	goal := &learning_db.LearningGoal{
		UserID: userID,
		Title:  req.Title,
		Source: "text",
		Status: "active",
	}
	if err := h.db.Goals.Create(c.Context(), goal); err != nil {
		return response.InternalServerCtx(c, "创建学习目标失败")
	}

	// 生成学习路线(Planner)。失败则回收目标,避免半成品。
	if err := h.planner.GeneratePath(c.Context(), goal); err != nil {
		_ = h.db.Goals.Delete(c.Context(), goal.ID)
		return response.InternalServerCtx(c, "学习路线生成失败,请重试")
	}

	return response.SuccessCtx(c, fiber.Map{"goal_id": goal.ID})
}

// 学习目标分页列表(仅本人)
func (h *GoalHandler) FindPage(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)

	var req pagination.PaginationRequest
	if err := c.QueryParser(&req); err != nil {
		return response.BadRequestCtx(c)
	}
	req.Validate()

	goals, total, err := h.db.Goals.FindByUser(c.Context(), userID, req.GetOffset(), req.PageSize)
	if err != nil {
		return response.InternalServerCtx(c, "获取学习目标失败")
	}
	return response.PaginateCtx(c, goals, total, req.Page, req.PageSize)
}

// 学习目标详情(含路线 + 进度)
func (h *GoalHandler) FindOne(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	id := c.Params("id")

	goal, err := h.db.Goals.FindOne(c.Context(), id)
	if err != nil {
		return response.NotFoundCtx(c, "学习目标不存在")
	}
	if goal.UserID != userID {
		return response.ForbiddenCtx(c, "无权访问")
	}

	resp := fiber.Map{"goal": goal}

	// 路线 + 小目标 + 进度(可能尚未生成)
	if path, err := h.db.Paths.FindByGoal(c.Context(), id); err == nil && path != nil {
		objectives, _ := h.db.Objectives.FindByPath(c.Context(), path.ID)
		completed, total, _ := h.db.Objectives.CountByPath(c.Context(), path.ID)
		resp["path"] = path
		resp["objectives"] = objectives
		resp["current_objective_id"] = path.CurrentObjectiveID
		resp["progress"] = fiber.Map{"completed": completed, "total": total}
	}

	return response.SuccessCtx(c, resp)
}

// 更新学习目标(标题 / 状态)
func (h *GoalHandler) Update(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)

	var req learning_payload.UpdateGoalRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequestCtx(c, err.Error())
	}

	goal, err := h.db.Goals.FindOne(c.Context(), req.ID)
	if err != nil {
		return response.NotFoundCtx(c, "学习目标不存在")
	}
	if goal.UserID != userID {
		return response.ForbiddenCtx(c, "无权访问")
	}

	if req.Title != nil {
		goal.Title = *req.Title
	}
	if req.Status != nil {
		goal.Status = *req.Status
	}
	if err := h.db.Goals.Update(c.Context(), goal); err != nil {
		return response.InternalServerCtx(c, "更新失败")
	}
	return response.SuccessMsgCtx(c, "更新成功")
}

// 删除学习目标
func (h *GoalHandler) Delete(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	id := c.Params("id")

	goal, err := h.db.Goals.FindOne(c.Context(), id)
	if err != nil {
		return response.NotFoundCtx(c, "学习目标不存在")
	}
	if goal.UserID != userID {
		return response.ForbiddenCtx(c, "无权访问")
	}

	if err := h.db.Goals.Delete(c.Context(), id); err != nil {
		return response.InternalServerCtx(c, "删除失败")
	}
	return response.SuccessMsgCtx(c, "删除成功")
}

// 继续上次 / 今日推荐
func (h *GoalHandler) Continue(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)

	goals, _, err := h.db.Goals.FindByUser(c.Context(), userID, 0, 1)
	if err != nil || len(goals) == 0 {
		return response.SuccessCtx[any](c, nil)
	}
	goal := goals[0]

	path, err := h.db.Paths.FindByGoal(c.Context(), goal.ID)
	if err != nil || path.CurrentObjectiveID == nil {
		return response.SuccessCtx(c, fiber.Map{"goal_id": goal.ID})
	}

	resp := fiber.Map{"goal_id": goal.ID, "objective_id": *path.CurrentObjectiveID}
	if obj, err := h.db.Objectives.FindOne(c.Context(), *path.CurrentObjectiveID); err == nil {
		resp["title"] = obj.Title
	}
	return response.SuccessCtx(c, resp)
}
