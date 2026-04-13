package audit

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/localinsights/portal/internal/model"
	"github.com/localinsights/portal/internal/repository"
)

var sensitiveFields = map[string]bool{
	"password":      true,
	"password_hash": true,
	"token":         true,
	"refresh_token": true,
	"csrf_token":    true,
	"secret":        true,
}

type Logger struct {
	repo repository.AuditLogRepository
}

func NewLogger(repo repository.AuditLogRepository) *Logger {
	return &Logger{repo: repo}
}

func (l *Logger) Log(ctx context.Context, actorID *uint64, actorRole *string, action string, targetType *string, targetID *uint64, ipAddress *string, requestID *string, details map[string]any) {
	masked := MaskDetails(details)
	detailsJSON, _ := json.Marshal(masked)

	entry := &model.AuditLog{
		ActorID:    actorID,
		ActorRole:  actorRole,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		IPAddress:  ipAddress,
		RequestID:  requestID,
		Details:    detailsJSON,
	}

	_ = l.repo.Create(ctx, entry)
}

func MaskDetails(details map[string]any) map[string]any {
	if details == nil {
		return nil
	}
	masked := make(map[string]any, len(details))
	for k, v := range details {
		if sensitiveFields[strings.ToLower(k)] {
			masked[k] = "***REDACTED***"
		} else if k == "email" {
			if email, ok := v.(string); ok {
				masked[k] = maskEmail(email)
			} else {
				masked[k] = v
			}
		} else {
			masked[k] = v
		}
	}
	return masked
}

func maskEmail(email string) string {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 || len(parts[0]) == 0 {
		return "***@***"
	}
	return string(parts[0][0]) + "***@" + parts[1]
}
