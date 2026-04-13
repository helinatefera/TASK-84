package request

type CreateItemRequest struct {
	Title       string `json:"title" binding:"required,max=255"`
	Description string `json:"description" binding:"omitempty"`
	Category    string `json:"category" binding:"omitempty,max=128"`
}

type UpdateItemRequest struct {
	Title       string `json:"title" binding:"omitempty,max=255"`
	Description string `json:"description" binding:"omitempty"`
	Category    string `json:"category" binding:"omitempty,max=128"`
}

type TransitionItemRequest struct {
	State string `json:"state" binding:"required,oneof=draft published archived"`
}
