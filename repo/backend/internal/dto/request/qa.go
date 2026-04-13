package request

type CreateQuestionRequest struct {
	Body string `json:"body" binding:"required,min=10,max=2000"`
}

type UpdateQuestionRequest struct {
	Body string `json:"body" binding:"required,min=10,max=2000"`
}

type CreateAnswerRequest struct {
	Body string `json:"body" binding:"required,min=10,max=2000"`
}

type UpdateAnswerRequest struct {
	Body string `json:"body" binding:"required,min=10,max=2000"`
}
