package request

import "encoding/json"

type BatchEventsRequest struct {
	SessionUUID string         `json:"session_uuid" binding:"required"`
	Events      []EventRequest `json:"events" binding:"required,min=1,max=100"`
}

type EventRequest struct {
	EventType    string `json:"event_type" binding:"required,oneof=impression click dwell favorite share comment"`
	ItemID       string `json:"item_id" binding:"omitempty"`
	DwellSeconds int    `json:"dwell_seconds" binding:"omitempty,min=0,max=600"`
	EventData    []byte `json:"event_data" binding:"omitempty"`
	ClientTS     string `json:"client_ts" binding:"required"`
}

type CreateSessionRequest struct {
	ItemID string `json:"item_id" binding:"omitempty"`
}

type DashboardFilterRequest struct {
	ItemID    string `form:"item_id"`
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
	Sentiment string `form:"sentiment" binding:"omitempty,oneof=positive neutral negative"`
	Keywords  string `form:"keywords"`
}

type CreateSavedViewRequest struct {
	Name         string          `json:"name" binding:"required,max=128"`
	FilterConfig json.RawMessage `json:"filter_config" binding:"required"`
}

type UpdateSavedViewRequest struct {
	Name         string          `json:"name" binding:"omitempty,max=128"`
	FilterConfig json.RawMessage `json:"filter_config" binding:"omitempty"`
}

type UpdateScoringWeightsRequest struct {
	ImpressionW float64 `json:"impression_w" binding:"required,min=0,max=1"`
	ClickW      float64 `json:"click_w" binding:"required,min=0,max=1"`
	DwellW      float64 `json:"dwell_w" binding:"required,min=0,max=1"`
	FavoriteW   float64 `json:"favorite_w" binding:"required,min=0,max=1"`
	ShareW      float64 `json:"share_w" binding:"required,min=0,max=1"`
	CommentW    float64 `json:"comment_w" binding:"required,min=0,max=1"`
}
