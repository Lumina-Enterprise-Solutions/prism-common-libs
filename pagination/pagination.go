// common/prism-common-libs/pagination/pagination.go
package pagination

import (
	"math"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	DefaultPage  = 1
	DefaultLimit = 10
	MaxLimit     = 100
)

// Params menampung semua parameter yang diekstrak untuk paginasi, sorting, dan filtering.
type Params struct {
	Page   int
	Limit  int
	Offset int
	SortBy string
	Order  string
	// Filters adalah map fleksibel untuk parameter query lainnya.
	Filters map[string]string
}

// Response adalah struktur standar untuk respons berpaginasi.
// Menggunakan generics [T any] agar bisa digunakan untuk data apa pun (User, Role, dll).
type Response[T any] struct {
	Data       []T `json:"data"`
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

// GetParams mengekstrak parameter paginasi, sorting, dan filter dari Gin context.
// allowedSorts adalah whitelist kolom yang diizinkan untuk sorting.
func GetParams(c *gin.Context, allowedSorts map[string]bool) *Params {
	page, _ := strconv.Atoi(c.DefaultQuery("page", strconv.Itoa(DefaultPage)))
	if page < 1 {
		page = DefaultPage
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(DefaultLimit)))
	if limit < 1 || limit > MaxLimit {
		limit = DefaultLimit
	}

	sortBy := c.DefaultQuery("sort_by", "created_at")
	if !allowedSorts[sortBy] {
		sortBy = "created_at" // Fallback ke default jika tidak diizinkan
	}

	order := strings.ToLower(c.DefaultQuery("order", "desc"))
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	// Ekstrak filter lain yang tidak termasuk dalam parameter standar
	filters := make(map[string]string)
	for key, values := range c.Request.URL.Query() {
		if key != "page" && key != "limit" && key != "sort_by" && key != "order" {
			if len(values) > 0 {
				filters[key] = values[0]
			}
		}
	}

	return &Params{
		Page:    page,
		Limit:   limit,
		Offset:  (page - 1) * limit,
		SortBy:  sortBy,
		Order:   order,
		Filters: filters,
	}
}

// NewResponse membuat objek respons paginasi standar.
func NewResponse[T any](data []T, totalItems int, params Params) Response[T] {
	totalPages := 0
	if totalItems > 0 {
		totalPages = int(math.Ceil(float64(totalItems) / float64(params.Limit)))
	}

	return Response[T]{
		Data:       data,
		Page:       params.Page,
		Limit:      params.Limit,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}
}
