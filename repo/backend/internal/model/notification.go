package model

import "time"

type NotificationTemplate struct {
	ID           uint64    `db:"id" json:"-"`
	TemplateKey  string    `db:"template_key" json:"template_key"`
	Locale       string    `db:"locale" json:"locale"`
	Subject      string    `db:"subject" json:"subject"`
	BodyTemplate string    `db:"body_template" json:"body_template"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

type Notification struct {
	ID              uint64     `db:"id" json:"-"`
	UserID          uint64     `db:"user_id" json:"-"`
	TemplateKey     string     `db:"template_key" json:"template_key"`
	Locale          string     `db:"locale" json:"locale"`
	RenderedSubject string     `db:"rendered_subject" json:"subject"`
	RenderedBody    string     `db:"rendered_body" json:"body"`
	Data            []byte     `db:"data" json:"data,omitempty"`
	IsRead          bool       `db:"is_read" json:"is_read"`
	ReadAt          *time.Time `db:"read_at" json:"read_at"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
}
