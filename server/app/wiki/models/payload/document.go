package payload

import "sag-wiki/common/pagination"

// 文档列表请求参数 - 组合通用分页参数 + 特有参数
type PageDocumentsRequest struct {
	pagination.PaginationRequest        // 匿名嵌入通用分页参数
	Name                         string `query:"name" form:"name"`
	FolderID                     string `query:"folder_id" form:"folder_id"`
}

// 更新文档内容请求
type UpdateContentRequest struct {
	Content string `json:"content"`
}

// 文档列表请求
type DocumentListRequest struct {
	Name     string `query:"name" form:"name"`
	FolderID string `query:"folder_id" form:"folder_id"`
}

// 文档列表响应
type PageDocumentsResponse struct {
	Documents interface{} `json:"documents"`
	Total     int         `json:"total"`
	Limit     int         `json:"limit"`
	Offset    int         `json:"offset"`
}
