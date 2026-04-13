package model

import "time"

type ExperimentStatus string

const (
	ExperimentStatusDraft      ExperimentStatus = "draft"
	ExperimentStatusRunning    ExperimentStatus = "running"
	ExperimentStatusPaused     ExperimentStatus = "paused"
	ExperimentStatusCompleted  ExperimentStatus = "completed"
	ExperimentStatusRolledBack ExperimentStatus = "rolled_back"
)

type Experiment struct {
	ID            uint64           `db:"id" json:"-"`
	UUID          string           `db:"uuid" json:"id"`
	Name          string           `db:"name" json:"name"`
	Slug          *string          `db:"slug" json:"slug,omitempty"`
	Description   *string          `db:"description" json:"description"`
	Status        ExperimentStatus `db:"status" json:"status"`
	HashSalt      string           `db:"hash_salt" json:"-"`
	MinSampleSize uint32           `db:"min_sample_size" json:"min_sample_size"`
	CreatedBy     uint64           `db:"created_by" json:"-"`
	StartedAt     *time.Time       `db:"started_at" json:"started_at"`
	EndedAt       *time.Time       `db:"ended_at" json:"ended_at"`
	CreatedAt     time.Time        `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time        `db:"updated_at" json:"updated_at"`
}

type ExperimentVariant struct {
	ID           uint64  `db:"id" json:"-"`
	ExperimentID uint64  `db:"experiment_id" json:"-"`
	Name         string  `db:"name" json:"name"`
	TrafficPct   float64 `db:"traffic_pct" json:"traffic_pct"`
	Config       []byte  `db:"config" json:"config"`
}

type ExperimentAssignment struct {
	ID           uint64    `db:"id" json:"-"`
	ExperimentID uint64    `db:"experiment_id" json:"-"`
	UserID       uint64    `db:"user_id" json:"-"`
	VariantID    uint64    `db:"variant_id" json:"-"`
	AssignedAt   time.Time `db:"assigned_at" json:"assigned_at"`
}

type ExperimentExposure struct {
	ID             uint64    `db:"id" json:"-"`
	ExperimentID   uint64    `db:"experiment_id" json:"-"`
	UserID         uint64    `db:"user_id" json:"-"`
	VariantID      uint64    `db:"variant_id" json:"-"`
	IdempotencyKey *string   `db:"idempotency_key" json:"-"`
	ExposedAt      time.Time `db:"exposed_at" json:"exposed_at"`
}

type ExperimentStatusHistory struct {
	ID           uint64    `db:"id" json:"-"`
	ExperimentID uint64    `db:"experiment_id" json:"-"`
	OldStatus    string    `db:"old_status" json:"old_status"`
	NewStatus    string    `db:"new_status" json:"new_status"`
	ChangedBy    uint64    `db:"changed_by" json:"-"`
	ChangedAt    time.Time `db:"changed_at" json:"changed_at"`
	Reason       *string   `db:"reason" json:"reason"`
}
