package model

import "time"

type ReportCategory string

const (
	ReportCategorySpam           ReportCategory = "spam"
	ReportCategoryHarassment     ReportCategory = "harassment"
	ReportCategoryMisinformation ReportCategory = "misinformation"
	ReportCategoryInappropriate  ReportCategory = "inappropriate"
	ReportCategoryCopyright      ReportCategory = "copyright"
	ReportCategoryOther          ReportCategory = "other"
)

type ReportStatus string

const (
	ReportStatusPending   ReportStatus = "pending"
	ReportStatusInReview  ReportStatus = "in_review"
	ReportStatusResolved  ReportStatus = "resolved"
	ReportStatusDismissed ReportStatus = "dismissed"
)

type Report struct {
	ID              uint64         `db:"id" json:"-"`
	UUID            string         `db:"uuid" json:"id"`
	ReporterID      uint64         `db:"reporter_id" json:"-"`
	TargetType      string         `db:"target_type" json:"target_type"`
	TargetID        uint64         `db:"target_id" json:"target_id"`
	Category        ReportCategory `db:"category" json:"category"`
	Description     *string        `db:"description" json:"description"`
	Status          ReportStatus   `db:"status" json:"status"`
	Priority        uint8          `db:"priority" json:"priority"`
	AssignedTo      *uint64        `db:"assigned_to" json:"-"`
	ResolvedAt      *time.Time     `db:"resolved_at" json:"resolved_at"`
	ResolutionNote  *string        `db:"resolution_note" json:"resolution_note"`
	UserVisibleNote *string        `db:"user_visible_note" json:"user_visible_note,omitempty"`
	IdempotencyKey  *string        `db:"idempotency_key" json:"-"`
	CreatedAt       time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time      `db:"updated_at" json:"updated_at"`
}

type AppealStatus string

const (
	AppealStatusPending   AppealStatus = "pending"
	AppealStatusApproved  AppealStatus = "approved"
	AppealStatusRejected  AppealStatus = "rejected"
	AppealStatusNeedsEdit AppealStatus = "needs_edit"
)

type Appeal struct {
	ID         uint64       `db:"id" json:"-"`
	UUID       string       `db:"uuid" json:"id"`
	ReportID   uint64       `db:"report_id" json:"-"`
	UserID     uint64       `db:"user_id" json:"-"`
	Body       string       `db:"body" json:"body"`
	Status     AppealStatus `db:"status" json:"status"`
	ReviewedBy *uint64      `db:"reviewed_by" json:"-"`
	ReviewedAt *time.Time   `db:"reviewed_at" json:"reviewed_at"`
	CreatedAt  time.Time    `db:"created_at" json:"created_at"`
}

type ModerationNote struct {
	ID        uint64    `db:"id" json:"-"`
	ReportID  uint64    `db:"report_id" json:"-"`
	AuthorID  uint64    `db:"author_id" json:"-"`
	Body      string    `db:"body" json:"body"`
	IsInternal bool     `db:"is_internal" json:"is_internal"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type SensitiveWordRule struct {
	ID          uint64    `db:"id" json:"-"`
	Pattern     string    `db:"pattern" json:"pattern"`
	Action      string    `db:"action" json:"action"`
	Replacement *string   `db:"replacement" json:"replacement"`
	Version     uint32    `db:"version" json:"version"`
	IsActive    bool      `db:"is_active" json:"is_active"`
	CreatedBy   uint64    `db:"created_by" json:"-"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}
