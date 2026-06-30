package handlers

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

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

	h.generatePathAsync(goal, goal.Title, "")

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
	if len(docs) == 0 {
		log.Printf("⚠️ 当前文件夹没有可生成学习路径的内容: folder_id=%s", req.FolderID)
		return response.BadRequestCtx(c, "当前文件夹没有可用于生成学习路径的内容")
	}

	title := fmt.Sprintf("学习 %s", folder.Name)
	description := fmt.Sprintf("基于文档管理中的「%s」目录结构生成。", folder.Name)
	goal := &learning_db.LearningGoal{
		UserID:      userID,
		Title:       title,
		Description: &description,
		Source:      "documents",
		Status:      "active",
	}
	if err := h.db.Goals.Create(c.Context(), goal); err != nil {
		log.Printf("❌ 创建文档来源学习目标失败: title=%s folder_id=%s err=%v", title, req.FolderID, err)
		return response.InternalServerCtx(c, "创建学习目标失败")
	}
	log.Printf("✅ 文档来源学习目标已创建: goal_id=%s source_folder_id=%s", goal.ID, req.FolderID)

	if err := h.createPathFromFolderStructure(c.Context(), goal, folder, selectedFolders, docs); err != nil {
		_ = h.db.Goals.Delete(c.Context(), goal.ID)
		log.Printf("❌ 根据文件夹结构创建学习路线失败，已回收目标: goal_id=%s folder_id=%s err=%v", goal.ID, req.FolderID, err)
		return response.InternalServerCtx(c, "学习路线生成失败,请重试")
	}
	log.Printf("✅ 根据文件夹结构创建学习路线成功: goal_id=%s folder_id=%s", goal.ID, req.FolderID)

	return response.SuccessCtx(c, fiber.Map{"goal_id": goal.ID})
}

type folderStage struct {
	Title      string
	Objectives []*learning_db.LearningObjective
}

func (h *GoalHandler) createPathFromFolderStructure(ctx context.Context, goal *learning_db.LearningGoal, root *wiki_db.Folder, folders []*wiki_db.Folder, docs []*wiki_db.Document) error {
	stages := buildStagesFromFolderStructure(goal, root, folders, docs)
	if len(stages) == 0 {
		return fmt.Errorf("文件夹中没有可生成学习路线的文档")
	}

	overview := make([]map[string]interface{}, 0, len(stages))
	for _, stage := range stages {
		overview = append(overview, map[string]interface{}{"title": stage.Title})
	}

	path := &learning_db.LearningPath{
		GoalID:   goal.ID,
		UserID:   goal.UserID,
		Overview: overview,
		Status:   "active",
	}
	if err := h.db.Paths.Create(ctx, path); err != nil {
		return err
	}

	objectives := make([]*learning_db.LearningObjective, 0)
	orderIndex := 0
	for _, stage := range stages {
		for _, objective := range stage.Objectives {
			objective.PathID = path.ID
			objective.OrderIndex = orderIndex
			if orderIndex == 0 {
				objective.Status = "active"
			}
			objectives = append(objectives, objective)
			orderIndex++
		}
	}
	if err := h.db.Objectives.BulkCreate(ctx, objectives); err != nil {
		return err
	}

	path.CurrentObjectiveID = &objectives[0].ID
	if err := h.db.Paths.Update(ctx, path); err != nil {
		return err
	}

	log.Printf("📌 文件夹结构学习路线已落库: goal_id=%s path_id=%s stage_count=%d objective_count=%d", goal.ID, path.ID, len(stages), len(objectives))
	return nil
}

func buildStagesFromFolderStructure(goal *learning_db.LearningGoal, root *wiki_db.Folder, folders []*wiki_db.Folder, docs []*wiki_db.Document) []folderStage {
	foldersByID := make(map[string]*wiki_db.Folder, len(folders))
	for _, folder := range folders {
		foldersByID[folder.ID] = folder
	}

	rootChildren := make([]*wiki_db.Folder, 0)
	for _, folder := range folders {
		if folder.ParentID != nil && *folder.ParentID == root.ID {
			rootChildren = append(rootChildren, folder)
		}
	}
	sort.Slice(rootChildren, func(i, j int) bool {
		return rootChildren[i].Name < rootChildren[j].Name
	})

	stageByFolderID := make(map[string]*folderStage, len(rootChildren))
	stages := make([]*folderStage, 0, len(rootChildren)+1)
	if hasRootDocuments(root, docs) {
		stages = append(stages, &folderStage{Title: "基础资料"})
	}
	for _, folder := range rootChildren {
		stage := &folderStage{Title: folder.Name}
		stages = append(stages, stage)
		stageByFolderID[folder.ID] = stage
	}

	docs = append([]*wiki_db.Document(nil), docs...)
	sort.Slice(docs, func(i, j int) bool {
		leftStage := stageSortKey(root, foldersByID, docs[i])
		rightStage := stageSortKey(root, foldersByID, docs[j])
		if leftStage != rightStage {
			return leftStage < rightStage
		}
		return docs[i].Filename < docs[j].Filename
	})

	for _, doc := range docs {
		stageFolder := findTopLevelFolder(root, foldersByID, doc.FolderID)
		stageTitle := "基础资料"
		if stageFolder != nil {
			stageTitle = stageFolder.Name
		}

		stage := findOrCreateStage(&stages, stageByFolderID, stageFolder, stageTitle)
		folderPath := folderPathName(root, foldersByID, doc.FolderID)
		detail := fmt.Sprintf("来源文档: %s", doc.Filename)
		if folderPath != "" {
			detail = fmt.Sprintf("来源目录: %s\n来源文档: %s", folderPath, doc.Filename)
		}
		stage.Objectives = append(stage.Objectives, &learning_db.LearningObjective{
			UserID:           goal.UserID,
			StageTitle:       &stage.Title,
			Title:            cleanLearningTitle(doc.Filename),
			Detail:           &detail,
			SourceDocumentID: &doc.ID,
			SourceFolderID:   &doc.FolderID,
			SourceFolderPath: &folderPath,
			Status:           "pending",
			MasteryLevel:     "none",
		})
	}

	result := make([]folderStage, 0, len(stages))
	for _, stage := range stages {
		if len(stage.Objectives) > 0 {
			result = append(result, *stage)
		}
	}
	return result
}

func hasRootDocuments(root *wiki_db.Folder, docs []*wiki_db.Document) bool {
	for _, doc := range docs {
		if doc.FolderID == root.ID {
			return true
		}
	}
	return false
}

func findOrCreateStage(stages *[]*folderStage, stageByFolderID map[string]*folderStage, stageFolder *wiki_db.Folder, title string) *folderStage {
	if stageFolder != nil {
		if stage, ok := stageByFolderID[stageFolder.ID]; ok {
			return stage
		}
	}
	for i := range *stages {
		if (*stages)[i].Title == title {
			return (*stages)[i]
		}
	}
	stage := &folderStage{Title: title}
	*stages = append(*stages, stage)
	return stage
}

func findTopLevelFolder(root *wiki_db.Folder, foldersByID map[string]*wiki_db.Folder, folderID string) *wiki_db.Folder {
	current := foldersByID[folderID]
	for current != nil && current.ParentID != nil {
		if *current.ParentID == root.ID {
			return current
		}
		current = foldersByID[*current.ParentID]
	}
	return nil
}

func stageSortKey(root *wiki_db.Folder, foldersByID map[string]*wiki_db.Folder, doc *wiki_db.Document) string {
	stage := findTopLevelFolder(root, foldersByID, doc.FolderID)
	if stage == nil {
		return ""
	}
	return stage.Name
}

func folderPathName(root *wiki_db.Folder, foldersByID map[string]*wiki_db.Folder, folderID string) string {
	names := make([]string, 0)
	for current := foldersByID[folderID]; current != nil; {
		names = append(names, current.Name)
		if current.ID == root.ID || current.ParentID == nil {
			break
		}
		current = foldersByID[*current.ParentID]
	}
	for i, j := 0, len(names)-1; i < j; i, j = i+1, j-1 {
		names[i], names[j] = names[j], names[i]
	}
	return strings.Join(names, " / ")
}

var titleNumberPrefixPattern = regexp.MustCompile(`^\d+[-_、.．\s]*`)

func cleanLearningTitle(filename string) string {
	title := strings.TrimSuffix(filename, filepath.Ext(filename))
	title = titleNumberPrefixPattern.ReplaceAllString(title, "")
	title = strings.TrimSpace(title)
	if title == "" {
		return filename
	}
	return title
}

func (h *GoalHandler) generatePathAsync(goal *learning_db.LearningGoal, query string, sourceFolderID string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		log.Printf("🚀 学习路径异步生成任务开始: goal_id=%s source=%s source_folder_id=%s", goal.ID, goal.Source, sourceFolderID)
		if err := h.planner.GeneratePathWithQuery(ctx, goal, query); err != nil {
			log.Printf("❌ 学习路径异步生成任务失败: goal_id=%s source=%s source_folder_id=%s err=%v", goal.ID, goal.Source, sourceFolderID, err)
			return
		}
		log.Printf("✅ 学习路径异步生成任务完成: goal_id=%s source=%s source_folder_id=%s", goal.ID, goal.Source, sourceFolderID)
	}()
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

// 学习目标分页列表(仅本人)
func (h *GoalHandler) FindPage(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	if userID == "" {
		log.Printf("⚠️ 学习目标分页查询缺少用户信息: path=%s", c.Path())
		return response.UnauthorizedCtx(c, "未授权")
	}

	var req pagination.PaginationRequest
	if err := c.QueryParser(&req); err != nil {
		return response.BadRequestCtx(c)
	}
	req.Validate()

	goals, total, err := h.db.Goals.FindByUser(c.Context(), userID, req.GetOffset(), req.PageSize)
	if err != nil {
		log.Printf("❌ 学习目标分页查询失败: user_id=%s page=%d page_size=%d err=%v", userID, req.Page, req.PageSize, err)
		return response.InternalServerCtx(c, "获取学习目标失败")
	}
	log.Printf("📚 学习目标分页查询: user_id=%s page=%d page_size=%d total=%d returned=%d", userID, req.Page, req.PageSize, total, len(goals))
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
