package pagination

// 通用分页请求参数
type PaginationRequest struct {
	Page     int `query:"page"`      // 当前页码，从1开始
	PageSize int `query:"page_size"` // 每页数量
}

// 默认页码
const DefaultPage int = 1

// 默认每页数量
const DefaultPageSize int = 10

// 最大每页数量
const MaxPageSize int = 100

// 计算分页偏移量
func (p *PaginationRequest) GetOffset() int {
	if p.Page < 1 {
		p.Page = DefaultPage
	}
	return (p.Page - 1) * p.PageSize
}

// 验证并修正分页参数
func (p *PaginationRequest) Validate() {
	if p.Page < 1 {
		p.Page = DefaultPage
	}
	if p.PageSize < 1 || p.PageSize > MaxPageSize {
		p.PageSize = DefaultPageSize
	}
}

// 通用分页响应结构体
type PaginationResponse[T any] struct {
	Data      []T   `json:"data"`       // 数据列表
	Total     int64 `json:"total"`      // 总记录数
	Page      int   `json:"page"`       // 当前页码
	PageSize  int   `json:"page_size"`  // 每页数量
	TotalPage int   `json:"total_page"` // 总页数
}

// 创建分页响应
func NewPaginationResponse[T any](data []T, total int64, page, pageSize int) PaginationResponse[T] {
	totalPage := int((total + int64(pageSize) - 1) / int64(pageSize))
	return PaginationResponse[T]{
		Data:      data,
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
		TotalPage: totalPage,
	}
}
