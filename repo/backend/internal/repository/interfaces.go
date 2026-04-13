package repository

import (
	"context"
	"time"

	"github.com/localinsights/portal/internal/model"
)

// Pagination holds page-based pagination parameters.
type Pagination struct {
	Page    int
	PerPage int
}

// Offset returns the SQL OFFSET for the current page.
func (p Pagination) Offset() int {
	return (p.Page - 1) * p.PerPage
}

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id uint64) (*model.User, error)
	GetByUUID(ctx context.Context, uuid string) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	List(ctx context.Context, page Pagination) ([]*model.User, int64, error)
	UpdateRole(ctx context.Context, id uint64, role model.Role) error
	SetActive(ctx context.Context, id uint64, active bool) error
}

type UserPreferencesRepository interface {
	Upsert(ctx context.Context, prefs *model.UserPreferences) error
	GetByUserID(ctx context.Context, userID uint64) (*model.UserPreferences, error)
}

type LoginAttemptRepository interface {
	Create(ctx context.Context, attempt *model.LoginAttempt) error
	CountRecentFailed(ctx context.Context, email string, ip string, window time.Duration) (int, error)
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, userID uint64, tokenHash string, expiresAt time.Time) error
	GetByHash(ctx context.Context, tokenHash string) (uint64, error) // returns user_id
	Revoke(ctx context.Context, tokenHash string) error
	RevokeAllForUser(ctx context.Context, userID uint64) error
}

type ItemRepository interface {
	Create(ctx context.Context, item *model.Item) error
	GetByID(ctx context.Context, id uint64) (*model.Item, error)
	GetByUUID(ctx context.Context, uuid string) (*model.Item, error)
	Update(ctx context.Context, item *model.Item) error
	ListPublished(ctx context.Context, search string, category string, page Pagination) ([]*model.Item, int64, error)
	GetRatingAggregate(ctx context.Context, itemID uint64) (*model.ItemRatingAggregate, error)
	UpsertRatingAggregate(ctx context.Context, agg *model.ItemRatingAggregate) error
}

type ReviewRepository interface {
	Create(ctx context.Context, review *model.Review) error
	GetByID(ctx context.Context, id uint64) (*model.Review, error)
	GetByUUID(ctx context.Context, uuid string) (*model.Review, error)
	ListByItem(ctx context.Context, itemID uint64, page Pagination) ([]*model.Review, int64, error)
	Update(ctx context.Context, review *model.Review) error
	Delete(ctx context.Context, id uint64) error
	UpdateFraudStatus(ctx context.Context, id uint64, status model.FraudStatus) error
	GetItemsWithReviewsSince(ctx context.Context, since time.Time) ([]uint64, error)
	ComputeAggregate(ctx context.Context, itemID uint64) (*model.ItemRatingAggregate, error)
}

type ImageRepository interface {
	Create(ctx context.Context, img *model.Image) error
	GetByHash(ctx context.Context, hash string) (*model.Image, error)
	GetByID(ctx context.Context, id uint64) (*model.Image, error)
	UpdateStatus(ctx context.Context, id uint64, status string, reason *string) error
	ListQuarantined(ctx context.Context, page Pagination) ([]*model.Image, int64, error)
}

type ReviewImageRepository interface {
	Create(ctx context.Context, ri *model.ReviewImage) error
	ListByReview(ctx context.Context, reviewID uint64) ([]*model.ReviewImage, error)
	DeleteByReview(ctx context.Context, reviewID uint64) error
}

type QuestionRepository interface {
	Create(ctx context.Context, q *model.Question) error
	GetByID(ctx context.Context, id uint64) (*model.Question, error)
	GetByUUID(ctx context.Context, uuid string) (*model.Question, error)
	ListByItem(ctx context.Context, itemID uint64, page Pagination) ([]*model.Question, int64, error)
	Update(ctx context.Context, q *model.Question) error
	SoftDelete(ctx context.Context, id uint64) error
}

type AnswerRepository interface {
	Create(ctx context.Context, a *model.Answer) error
	GetByID(ctx context.Context, id uint64) (*model.Answer, error)
	GetByUUID(ctx context.Context, uuid string) (*model.Answer, error)
	ListByQuestion(ctx context.Context, questionID uint64, page Pagination) ([]*model.Answer, int64, error)
	Update(ctx context.Context, a *model.Answer) error
	SoftDelete(ctx context.Context, id uint64) error
}

type FavoriteRepository interface {
	Add(ctx context.Context, userID, itemID uint64) error
	Remove(ctx context.Context, userID, itemID uint64) error
	ListByUser(ctx context.Context, userID uint64, page Pagination) ([]*model.Favorite, int64, error)
	Exists(ctx context.Context, userID, itemID uint64) (bool, error)
}

type WishlistRepository interface {
	Create(ctx context.Context, w *model.Wishlist) error
	GetByID(ctx context.Context, id uint64) (*model.Wishlist, error)
	GetByUUID(ctx context.Context, uuid string) (*model.Wishlist, error)
	ListByUser(ctx context.Context, userID uint64) ([]*model.Wishlist, error)
	Update(ctx context.Context, w *model.Wishlist) error
	Delete(ctx context.Context, id uint64) error
	AddItem(ctx context.Context, wishlistID, itemID uint64) error
	RemoveItem(ctx context.Context, wishlistID, itemID uint64) error
	ListItems(ctx context.Context, wishlistID uint64) ([]*model.WishlistItem, error)
}

type ReportRepository interface {
	Create(ctx context.Context, r *model.Report) error
	GetByID(ctx context.Context, id uint64) (*model.Report, error)
	GetByUUID(ctx context.Context, uuid string) (*model.Report, error)
	Update(ctx context.Context, r *model.Report) error
	ListQueue(ctx context.Context, status string, targetType string, page Pagination) ([]*model.Report, int64, error)
	ListByReporter(ctx context.Context, reporterID uint64, page Pagination) ([]*model.Report, int64, error)
}

type AppealRepository interface {
	Create(ctx context.Context, a *model.Appeal) error
	GetByReportID(ctx context.Context, reportID uint64) (*model.Appeal, error)
	Update(ctx context.Context, a *model.Appeal) error
	Resubmit(ctx context.Context, id uint64, body string) error
	List(ctx context.Context, status string, p Pagination) ([]*model.Appeal, int64, error)
}

type ModerationNoteRepository interface {
	Create(ctx context.Context, n *model.ModerationNote) error
	ListByReport(ctx context.Context, reportID uint64) ([]*model.ModerationNote, error)
}

type SensitiveWordRuleRepository interface {
	Create(ctx context.Context, r *model.SensitiveWordRule) error
	GetByID(ctx context.Context, id uint64) (*model.SensitiveWordRule, error)
	Update(ctx context.Context, r *model.SensitiveWordRule) error
	Delete(ctx context.Context, id uint64) error
	ListActive(ctx context.Context) ([]*model.SensitiveWordRule, error)
}

type NotificationRepository interface {
	Create(ctx context.Context, n *model.Notification) error
	GetByID(ctx context.Context, id uint64) (*model.Notification, error)
	ListByUser(ctx context.Context, userID uint64, unreadOnly bool, page Pagination) ([]*model.Notification, int64, error)
	MarkRead(ctx context.Context, id uint64) error
	MarkAllRead(ctx context.Context, userID uint64) error
	UnreadCount(ctx context.Context, userID uint64) (int64, error)
}

type NotificationTemplateRepository interface {
	GetByKeyAndLocale(ctx context.Context, key, locale string) (*model.NotificationTemplate, error)
}

type AnalyticsSessionRepository interface {
	Create(ctx context.Context, s *model.AnalyticsSession) error
	GetByUUID(ctx context.Context, uuid string) (*model.AnalyticsSession, error)
	UpdateHeartbeat(ctx context.Context, id uint64) error
	EndSession(ctx context.Context, id uint64) error
	GetExpiredSessions(ctx context.Context, inactiveThreshold time.Duration) ([]*model.AnalyticsSession, error)
}

type BehaviorEventRepository interface {
	Create(ctx context.Context, e *model.BehaviorEvent) error
	CreateBatch(ctx context.Context, events []*model.BehaviorEvent) (int, error) // returns inserted count
	ListBySession(ctx context.Context, sessionID uint64) ([]*model.BehaviorEvent, error)
	IncrementUserEventCount(ctx context.Context, userID uint64, hourBucket time.Time) error
	GetUserEventCount(ctx context.Context, userID uint64, hourBucket time.Time) (uint32, error)
}

type IdempotencyKeyRepository interface {
	Get(ctx context.Context, keyHash string, userID uint64) (*model.IdempotencyKey, error)
	Create(ctx context.Context, k *model.IdempotencyKey) error
	UpdateResponse(ctx context.Context, keyHash string, userID uint64, code uint16, body string) error
	DeleteExpired(ctx context.Context) (int64, error)
}

type AuditLogRepository interface {
	Create(ctx context.Context, log *model.AuditLog) error
	List(ctx context.Context, page Pagination) ([]*model.AuditLog, int64, error)
}

type IPRuleRepository interface {
	Create(ctx context.Context, cidr string, ruleType string, description string, createdBy uint64) error
	Delete(ctx context.Context, id uint64) error
	ListAll(ctx context.Context) ([]struct {
		CIDR     string
		RuleType string
	}, error)
}
