package model

import (
	"encoding/json"
	"time"
)

type EventType string

const (
	EventTypeImpression EventType = "impression"
	EventTypeClick      EventType = "click"
	EventTypeDwell      EventType = "dwell"
	EventTypeFavorite   EventType = "favorite"
	EventTypeShare      EventType = "share"
	EventTypeComment    EventType = "comment"
)

type AnalyticsSession struct {
	ID                  uint64     `db:"id" json:"-"`
	SessionUUID         string     `db:"session_uuid" json:"session_id"`
	UserID              *uint64    `db:"user_id" json:"-"`
	ItemID              *uint64    `db:"item_id" json:"item_id,omitempty"`
	ExperimentVariantID *uint64    `db:"experiment_variant_id" json:"experiment_variant_id,omitempty"`
	StartedAt           time.Time  `db:"started_at" json:"started_at"`
	EndedAt             *time.Time `db:"ended_at" json:"ended_at"`
	LastActiveAt        time.Time  `db:"last_active_at" json:"last_active_at"`
	UserAgent           *string    `db:"user_agent" json:"-"`
	IPAddress           *string    `db:"ip_address" json:"-"`
}

type BehaviorEvent struct {
	ID           uint64    `db:"id" json:"-"`
	SessionID    uint64    `db:"session_id" json:"-"`
	UserID       *uint64   `db:"user_id" json:"-"`
	EventType    EventType `db:"event_type" json:"event_type"`
	ItemID       *uint64   `db:"item_id" json:"item_id,omitempty"`
	DwellSeconds *uint16   `db:"dwell_seconds" json:"dwell_seconds,omitempty"`
	EventData    []byte    `db:"event_data" json:"event_data,omitempty"`
	ClientTS     time.Time `db:"client_ts" json:"client_ts"`
	ServerTS     time.Time `db:"server_ts" json:"server_ts"`
	DedupHash    *string   `db:"dedup_hash" json:"-"`
}

type SessionSequenceFingerprint struct {
	ID           uint64    `db:"id" json:"-"`
	UserID       uint64    `db:"user_id" json:"-"`
	SessionID    uint64    `db:"session_id" json:"-"`
	SequenceHash string    `db:"sequence_hash" json:"sequence_hash"`
	EventCount   uint32    `db:"event_count" json:"event_count"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

type AnalyticsAggregate struct {
	ID           uint64    `db:"id" json:"-"`
	ItemID       uint64    `db:"item_id" json:"item_id"`
	PeriodStart  time.Time `db:"period_start" json:"period_start"`
	Impressions  uint32    `db:"impressions" json:"impressions"`
	Clicks       uint32    `db:"clicks" json:"clicks"`
	AvgDwellSecs float64   `db:"avg_dwell_secs" json:"avg_dwell_secs"`
	Favorites    uint32    `db:"favorites" json:"favorites"`
	Shares       uint32    `db:"shares" json:"shares"`
	Comments     uint32    `db:"comments" json:"comments"`
	ComputedAt   time.Time `db:"computed_at" json:"computed_at"`
}

type UserEventCount struct {
	ID         uint64    `db:"id" json:"-"`
	UserID     uint64    `db:"user_id" json:"-"`
	HourBucket time.Time `db:"hour_bucket" json:"hour_bucket"`
	EventCount uint32    `db:"event_count" json:"event_count"`
}

type ScoringWeights struct {
	ID           uint64    `db:"id" json:"-"`
	Name         string    `db:"name" json:"name"`
	ImpressionW  float64   `db:"impression_w" json:"impression_w"`
	ClickW       float64   `db:"click_w" json:"click_w"`
	DwellW       float64   `db:"dwell_w" json:"dwell_w"`
	FavoriteW    float64   `db:"favorite_w" json:"favorite_w"`
	ShareW       float64   `db:"share_w" json:"share_w"`
	CommentW     float64   `db:"comment_w" json:"comment_w"`
	IsActive     bool      `db:"is_active" json:"is_active"`
	Version      uint32    `db:"version" json:"version"`
	UpdatedBy    *uint64   `db:"updated_by" json:"-"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

type CooccurrenceTerm struct {
	ID          uint64    `db:"id" json:"-"`
	TermA       string    `db:"term_a" json:"term_a"`
	TermB       string    `db:"term_b" json:"term_b"`
	ItemID      uint64    `db:"item_id" json:"item_id"`
	Frequency   uint32    `db:"frequency" json:"frequency"`
	PeriodStart time.Time `db:"period_start" json:"period_start"`
}

type SavedView struct {
	ID           uint64    `db:"id" json:"-"`
	UUID         string    `db:"uuid" json:"id"`
	UserID       uint64    `db:"user_id" json:"-"`
	Name         string    `db:"name" json:"name"`
	FilterConfig json.RawMessage `db:"filter_config" json:"filter_config"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

type ShareLink struct {
	ID          uint64    `db:"id" json:"-"`
	Token       string    `db:"token" json:"token"`
	SavedViewID uint64    `db:"saved_view_id" json:"-"`
	CreatedBy   uint64    `db:"created_by" json:"-"`
	ExpiresAt   time.Time `db:"expires_at" json:"expires_at"`
	IsRevoked   bool      `db:"is_revoked" json:"is_revoked"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}
