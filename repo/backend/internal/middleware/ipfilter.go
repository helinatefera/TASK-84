package middleware

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/localinsights/portal/internal/errs"
	"github.com/localinsights/portal/internal/pkg/database"
)

type ipRule struct {
	CIDR string `db:"cidr"`
	Type string `db:"rule_type"` // "allow" or "deny"
}

type ipRuleCache struct {
	mu          sync.RWMutex
	rules       []ipRule
	lastRefresh time.Time
	ttl         time.Duration
}

func (c *ipRuleCache) needsRefresh() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return time.Since(c.lastRefresh) > c.ttl
}

func (c *ipRuleCache) getRules() []ipRule {
	c.mu.RLock()
	defer c.mu.RUnlock()
	dst := make([]ipRule, len(c.rules))
	copy(dst, c.rules)
	return dst
}

func (c *ipRuleCache) setRules(rules []ipRule) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.rules = rules
	c.lastRefresh = time.Now()
}

// NewIPFilter returns a middleware that checks the client IP against
// ip_rules stored in the database. Deny rules take precedence over allow
// rules. If an allowlist exists and the IP is not in it the request is
// rejected. Rules are cached in memory and refreshed every 60 seconds.
func NewIPFilter(db *database.DB) gin.HandlerFunc {
	cache := &ipRuleCache{
		ttl: 60 * time.Second,
	}

	loadRules := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var rules []ipRule
		err := db.SelectContext(ctx, &rules, "SELECT cidr, rule_type FROM ip_rules")
		if err != nil {
			slog.Error("failed to load IP rules", "error", err)
			return
		}
		cache.setRules(rules)
	}

	// Initial load.
	loadRules()

	return func(c *gin.Context) {
		if cache.needsRefresh() {
			loadRules()
		}

		clientIP := net.ParseIP(c.ClientIP())
		if clientIP == nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    errs.ErrIPDenied.Code,
				"msg": errs.ErrIPDenied.Message,
			})
			return
		}

		rules := cache.getRules()

		// If there are no rules, allow all traffic.
		if len(rules) == 0 {
			c.Next()
			return
		}

		// Check deny rules first — denylist takes precedence.
		for _, rule := range rules {
			if rule.Type != "deny" {
				continue
			}
			_, network, err := net.ParseCIDR(rule.CIDR)
			if err != nil {
				slog.Error("invalid CIDR in ip_rules", "cidr", rule.CIDR, "error", err)
				continue
			}
			if network.Contains(clientIP) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"code":    errs.ErrIPDenied.Code,
					"msg": errs.ErrIPDenied.Message,
				})
				return
			}
		}

		// Determine whether an allowlist exists.
		hasAllowlist := false
		allowed := false
		for _, rule := range rules {
			if rule.Type != "allow" {
				continue
			}
			hasAllowlist = true
			_, network, err := net.ParseCIDR(rule.CIDR)
			if err != nil {
				slog.Error("invalid CIDR in ip_rules", "cidr", rule.CIDR, "error", err)
				continue
			}
			if network.Contains(clientIP) {
				allowed = true
				break
			}
		}

		if hasAllowlist && !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    errs.ErrIPDenied.Code,
				"msg": errs.ErrIPDenied.Message,
			})
			return
		}

		c.Next()
	}
}
