package handlers

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/gofiber/fiber/v2"

	learning_db "sag-wiki/app/learning/models/db"
	learning_payload "sag-wiki/app/learning/models/payload"
	learning_service "sag-wiki/app/learning/service"
	wiki_db "sag-wiki/app/wiki/models/db"
	wiki_repo "sag-wiki/app/wiki/repository"
	"sag-wiki/common/pagination"
	"sag-wiki/common/response"
	"sag-wiki/infrastructure/database"
)

// 学习目标处理器
type GoalHandler struct {
	db        *database.DatabaseService
	planner   *learning_service.PlannerService
	folders   wiki_repo.FolderRepository
	documents *wiki_repo.DocumentRepository
}

func NewGoalHandler(db *database.DatabaseService) *GoalHandler {
	bunDB := db.GetDB()
	return &GoalHandler{
		db:        db,
		planner:   learning_service.NewPlannerService(db),
		folders:   wiki_repo.NewFolderRepository(bunDB),
		documents: wiki_repo.NewDocumentRepository(bunDB),
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

// 从文档管理文件夹创建学习目标，并生成学习路线
func (h *GoalHandler) CreateFromFolder(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	if userID == "" {
		return response.UnauthorizedCtx(c, "未授权")
	}

	var req learning_payload.CreateGoalFromFolderRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequestCtx(c, err.Error())
	}
	if req.FolderID == "" {
		return response.BadRequestCtx(c, "文件夹不能为空")
	}
	log.Printf("📚 从文件夹生成学习路径: user_id=%s folder_id=%s", userID, req.FolderID)

	folder, err := h.folders.FindOne(c.Context(), req.FolderID)
	if err != nil {
		log.Printf("❌ 查找学习路径来源文件夹失败: folder_id=%s err=%v", req.FolderID, err)
		return response.NotFoundCtx(c, "文件夹不存在")
	}
	log.Printf("📁 学习路径来源文件夹: id=%s name=%s", folder.ID, folder.Name)

	folderIDs, err := h.folders.GetAllSubFolderIDs(c.Context(), req.FolderID)
	if err != nil {
		log.Printf("❌ 读取学习路径来源文件夹树失败: folder_id=%s err=%v", req.FolderID, err)
		return response.InternalServerCtx(c, "读取文件夹结构失败")
	}

	allFolders, err := h.folders.GetAll(c.Context())
	if err != nil {
		log.Printf("❌ 读取全部文件夹失败: folder_id=%s err=%v", req.FolderID, err)
		return response.InternalServerCtx(c, "读取文件夹结构失败")
	}
	selectedFolders := filterFoldersByID(allFolders, folderIDs)

	docs, err := h.documents.GetDocumentsByFolderIDs(c.Context(), folderIDs)
	if err != nil {
		log.Printf("❌ 读取学习路径来源文档失败: folder_id=%s folder_count=%d err=%v", req.FolderID, len(folderIDs), err)
		return response.InternalServerCtx(c, "读取文档列表失败")
	}
	log.Printf("📄 学习路径来源统计: root_folder=%s folder_count=%d selected_folder_count=%d document_count=%d", folder.Name, len(folderIDs), len(selectedFolders), len(docs))
	if len(folderIDs) <= 1 && len(docs) == 0 {
		log.Printf("⚠️ 当前文件夹没有可生成学习路径的内容: folder_id=%s", req.FolderID)
		return response.BadRequestCtx(c, "当前文件夹没有可用于生成学习路径的内容")
	}

	title := fmt.Sprintf("学习 %s", folder.Name)
	sourcePrompt := buildFolderLearningPrompt(folder, selectedFolders, docs)
	log.Printf("🧭 学习路径生成输入: title=%s prompt_chars=%d prompt_preview=%q", title, len(sourcePrompt), truncateForLog(sourcePrompt, 600))
	goal := &learning_db.LearningGoal{
		UserID:      userID,
		Title:       title,
		Description: &sourcePrompt,
		Source:      "documents",
		Status:      "active",
	}
	if err := h.db.Goals.Create(c.Context(), goal); err != nil {
		log.Printf("❌ 创建文档来源学习目标失败: title=%s folder_id=%s err=%v", title, req.FolderID, err)
		return response.InternalServerCtx(c, "创建学习目标失败")
	}
	log.Printf("✅ 文档来源学习目标已创建: goal_id=%s source_folder_id=%s", goal.ID, req.FolderID)

	if err := h.planner.GeneratePathWithQuery(c.Context(), goal, sourcePrompt); err != nil {
		_ = h.db.Goals.Delete(c.Context(), goal.ID)
		log.Printf("❌ 文档来源学习路径生成失败，已回收目标: goal_id=%s folder_id=%s err=%v", goal.ID, req.FolderID, err)
		return response.InternalServerCtx(c, "学习路线生成失败,请重试")
	}
	log.Printf("✅ 文档来源学习路径生成成功: goal_id=%s folder_id=%s", goal.ID, req.FolderID)

	return response.SuccessCtx(c, fiber.Map{"goal_id": goal.ID})
}

func filterFoldersByID(folders []*wiki_db.Folder, folderIDs []string) []*wiki_db.Folder {
	idSet := make(map[string]struct{}, len(folderIDs))
	for _, id := range folderIDs {
		idSet[id] = struct{}{}
	}

	selected := make([]*wiki_db.Folder, 0, len(folderIDs))
	for _, folder := range folders {
		if _, ok := idSet[folder.ID]; ok {
			selected = append(selected, folder)
		}
	}
	return selected
}

func buildFolderLearningPrompt(root *wiki_db.Folder, folders []*wiki_db.Folder, docs []*wiki_db.Document) string {
	var sb strings.Builder
	sb.WriteString("请根据文档管理中的资料结构，为学习者生成一条由浅入深的学习路径。\n")
	sb.WriteString("资料根目录: ")
	sb.WriteString(root.Name)
	sb.WriteString("\n")
	if root.Description != nil && strings.TrimSpace(*root.Description) != "" {
		sb.WriteString("目录说明: ")
		sb.WriteString(strings.TrimSpace(*root.Description))
		sb.WriteString("\n")
	}

	docsByFolder := make(map[string][]*wiki_db.Document)
	for _, doc := range docs {
		docsByFolder[doc.FolderID] = append(docsByFolder[doc.FolderID], doc)
	}
	for folderID := range docsByFolder {
		sort.Slice(docsByFolder[folderID], func(i, j int) bool {
			return docsByFolder[folderID][i].Filename < docsByFolder[folderID][j].Filename
		})
	}
	childrenByParent := make(map[string][]*wiki_db.Folder)
	for _, folder := range folders {
		if folder.ID == root.ID || folder.ParentID == nil {
			continue
		}
		childrenByParent[*folder.ParentID] = append(childrenByParent[*folder.ParentID], folder)
	}
	for parentID := range childrenByParent {
		sort.Slice(childrenByParent[parentID], func(i, j int) bool {
			return childrenByParent[parentID][i].Name < childrenByParent[parentID][j].Name
		})
	}

	sb.WriteString("资料目录结构:\n")
	writeFolderLearningPromptNode(&sb, root, childrenByParent, docsByFolder, 0)
	sb.WriteString("\n生成要求: 优先尊重目录和文件名中已有的编号顺序；把每个阶段拆成可练习、可验证的小目标；不要生成和资料主题无关的内容。\n")

	return sb.String()
}

func writeFolderLearningPromptNode(sb *strings.Builder, folder *wiki_db.Folder, childrenByParent map[string][]*wiki_db.Folder, docsByFolder map[string][]*wiki_db.Document, depth int) {
	indent := strings.Repeat("  ", depth)
	sb.WriteString(fmt.Sprintf("%s- 文件夹: %s\n", indent, folder.Name))
	for _, doc := range docsByFolder[folder.ID] {
		sb.WriteString(fmt.Sprintf("%s  - 文档: %s\n", indent, doc.Filename))
	}
	for _, child := range childrenByParent[folder.ID] {
		writeFolderLearningPromptNode(sb, child, childrenByParent, docsByFolder, depth+1)
	}
}

func truncateForLog(text string, limit int) string {
	text = strings.TrimSpace(text)
	if len(text) <= limit {
		return text
	}
	return text[:limit] + "...(truncated)"
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
