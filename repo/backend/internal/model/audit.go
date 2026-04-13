package model

import "time"

type AuditLog struct {
	ID         uint64    `db:"id" json:"-"`
	ActorID    *uint64   `db:"actor_id" json:"actor_id"`
	ActorRole  *string   `db:"actor_role" json:"actor_role"`
	Action     string    `db:"action" json:"action"`
	TargetType *string   `db:"target_type" json:"target_type"`
	TargetID   *uint64   `db:"target_id" json:"target_id"`
	IPAddress  *string   `db:"ip_address" json:"ip_address"`
	RequestID  *string   `db:"request_id" json:"request_id"`
	Details    []byte    `db:"details" json:"details"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

type IdempotencyKey struct {
	ID           uint64    `db:"id" json:"-"`
	KeyHash      string    `db:"key_hash" json:"-"`
	UserID       uint64    `db:"user_id" json:"-"`
	Endpoint     string    `db:"endpoint" json:"-"`
	ResponseCode *uint16   `db:"response_code" json:"-"`
	ResponseBody *string   `db:"response_body" json:"-"`
	CreatedAt    time.Time `db:"created_at" json:"-"`
	ExpiresAt    time.Time `db:"expires_at" json:"-"`
}

type JobState struct {
	JobName   string     `db:"job_name" json:"job_name"`
	LastRunAt time.Time  `db:"last_run_at" json:"last_run_at"`
	Watermark *string    `db:"watermark" json:"watermark"`
	Status    string     `db:"status" json:"status"`
	LastError *string    `db:"last_error" json:"last_error"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
}

type RecoveryDrill struct {
	ID             uint64     `db:"id" json:"-"`
	BackupFile     string     `db:"backup_file" json:"backup_file"`
	StartedAt      time.Time  `db:"started_at" json:"started_at"`
	CompletedAt    *time.Time `db:"completed_at" json:"completed_at"`
	Status         string     `db:"status" json:"status"`
	VerifiedTables uint32     `db:"verified_tables" json:"verified_tables"`
	ErrorLog       *string    `db:"error_log" json:"error_log"`
	TriggeredBy    *uint64    `db:"triggered_by" json:"-"`
}

type MonitoringMetric struct {
	ID          uint64    `db:"id" json:"-"`
	MetricName  string    `db:"metric_name" json:"metric_name"`
	MetricValue float64   `db:"metric_value" json:"metric_value"`
	Tags        []byte    `db:"tags" json:"tags"`
	RecordedAt  time.Time `db:"recorded_at" json:"recorded_at"`
}
