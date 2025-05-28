package utils

import "math"

// PaginationParams 分页参数
type PaginationParams struct {
	Page     int `json:"page" form:"page" validate:"min=1"`
	PageSize int `json:"page_size" form:"page_size" validate:"min=1,max=100"`
}

// PaginatedResult 分页结果
type PaginatedResult[T any] struct {
	Data       []T `json:"data"`
	TotalCount int `json:"total_count"`
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalPages int `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

// NewPaginatedResult 创建分页结果
func NewPaginatedResult[T any](data []T, totalCount, page, pageSize int) *PaginatedResult[T] {
	totalPages := int(math.Ceil(float64(totalCount) / float64(pageSize)))
	
	return &PaginatedResult[T]{
		Data:       data,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

// DefaultPaginationParams 默认分页参数
func DefaultPaginationParams() PaginationParams {
	return PaginationParams{
		Page:     1,
		PageSize: 20,
	}
}

// ValidatePaginationParams 验证分页参数
func ValidatePaginationParams(params *PaginationParams) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}
	if params.PageSize > 100 {
		params.PageSize = 100
	}
} 