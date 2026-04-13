package model

import "time"

type Role string

const (
	RoleAdmin          Role = "admin"
	RoleModerator      Role = "moderator"
	RoleProductAnalyst Role = "product_analyst"
	RoleRegularUser    Role = "regular_user"
)

type UserFraudStatus string

const (
	UserFraudClean     UserFraudStatus = "clean"
	UserFraudSuspected UserFraudStatus = "suspected"
	UserFraudConfirmed UserFraudStatus = "confirmed"
)

type User struct {
	ID           uint64          `db:"id" json:"-"`
	UUID         string          `db:"uuid" json:"id"`
	Username     string          `db:"username" json:"username"`
	Email        string          `db:"email" json:"email"`
	PasswordHash string          `db:"password_hash" json:"-"`
	Role         Role            `db:"role" json:"role"`
	IsActive     bool            `db:"is_active" json:"is_active"`
	FraudStatus  UserFraudStatus `db:"fraud_status" json:"fraud_status"`
	CreatedAt    time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time       `db:"updated_at" json:"updated_at"`
}

type UserPreferences struct {
	ID                   uint64 `db:"id" json:"-"`
	UserID               uint64 `db:"user_id" json:"-"`
	Locale               string `db:"locale" json:"locale"`
	Timezone             string `db:"timezone" json:"timezone"`
	NotificationSettings []byte `db:"notification_settings" json:"notification_settings"`
	CreatedAt            time.Time `db:"created_at" json:"-"`
	UpdatedAt            time.Time `db:"updated_at" json:"-"`
}

type LoginAttempt struct {
	ID          uint64    `db:"id"`
	IPAddress   string    `db:"ip_address"`
	Email       string    `db:"email"`
	AttemptedAt time.Time `db:"attempted_at"`
	Success     bool      `db:"success"`
}
