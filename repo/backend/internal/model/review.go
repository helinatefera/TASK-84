package model

import "time"

type FraudStatus string

const (
	FraudStatusNormal         FraudStatus = "normal"
	FraudStatusSuspected      FraudStatus = "suspected_fraud"
	FraudStatusConfirmed      FraudStatus = "confirmed_fraud"
	FraudStatusCleared        FraudStatus = "cleared"
)

type Review struct {
	ID             uint64      `db:"id" json:"-"`
	UUID           string      `db:"uuid" json:"id"`
	ItemID         uint64      `db:"item_id" json:"-"`
	UserID         uint64      `db:"user_id" json:"-"`
	Rating         uint8       `db:"rating" json:"rating"`
	Body           *string     `db:"body" json:"body"`
	FraudStatus    FraudStatus `db:"fraud_status" json:"fraud_status"`
	IdempotencyKey *string     `db:"idempotency_key" json:"-"`
	CreatedAt      time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time   `db:"updated_at" json:"updated_at"`
}

type ReviewImage struct {
	ID        uint64 `db:"id" json:"-"`
	ReviewID  uint64 `db:"review_id" json:"-"`
	ImageID   uint64 `db:"image_id" json:"-"`
	SortOrder uint8  `db:"sort_order" json:"sort_order"`
}

type Image struct {
	ID               uint64    `db:"id" json:"-"`
	SHA256Hash       string    `db:"sha256_hash" json:"hash"`
	OriginalName     string    `db:"original_name" json:"original_name"`
	MimeType         string    `db:"mime_type" json:"mime_type"`
	FileSize         uint32    `db:"file_size" json:"file_size"`
	StoragePath      string    `db:"storage_path" json:"-"`
	Width            *uint32   `db:"width" json:"width"`
	Height           *uint32   `db:"height" json:"height"`
	Status           string    `db:"status" json:"status"`
	QuarantineReason *string   `db:"quarantine_reason" json:"quarantine_reason,omitempty"`
	UploadedBy       uint64    `db:"uploaded_by" json:"-"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
}

type ReviewSentiment struct {
	ID             uint64    `db:"id" json:"-"`
	ReviewID       uint64    `db:"review_id" json:"-"`
	SentimentLabel string    `db:"sentiment_label" json:"sentiment_label"`
	Confidence     float64   `db:"confidence" json:"confidence"`
	ProcessedAt    time.Time `db:"processed_at" json:"processed_at"`
}

type ReviewKeyword struct {
	ID       uint64  `db:"id" json:"-"`
	ReviewID uint64  `db:"review_id" json:"-"`
	Keyword  string  `db:"keyword" json:"keyword"`
	Weight   float64 `db:"weight" json:"weight"`
}

type ReviewTopic struct {
	ID         uint64  `db:"id" json:"-"`
	ReviewID   uint64  `db:"review_id" json:"-"`
	Topic      string  `db:"topic" json:"topic"`
	Confidence float64 `db:"confidence" json:"confidence"`
}
