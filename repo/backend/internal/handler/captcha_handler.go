package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/localinsights/portal/internal/pkg/captcha"
)

type CaptchaHandler struct {
	store *captcha.Store
}

func NewCaptchaHandler(store *captcha.Store) *CaptchaHandler {
	return &CaptchaHandler{store: store}
}

// Generate creates a new captcha challenge and returns the ID and image.
func (h *CaptchaHandler) Generate(c *gin.Context) {
	challenge, err := h.store.Generate()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "failed to generate captcha"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"captcha_id":    challenge.ID,
		"captcha_image": challenge.Image,
	})
}

// Verify checks a captcha answer and returns whether it is valid.
func (h *CaptchaHandler) Verify(c *gin.Context) {
	var req struct {
		CaptchaID     string `json:"captcha_id" binding:"required"`
		CaptchaAnswer string `json:"captcha_answer" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"code": http.StatusUnprocessableEntity, "msg": err.Error()})
		return
	}

	valid := h.store.Verify(req.CaptchaID, req.CaptchaAnswer)

	c.JSON(http.StatusOK, gin.H{"valid": valid})
}
