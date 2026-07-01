package payload

// 创建学习目标请求(一句话主题驱动)
type CreateGoalRequest struct {
	Title string `json:"title"`
}

// 从文档管理文件夹创建学习目标
type CreateGoalFromFolderRequest struct {
	FolderID string `json:"folder_id"`
}

// 更新学习目标请求
type UpdateGoalRequest struct {
	ID     string  `json:"id"`     // 主键ID
	Title  *string `json:"title"`  // 标题
	Status *string `json:"status"` // active / archived / completed
}
