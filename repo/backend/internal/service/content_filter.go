package service

import (
	"context"
	"regexp"

	"github.com/localinsights/portal/internal/repository"
)

// ContentFilterResult describes the outcome of running text through the
// sensitive-word rule set.
type ContentFilterResult struct {
	Blocked bool   // true if any "block" rule matched
	Flagged bool   // true if any "flag" rule matched
	Text    string // the (possibly replaced) text
}

// ContentFilter checks user-submitted text against active sensitive word rules.
type ContentFilter struct {
	repo repository.SensitiveWordRuleRepository
}

func NewContentFilter(repo repository.SensitiveWordRuleRepository) *ContentFilter {
	return &ContentFilter{repo: repo}
}

// Check loads active rules and applies them to the input text.
func (f *ContentFilter) Check(ctx context.Context, text string) (ContentFilterResult, error) {
	rules, err := f.repo.ListActive(ctx)
	if err != nil {
		return ContentFilterResult{Text: text}, err
	}

	result := ContentFilterResult{Text: text}
	for _, rule := range rules {
		re, err := regexp.Compile("(?i)" + rule.Pattern)
		if err != nil {
			continue // skip invalid patterns
		}
		if !re.MatchString(result.Text) {
			continue
		}

		switch rule.Action {
		case "block":
			result.Blocked = true
			return result, nil
		case "flag":
			result.Flagged = true
		case "replace":
			replacement := "***"
			if rule.Replacement != nil && *rule.Replacement != "" {
				replacement = *rule.Replacement
			}
			result.Text = re.ReplaceAllString(result.Text, replacement)
		}
	}

	return result, nil
}

// CheckOrBlock is a convenience that returns an error message if blocked, or
// the (possibly replaced) text if allowed.
func (f *ContentFilter) CheckOrBlock(ctx context.Context, text string) (filtered string, blocked bool, err error) {
	res, err := f.Check(ctx, text)
	if err != nil {
		return text, false, err
	}
	return res.Text, res.Blocked, nil
}

// Apply runs content filter and returns cleaned text plus whether the content
// should be flagged for moderator review. Block-action matches return a non-empty
// reason string.
func (f *ContentFilter) Apply(ctx context.Context, text string) (cleaned string, blockReason string, flagged bool) {
	if text == "" {
		return text, "", false
	}
	res, err := f.Check(ctx, text)
	if err != nil {
		// On error, allow through but don't silently drop the filter.
		return text, "", false
	}
	if res.Blocked {
		return text, "Content contains prohibited words", false
	}
	return res.Text, "", res.Flagged
}

