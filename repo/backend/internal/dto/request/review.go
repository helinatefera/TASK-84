package request

type CreateReviewRequest struct {
	Rating   int      `json:"rating" binding:"required,min=1,max=5"`
	Body     string   `json:"body" binding:"omitempty,max=2000"`
	ImageIDs []uint64 `json:"image_ids" binding:"omitempty,max=6"`
}

type UpdateReviewRequest struct {
	Rating int    `json:"rating" binding:"omitempty,min=1,max=5"`
	Body   string `json:"body" binding:"omitempty,max=2000"`
}

type FraudActionRequest struct {
	Action string `json:"action" binding:"required,oneof=confirm clear"`
	Notes  string `json:"notes" binding:"required,min=10"`
}
