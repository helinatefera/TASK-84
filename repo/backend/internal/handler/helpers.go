package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/localinsights/portal/internal/repository"
)

// respondError sends a standardized JSON error response.
// Format: {"code": <http_status>, "msg": "<message>"}
// The second string parameter (code label) is accepted for compatibility but ignored
// in favor of the numeric HTTP status in the response body.
func respondError(c *gin.Context, status int, _ string, msg string) {
	c.JSON(status, gin.H{"code": status, "msg": msg})
}

// getPagination extracts page and per_page from query string with defaults.
func getPagination(c *gin.Context) repository.Pagination {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	return repository.Pagination{Page: page, PerPage: perPage}
}

// paginatedResponse builds a standard paginated JSON response.
func paginatedResponse(data any, p repository.Pagination, total int64) gin.H {
	totalPages := total / int64(p.PerPage)
	if total%int64(p.PerPage) != 0 {
		totalPages++
	}
	return gin.H{
		"data":        data,
		"page":        p.Page,
		"per_page":    p.PerPage,
		"total":       total,
		"total_pages": totalPages,
	}
}

// parseUintParam parses a uint64 from a URL parameter.
func parseUintParam(c *gin.Context, name string) (uint64, bool) {
	s := c.Param(name)
	id, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "Invalid " + name + " parameter"})
		return 0, false
	}
	return id, true
}
