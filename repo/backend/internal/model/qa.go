package model

import "time"

type Question struct {
	ID        uint64    `db:"id" json:"-"`
	UUID      string    `db:"uuid" json:"id"`
	ItemID    uint64    `db:"item_id" json:"-"`
	UserID    uint64    `db:"user_id" json:"-"`
	Body      string    `db:"body" json:"body"`
	IsDeleted bool      `db:"is_deleted" json:"-"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type Answer struct {
	ID         uint64    `db:"id" json:"-"`
	UUID       string    `db:"uuid" json:"id"`
	QuestionID uint64    `db:"question_id" json:"-"`
	UserID     uint64    `db:"user_id" json:"-"`
	Body       string    `db:"body" json:"body"`
	IsDeleted  bool      `db:"is_deleted" json:"-"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}
