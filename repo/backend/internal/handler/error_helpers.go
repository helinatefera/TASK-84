package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/localinsights/portal/internal/errs"
)

// respondAppError handles either an *errs.AppError or a generic error by
// writing the appropriate JSON error response.
func respondAppError(c *gin.Context, err error) {
	var appErr *errs.AppError
	if errors.As(err, &appErr) {
		c.JSON(appErr.HTTPStatus, gin.H{"code": appErr.HTTPStatus, "msg": appErr.Message})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "An internal error occurred"})
	}
}
