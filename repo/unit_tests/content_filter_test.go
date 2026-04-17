package unit_tests_test

import (
	"context"
	"errors"
	"testing"

	"github.com/localinsights/portal/internal/model"
	"github.com/localinsights/portal/internal/service"
)

// --- Fake repository that implements the SensitiveWordRuleRepository
// interface (only ListActive is used by ContentFilter). We stub the
// unused methods to satisfy the interface at compile time.

type fakeWordRuleRepo struct {
	rules []*model.SensitiveWordRule
	err   error
	calls int
}

func (f *fakeWordRuleRepo) ListActive(ctx context.Context) ([]*model.SensitiveWordRule, error) {
	f.calls++
	return f.rules, f.err
}
func (f *fakeWordRuleRepo) Create(ctx context.Context, r *model.SensitiveWordRule) error {
	return nil
}
func (f *fakeWordRuleRepo) GetByID(ctx context.Context, id uint64) (*model.SensitiveWordRule, error) {
	return nil, nil
}
func (f *fakeWordRuleRepo) Update(ctx context.Context, r *model.SensitiveWordRule) error {
	return nil
}
func (f *fakeWordRuleRepo) Delete(ctx context.Context, id uint64) error { return nil }

func strPtr(s string) *string { return &s }

func TestContentFilterNoRulesLeavesTextUntouched(t *testing.T) {
	repo := &fakeWordRuleRepo{rules: nil}
	cf := service.NewContentFilter(repo)

	res, err := cf.Check(context.Background(), "hello world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Text != "hello world" {
		t.Errorf("Text = %q, want %q", res.Text, "hello world")
	}
	if res.Blocked || res.Flagged {
		t.Errorf("unexpected flags: blocked=%v flagged=%v", res.Blocked, res.Flagged)
	}
}

func TestContentFilterBlocksOnBlockAction(t *testing.T) {
	repo := &fakeWordRuleRepo{rules: []*model.SensitiveWordRule{
		{Pattern: "badword", Action: "block"},
	}}
	cf := service.NewContentFilter(repo)

	res, err := cf.Check(context.Background(), "This has a BadWord in it")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Blocked {
		t.Error("expected Blocked=true")
	}
	// Short-circuit: further rules are not applied after a block.
	if res.Flagged {
		t.Error("should not be flagged after block")
	}
}

func TestContentFilterFlagsOnFlagAction(t *testing.T) {
	repo := &fakeWordRuleRepo{rules: []*model.SensitiveWordRule{
		{Pattern: "suspicious", Action: "flag"},
	}}
	cf := service.NewContentFilter(repo)

	res, err := cf.Check(context.Background(), "this looks Suspicious to me")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Blocked {
		t.Error("should not be blocked on flag action")
	}
	if !res.Flagged {
		t.Error("expected Flagged=true")
	}
	if res.Text != "this looks Suspicious to me" {
		t.Errorf("Text should be unchanged for flag action, got %q", res.Text)
	}
}

func TestContentFilterReplacesTextOnReplaceAction(t *testing.T) {
	repo := &fakeWordRuleRepo{rules: []*model.SensitiveWordRule{
		{Pattern: "darn", Action: "replace", Replacement: strPtr("****")},
	}}
	cf := service.NewContentFilter(repo)

	res, err := cf.Check(context.Background(), "Oh darn, that darn thing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Text != "Oh ****, that **** thing" {
		t.Errorf("Text = %q, want all occurrences replaced", res.Text)
	}
}

func TestContentFilterReplaceWithNilReplacementDefaultsToStars(t *testing.T) {
	repo := &fakeWordRuleRepo{rules: []*model.SensitiveWordRule{
		{Pattern: "darn", Action: "replace", Replacement: nil},
	}}
	cf := service.NewContentFilter(repo)

	res, _ := cf.Check(context.Background(), "oh darn")
	if res.Text != "oh ***" {
		t.Errorf("Text = %q, want default '***' replacement", res.Text)
	}
}

func TestContentFilterSkipsInvalidRegex(t *testing.T) {
	// "[invalid" is not a valid regex; the filter must not crash and must
	// continue evaluating subsequent rules.
	repo := &fakeWordRuleRepo{rules: []*model.SensitiveWordRule{
		{Pattern: "[invalid", Action: "block"},
		{Pattern: "hello", Action: "flag"},
	}}
	cf := service.NewContentFilter(repo)

	res, err := cf.Check(context.Background(), "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Blocked {
		t.Error("invalid regex should not block")
	}
	if !res.Flagged {
		t.Error("subsequent rule should still fire")
	}
}

func TestContentFilterPropagatesRepoError(t *testing.T) {
	repo := &fakeWordRuleRepo{err: errors.New("db down")}
	cf := service.NewContentFilter(repo)
	res, err := cf.Check(context.Background(), "x")
	if err == nil {
		t.Fatal("expected error from repo to propagate")
	}
	if res.Text != "x" {
		t.Errorf("Text should echo input when error occurs, got %q", res.Text)
	}
}

// --- CheckOrBlock -----------------------------------------------------------

func TestCheckOrBlockReturnsBlockedFlag(t *testing.T) {
	repo := &fakeWordRuleRepo{rules: []*model.SensitiveWordRule{
		{Pattern: "badword", Action: "block"},
	}}
	cf := service.NewContentFilter(repo)

	_, blocked, err := cf.CheckOrBlock(context.Background(), "contains badword")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !blocked {
		t.Error("expected blocked=true")
	}
}

func TestCheckOrBlockReturnsFilteredTextWhenReplaced(t *testing.T) {
	repo := &fakeWordRuleRepo{rules: []*model.SensitiveWordRule{
		{Pattern: "darn", Action: "replace", Replacement: strPtr("****")},
	}}
	cf := service.NewContentFilter(repo)

	filtered, blocked, err := cf.CheckOrBlock(context.Background(), "oh darn")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if blocked {
		t.Error("replace should not count as blocked")
	}
	if filtered != "oh ****" {
		t.Errorf("filtered = %q, want %q", filtered, "oh ****")
	}
}

// --- Apply ------------------------------------------------------------------

func TestApplyEmptyStringShortCircuits(t *testing.T) {
	repo := &fakeWordRuleRepo{rules: []*model.SensitiveWordRule{
		{Pattern: "bad", Action: "block"},
	}}
	cf := service.NewContentFilter(repo)

	cleaned, reason, flagged := cf.Apply(context.Background(), "")
	if cleaned != "" || reason != "" || flagged {
		t.Errorf("empty input should return zero values, got (%q, %q, %v)", cleaned, reason, flagged)
	}
	if repo.calls != 0 {
		t.Errorf("empty input must not hit the repo; got %d calls", repo.calls)
	}
}

func TestApplyReturnsBlockReasonOnBlock(t *testing.T) {
	repo := &fakeWordRuleRepo{rules: []*model.SensitiveWordRule{
		{Pattern: "bad", Action: "block"},
	}}
	cf := service.NewContentFilter(repo)

	_, reason, flagged := cf.Apply(context.Background(), "this is bad")
	if reason == "" {
		t.Error("expected a non-empty block reason")
	}
	if flagged {
		t.Error("block should not also flag")
	}
}

func TestApplyPassesThroughOnRepoError(t *testing.T) {
	repo := &fakeWordRuleRepo{err: errors.New("db down")}
	cf := service.NewContentFilter(repo)

	cleaned, reason, flagged := cf.Apply(context.Background(), "anything")
	if cleaned != "anything" {
		t.Errorf("should pass text through on error, got %q", cleaned)
	}
	if reason != "" || flagged {
		t.Errorf("should not block or flag on repo error; got reason=%q flagged=%v", reason, flagged)
	}
}
