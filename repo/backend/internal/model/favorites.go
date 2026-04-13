package model

import "time"

type Favorite struct {
	ID        uint64    `db:"id" json:"-"`
	UserID    uint64    `db:"user_id" json:"-"`
	ItemID    uint64    `db:"item_id" json:"-"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type Wishlist struct {
	ID        uint64    `db:"id" json:"-"`
	UUID      string    `db:"uuid" json:"id"`
	UserID    uint64    `db:"user_id" json:"-"`
	Name      string    `db:"name" json:"name"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type WishlistItem struct {
	ID         uint64    `db:"id" json:"-"`
	WishlistID uint64    `db:"wishlist_id" json:"-"`
	ItemID     uint64    `db:"item_id" json:"-"`
	AddedAt    time.Time `db:"added_at" json:"added_at"`
}
