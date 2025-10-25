package model

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Pagination 分页参数
type Pagination struct {
	Page      int    `json:"page"`       // 页码，从1开始
	PageSize  int    `json:"page_size"`  // 每页数量
	SortBy    string `json:"sort_by"`    // 排序字段
	SortOrder string `json:"sort_order"` // 排序顺序: asc, desc
}

// NewPagination 创建分页参数
func NewPagination(page, pageSize int) *Pagination {
	// 设置默认值
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	return &Pagination{
		Page:     page,
		PageSize: pageSize,
	}
}

// NewPaginationWithSort 创建带排序的分页参数
func NewPaginationWithSort(page, pageSize int, sortBy, sortOrder string) *Pagination {
	pagination := NewPagination(page, pageSize)
	pagination.SortBy = sortBy
	pagination.SortOrder = sortOrder
	return pagination
}

// NewPaginationWithDefaultSort 创建带默认排序的分页参数
func NewPaginationWithDefaultSort(page, pageSize int) *Pagination {
	return NewPaginationWithSort(page, pageSize, "createdTime", "desc")
}

// ToFindOptions 转换为 MongoDB FindOptions
func (p *Pagination) ToFindOptions() *options.FindOptions {
	findOptions := options.Find()

	// 计算跳过的记录数
	skip := int64((p.Page - 1) * p.PageSize)
	limit := int64(p.PageSize)

	findOptions.SetSkip(skip)
	findOptions.SetLimit(limit)

	// 设置排序
	if p.SortBy != "" {
		sortOrder := 1 // 默认升序
		if p.SortOrder == "desc" {
			sortOrder = -1
		}
		findOptions.SetSort(bson.D{bson.E{Key: p.SortBy, Value: sortOrder}})
	}

	return findOptions
}

// GetOffset 获取偏移量（用于其他数据库）
func (p *Pagination) GetOffset() int {
	return (p.Page - 1) * p.PageSize
}

// IsEmpty 判断是否为空分页（不分页）
func (p *Pagination) IsEmpty() bool {
	return p == nil || (p.Page <= 0 && p.PageSize <= 0)
}

// IsValid 判断是否为有效的分页参数
func (p *Pagination) IsValid() bool {
	return p != nil && p.Page > 0 && p.PageSize > 0
}
