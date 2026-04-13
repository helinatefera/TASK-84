package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/localinsights/portal/internal/dto/request"
	"github.com/localinsights/portal/internal/dto/response"
	"github.com/localinsights/portal/internal/middleware"
	"github.com/localinsights/portal/internal/model"
	"github.com/localinsights/portal/internal/pkg/database"
	"github.com/localinsights/portal/internal/repository"
	"github.com/localinsights/portal/internal/service"
)

type QAHandler struct {
	questionRepo  repository.QuestionRepository
	answerRepo    repository.AnswerRepository
	itemRepo      repository.ItemRepository
	contentFilter *service.ContentFilter
	db            *database.DB
}

func NewQAHandler(questionRepo repository.QuestionRepository, answerRepo repository.AnswerRepository, itemRepo repository.ItemRepository, contentFilter *service.ContentFilter, db *database.DB) *QAHandler {
	return &QAHandler{
		questionRepo:  questionRepo,
		answerRepo:    answerRepo,
		itemRepo:      itemRepo,
		contentFilter: contentFilter,
		db:            db,
	}
}

// ListQuestions handles GET /items/:id/questions
// The :id param is the item's UUID.
func (h *QAHandler) ListQuestions(c *gin.Context) {
	itemUUID := c.Param("id")
	item, err := h.itemRepo.GetByUUID(c.Request.Context(), itemUUID)
	if err != nil || item == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": http.StatusNotFound, "msg": "Item not found"})
		return
	}
	itemID := item.ID

	pg := getPagination(c)

	questions, total, err := h.questionRepo.ListByItem(c.Request.Context(), itemID, pg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "Failed to list questions"})
		return
	}

	c.JSON(http.StatusOK, response.NewPaginated(questions, pg.Page, pg.PerPage, total))
}

// CreateQuestion handles POST /items/:id/questions
func (h *QAHandler) CreateQuestion(c *gin.Context) {
	if !middleware.IsAuthenticated(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "msg": "Authentication required"})
		return
	}

	itemUUID := c.Param("id")
	item, err := h.itemRepo.GetByUUID(c.Request.Context(), itemUUID)
	if err != nil || item == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": http.StatusNotFound, "msg": "Item not found"})
		return
	}
	itemID := item.ID

	var req request.CreateQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"code": http.StatusUnprocessableEntity, "msg": err.Error()})
		return
	}

	userID := middleware.GetUserID(c)
	now := time.Now().UTC()

	// Run content through sensitive-word filter.
	cleaned, blockReason, flagged := h.contentFilter.Apply(c.Request.Context(), req.Body)
	if blockReason != "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"code": 422, "msg": blockReason})
		return
	}
	if flagged {
		h.db.ExecContext(c.Request.Context(),
			`INSERT INTO audit_logs (actor_id, action, target_type, details, created_at)
			VALUES (?, 'content.flagged', 'question', '{"reason":"sensitive_word_match"}', NOW(3))`, userID)
	}

	question := &model.Question{
		UUID:      uuid.New().String(),
		ItemID:    itemID,
		UserID:    userID,
		Body:      cleaned,
		IsDeleted: false,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := h.questionRepo.Create(c.Request.Context(), question); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "Failed to create question"})
		return
	}

	c.JSON(http.StatusCreated, question)
}

// UpdateQuestion handles PUT /questions/:id
func (h *QAHandler) UpdateQuestion(c *gin.Context) {
	if !middleware.IsAuthenticated(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "msg": "Authentication required"})
		return
	}

	questionUUID := c.Param("id")
	question, err := h.questionRepo.GetByUUID(c.Request.Context(), questionUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "Failed to retrieve question"})
		return
	}
	if question == nil || question.IsDeleted {
		c.JSON(http.StatusNotFound, gin.H{"code": http.StatusNotFound, "msg": "Question not found"})
		return
	}

	userID := middleware.GetUserID(c)
	if question.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"code": http.StatusForbidden, "msg": "You can only edit your own questions"})
		return
	}

	var req request.UpdateQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"code": http.StatusUnprocessableEntity, "msg": err.Error()})
		return
	}

	cleanedQ, blockReasonQ, flaggedQ := h.contentFilter.Apply(c.Request.Context(), req.Body)
	if blockReasonQ != "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"code": 422, "msg": blockReasonQ})
		return
	}
	if flaggedQ {
		h.db.ExecContext(c.Request.Context(),
			`INSERT INTO audit_logs (actor_id, action, target_type, details, created_at)
			VALUES (?, 'content.flagged', 'question', '{"reason":"sensitive_word_match"}', NOW(3))`, userID)
	}
	question.Body = cleanedQ
	question.UpdatedAt = time.Now().UTC()

	if err := h.questionRepo.Update(c.Request.Context(), question); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "Failed to update question"})
		return
	}

	c.JSON(http.StatusOK, question)
}

// DeleteQuestion handles DELETE /questions/:id
func (h *QAHandler) DeleteQuestion(c *gin.Context) {
	if !middleware.IsAuthenticated(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "msg": "Authentication required"})
		return
	}

	questionUUID := c.Param("id")
	question, err := h.questionRepo.GetByUUID(c.Request.Context(), questionUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "Failed to retrieve question"})
		return
	}
	if question == nil || question.IsDeleted {
		c.JSON(http.StatusNotFound, gin.H{"code": http.StatusNotFound, "msg": "Question not found"})
		return
	}

	userID := middleware.GetUserID(c)
	if question.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"code": http.StatusForbidden, "msg": "You can only delete your own questions"})
		return
	}

	if err := h.questionRepo.SoftDelete(c.Request.Context(), question.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "Failed to delete question"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "msg": "Question deleted"})
}

// ListAnswers handles GET /questions/:id/answers
func (h *QAHandler) ListAnswers(c *gin.Context) {
	questionUUID := c.Param("id")
	question, err := h.questionRepo.GetByUUID(c.Request.Context(), questionUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "Failed to retrieve question"})
		return
	}
	if question == nil || question.IsDeleted {
		c.JSON(http.StatusNotFound, gin.H{"code": http.StatusNotFound, "msg": "Question not found"})
		return
	}

	pg := getPagination(c)

	answers, total, err := h.answerRepo.ListByQuestion(c.Request.Context(), question.ID, pg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "Failed to list answers"})
		return
	}

	c.JSON(http.StatusOK, response.NewPaginated(answers, pg.Page, pg.PerPage, total))
}

// CreateAnswer handles POST /questions/:id/answers
func (h *QAHandler) CreateAnswer(c *gin.Context) {
	if !middleware.IsAuthenticated(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "msg": "Authentication required"})
		return
	}

	questionUUID := c.Param("id")
	question, err := h.questionRepo.GetByUUID(c.Request.Context(), questionUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "Failed to retrieve question"})
		return
	}
	if question == nil || question.IsDeleted {
		c.JSON(http.StatusNotFound, gin.H{"code": http.StatusNotFound, "msg": "Question not found"})
		return
	}

	var req request.CreateAnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"code": http.StatusUnprocessableEntity, "msg": err.Error()})
		return
	}

	userID := middleware.GetUserID(c)
	now := time.Now().UTC()

	// Run content through sensitive-word filter.
	cleanedBody, blockReason, flaggedA := h.contentFilter.Apply(c.Request.Context(), req.Body)
	if blockReason != "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"code": 422, "msg": blockReason})
		return
	}
	if flaggedA {
		h.db.ExecContext(c.Request.Context(),
			`INSERT INTO audit_logs (actor_id, action, target_type, details, created_at)
			VALUES (?, 'content.flagged', 'answer', '{"reason":"sensitive_word_match"}', NOW(3))`, userID)
	}

	answer := &model.Answer{
		UUID:       uuid.New().String(),
		QuestionID: question.ID,
		UserID:     userID,
		Body:       cleanedBody,
		IsDeleted:  false,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := h.answerRepo.Create(c.Request.Context(), answer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "Failed to create answer"})
		return
	}

	c.JSON(http.StatusCreated, answer)
}

// UpdateAnswer handles PUT /answers/:id
func (h *QAHandler) UpdateAnswer(c *gin.Context) {
	if !middleware.IsAuthenticated(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "msg": "Authentication required"})
		return
	}

	answerUUID := c.Param("id")
	answer, err := h.answerRepo.GetByUUID(c.Request.Context(), answerUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "Failed to retrieve answer"})
		return
	}
	if answer == nil || answer.IsDeleted {
		c.JSON(http.StatusNotFound, gin.H{"code": http.StatusNotFound, "msg": "Answer not found"})
		return
	}

	userID := middleware.GetUserID(c)
	if answer.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"code": http.StatusForbidden, "msg": "You can only edit your own answers"})
		return
	}

	var req request.UpdateAnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"code": http.StatusUnprocessableEntity, "msg": err.Error()})
		return
	}

	cleanedAU, blockReasonAU, flaggedAU := h.contentFilter.Apply(c.Request.Context(), req.Body)
	if blockReasonAU != "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"code": 422, "msg": blockReasonAU})
		return
	}
	if flaggedAU {
		h.db.ExecContext(c.Request.Context(),
			`INSERT INTO audit_logs (actor_id, action, target_type, details, created_at)
			VALUES (?, 'content.flagged', 'answer', '{"reason":"sensitive_word_match"}', NOW(3))`, userID)
	}
	answer.Body = cleanedAU
	answer.UpdatedAt = time.Now().UTC()

	if err := h.answerRepo.Update(c.Request.Context(), answer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "Failed to update answer"})
		return
	}

	c.JSON(http.StatusOK, answer)
}

// DeleteAnswer handles DELETE /answers/:id
func (h *QAHandler) DeleteAnswer(c *gin.Context) {
	if !middleware.IsAuthenticated(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "msg": "Authentication required"})
		return
	}

	answerUUID := c.Param("id")
	answer, err := h.answerRepo.GetByUUID(c.Request.Context(), answerUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "Failed to retrieve answer"})
		return
	}
	if answer == nil || answer.IsDeleted {
		c.JSON(http.StatusNotFound, gin.H{"code": http.StatusNotFound, "msg": "Answer not found"})
		return
	}

	userID := middleware.GetUserID(c)
	if answer.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"code": http.StatusForbidden, "msg": "You can only delete your own answers"})
		return
	}

	if err := h.answerRepo.SoftDelete(c.Request.Context(), answer.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "msg": "Failed to delete answer"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "msg": "Answer deleted"})
}
