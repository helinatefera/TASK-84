package model

import "time"

type LifecycleState string

const (
	LifecycleStateDraft     LifecycleState = "draft"
	LifecycleStatePublished LifecycleState = "published"
	LifecycleStateArchived  LifecycleState = "archived"
)

type Item struct {
	ID             uint64         `db:"id" json:"-"`
	UUID           string         `db:"uuid" json:"id"`
	Title          string         `db:"title" json:"title"`
	Description    *string        `db:"description" json:"description"`
	Category       *string        `db:"category" json:"category"`
	LifecycleState LifecycleState `db:"lifecycle_state" json:"lifecycle_state"`
	CreatedBy      uint64         `db:"created_by" json:"-"`
	PublishedAt    *time.Time     `db:"published_at" json:"published_at"`
	ArchivedAt     *time.Time     `db:"archived_at" json:"archived_at"`
	CreatedAt      time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at" json:"updated_at"`
}

type ItemRatingAggregate struct {
	ItemID        uint64    `db:"item_id" json:"item_id"`
	AvgRating     float64   `db:"avg_rating" json:"avg_rating"`
	RatingCount   uint32    `db:"rating_count" json:"rating_count"`
	Rating1       uint32    `db:"rating_1" json:"rating_1"`
	Rating2       uint32    `db:"rating_2" json:"rating_2"`
	Rating3       uint32    `db:"rating_3" json:"rating_3"`
	Rating4       uint32    `db:"rating_4" json:"rating_4"`
	Rating5       uint32    `db:"rating_5" json:"rating_5"`
	LastRefreshed time.Time `db:"last_refreshed" json:"last_refreshed"`
}
