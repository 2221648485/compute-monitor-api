package page

import "gorm.io/gorm"

const (
	defaultPage = 1
	defaultSize = 20
	maxSize     = 100
)

// Query 是通用分页请求参数。
// Handler 可以通过 ShouldBindQuery 直接绑定 page 和 size。
type Query struct {
	Page int `form:"page"`
	Size int `form:"size"`
}

// Result 是通用分页响应结构。
type Result[T any] struct {
	Items []T   `json:"items"`
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Size  int   `json:"size"`
}

// Normalize 统一修正分页参数，避免每个模块重复写 page/size 兜底逻辑。
func Normalize(query Query) Query {
	if query.Page <= 0 {
		query.Page = defaultPage
	}
	if query.Size <= 0 {
		query.Size = defaultSize
	}
	if query.Size > maxSize {
		query.Size = maxSize
	}
	return query
}

// Offset 返回数据库分页偏移量。
func Offset(query Query) int {
	query = Normalize(query)
	return (query.Page - 1) * query.Size
}

// Apply 给 GORM 查询追加分页条件。
// Count 统计总数时不要调用 Apply，只在查询当前页数据时调用。
func Apply(db *gorm.DB, query Query) *gorm.DB {
	query = Normalize(query)
	return db.Offset(Offset(query)).Limit(query.Size)
}

// NewResult 创建统一分页返回结构。
func NewResult[T any](items []T, total int64, query Query) Result[T] {
	query = Normalize(query)
	return Result[T]{
		Items: items,
		Total: total,
		Page:  query.Page,
		Size:  query.Size,
	}
}

// Slice 对内存切片做分页，适合少量聚合结果或已经从其他系统读出的数据。
func Slice[T any](items []T, query Query) Result[T] {
	query = Normalize(query)
	total := int64(len(items))
	offset := Offset(query)
	if offset >= len(items) {
		return NewResult([]T{}, total, query)
	}

	end := offset + query.Size
	if end > len(items) {
		end = len(items)
	}
	return NewResult(items[offset:end], total, query)
}
