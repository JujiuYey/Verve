package option

// 通用选项结构体，用于下拉框等场景
type Option struct {
	Value       string  `bun:"id" json:"value"`                // ID
	Label       string  `bun:"name" json:"label"`              // 名称
	Description *string `bun:"description" json:"description"` // 描述
}
