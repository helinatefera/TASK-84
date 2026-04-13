package request

type CreateReportRequest struct {
	TargetType  string `json:"target_type" binding:"required,oneof=review question answer item user"`
	TargetID    string `json:"target_id" binding:"required"`
	Category    string `json:"category" binding:"required,oneof=spam harassment misinformation inappropriate copyright other"`
	Description string `json:"description" binding:"omitempty,max=1000"`
}

type UpdateReportRequest struct {
	Status          string `json:"status" binding:"required,oneof=pending in_review resolved dismissed"`
	ResolutionNote  string `json:"resolution_note" binding:"omitempty"`
	UserVisibleNote string `json:"user_visible_note" binding:"omitempty"`
}

type CreateAppealRequest struct {
	Body string `json:"body" binding:"required,min=10,max=2000"`
}

type HandleAppealRequest struct {
	Status string `json:"status" binding:"required,oneof=approved rejected needs_edit"`
	Note   string `json:"note" binding:"required,min=5"`
}

type CreateModerationNoteRequest struct {
	Body       string `json:"body" binding:"required,min=10"`
	IsInternal bool   `json:"is_internal"`
}

type CreateSensitiveWordRequest struct {
	Pattern     string `json:"pattern" binding:"required,max=255"`
	Action      string `json:"action" binding:"required,oneof=block flag replace"`
	Replacement string `json:"replacement" binding:"omitempty,max=255"`
}

type UpdateSensitiveWordRequest struct {
	Pattern     string `json:"pattern" binding:"omitempty,max=255"`
	Action      string `json:"action" binding:"omitempty,oneof=block flag replace"`
	Replacement string `json:"replacement" binding:"omitempty,max=255"`
	IsActive    *bool  `json:"is_active" binding:"omitempty"`
}

type QuarantineActionRequest struct {
	Action string `json:"action" binding:"required,oneof=approve reject keep"`
}
