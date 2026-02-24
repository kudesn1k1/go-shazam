package pagination

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 100
)

type Params struct {
	Page  int
	Limit int
}

type PaginatedResponse[T any] struct {
	Data  []T `json:"data"`
	Total int `json:"total"`
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

func ParseParams(c *gin.Context) Params {
	page := parseIntQuery(c, "page", DefaultPage)
	limit := parseIntQuery(c, "limit", DefaultLimit)

	if page < 1 {
		page = DefaultPage
	}
	if limit < 1 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}

	return Params{Page: page, Limit: limit}
}

func NewResponse[T any](data []T, total, page, limit int) PaginatedResponse[T] {
	if data == nil {
		data = []T{}
	}
	return PaginatedResponse[T]{
		Data:  data,
		Total: total,
		Page:  page,
		Limit: limit,
	}
}

func parseIntQuery(c *gin.Context, key string, defaultVal int) int {
	raw := c.Query(key)
	if raw == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(raw)
	if err != nil {
		return defaultVal
	}
	return val
}
