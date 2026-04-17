package unit_tests_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/localinsights/portal/internal/middleware"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// --- SecurityHeaders ------------------------------------------------------

func TestSecurityHeadersSetsAllExpectedHeaders(t *testing.T) {
	r := gin.New()
	r.Use(middleware.SecurityHeaders())
	r.GET("/x", func(c *gin.Context) { c.String(200, "ok") })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	r.ServeHTTP(rec, req)

	checks := map[string]string{
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":           "DENY",
		"X-XSS-Protection":          "1; mode=block",
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
		"Referrer-Policy":           "strict-origin-when-cross-origin",
	}
	for h, want := range checks {
		if got := rec.Header().Get(h); got != want {
			t.Errorf("%s = %q, want %q", h, got, want)
		}
	}
	if csp := rec.Header().Get("Content-Security-Policy"); csp == "" {
		t.Error("Content-Security-Policy header is missing")
	}
}

// --- RequestID ------------------------------------------------------------

func TestRequestIDGeneratesUUIDWhenAbsent(t *testing.T) {
	r := gin.New()
	r.Use(middleware.RequestID())
	var ctxID string
	r.GET("/x", func(c *gin.Context) {
		v, _ := c.Get("request_id")
		ctxID, _ = v.(string)
		c.String(200, "ok")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	r.ServeHTTP(rec, req)

	got := rec.Header().Get("X-Request-ID")
	if got == "" {
		t.Fatal("X-Request-ID header was not set")
	}
	if got != ctxID {
		t.Errorf("context id %q != response header %q", ctxID, got)
	}
	if len(got) < 10 {
		t.Errorf("generated request id looks too short: %q", got)
	}
}

func TestRequestIDHonorsIncomingHeader(t *testing.T) {
	r := gin.New()
	r.Use(middleware.RequestID())
	r.GET("/x", func(c *gin.Context) { c.String(200, "ok") })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("X-Request-ID", "client-supplied-123")
	r.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Request-ID"); got != "client-supplied-123" {
		t.Errorf("X-Request-ID = %q, want client-supplied-123", got)
	}
}

// --- Recovery -------------------------------------------------------------

func TestRecoveryCatchesPanicsAndReturns500(t *testing.T) {
	r := gin.New()
	r.Use(middleware.Recovery())
	r.GET("/boom", func(c *gin.Context) { panic("kaboom") })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/boom", nil)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
	body := rec.Body.String()
	// Response must include our {code, msg} contract — not a stack trace.
	if !containsAll(body, `"code"`, `"msg"`, "INTERNAL_ERROR") {
		t.Errorf("body missing expected fields: %s", body)
	}
	if containsAny(body, "runtime.gopanic", "goroutine ") {
		t.Errorf("body leaked stack trace: %s", body)
	}
}

func TestRecoveryAllowsNormalResponsesThrough(t *testing.T) {
	r := gin.New()
	r.Use(middleware.Recovery())
	r.GET("/ok", func(c *gin.Context) { c.String(200, "fine") })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	if rec.Body.String() != "fine" {
		t.Errorf("body = %q, want %q", rec.Body.String(), "fine")
	}
}

// --- RBAC: RequireAuth + RequireRole -------------------------------------

func TestRequireAuthRejectsAnonymous(t *testing.T) {
	r := gin.New()
	r.Use(middleware.RequireAuth())
	r.GET("/protected", func(c *gin.Context) { c.String(200, "ok") })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestRequireAuthAllowsAuthenticatedRequest(t *testing.T) {
	r := gin.New()
	// Simulate that an upstream middleware set the auth context.
	r.Use(func(c *gin.Context) {
		c.Set("authenticated", true)
		c.Set("user_role", "regular_user")
		c.Next()
	})
	r.Use(middleware.RequireAuth())
	r.GET("/protected", func(c *gin.Context) { c.String(200, "ok") })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	r.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestRequireRoleReturns403WhenRoleMismatch(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("authenticated", true)
		c.Set("user_role", "regular_user")
		c.Next()
	})
	r.Use(middleware.RequireRole("admin", "moderator"))
	r.GET("/admin-only", func(c *gin.Context) { c.String(200, "ok") })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin-only", nil)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", rec.Code)
	}
}

func TestRequireRoleAllowsWhenRoleMatchesAny(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("authenticated", true)
		c.Set("user_role", "moderator")
		c.Next()
	})
	r.Use(middleware.RequireRole("admin", "moderator"))
	r.GET("/modify", func(c *gin.Context) { c.String(200, "ok") })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/modify", nil)
	r.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("status = %d, want 200 (moderator should be allowed), body=%s", rec.Code, rec.Body.String())
	}
}

func TestRequireRoleReturns401WhenUnauthenticated(t *testing.T) {
	r := gin.New()
	r.Use(middleware.RequireRole("admin"))
	r.GET("/a", func(c *gin.Context) { c.String(200, "ok") })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/a", nil)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

// --- CSRF -----------------------------------------------------------------

func TestCSRFDisabledIsNoOp(t *testing.T) {
	r := gin.New()
	r.Use(middleware.CSRF(false))
	r.POST("/x", func(c *gin.Context) { c.String(200, "ok") })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/x", nil)
	r.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("status = %d, want 200 (CSRF disabled)", rec.Code)
	}
}

func TestCSRFGenerateEndpointReturnsTokenAndCookie(t *testing.T) {
	r := gin.New()
	r.Use(middleware.CSRF(true))
	// The handler would never run because the middleware aborts for GET /api/v1/csrf,
	// but we register one anyway to match production routing shape.
	r.GET("/api/v1/csrf", func(c *gin.Context) { c.String(200, "should-be-aborted") })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/csrf", nil)
	r.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !containsAll(body, `"csrf_token"`) {
		t.Errorf("body missing csrf_token field: %s", body)
	}
	// Set-Cookie must be present and name the csrf_token cookie.
	sc := rec.Header().Get("Set-Cookie")
	if sc == "" || !containsAll(sc, "csrf_token=") {
		t.Errorf("Set-Cookie header did not set csrf_token: %q", sc)
	}
	// Must not leak the stub handler output.
	if containsAny(body, "should-be-aborted") {
		t.Errorf("middleware did not abort the downstream handler: %s", body)
	}
}

func TestCSRFRejectsMutatingRequestWithoutToken(t *testing.T) {
	r := gin.New()
	r.Use(middleware.CSRF(true))
	r.POST("/api/v1/anything", func(c *gin.Context) { c.String(200, "ok") })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/anything", nil)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", rec.Code)
	}
	if !containsAll(rec.Body.String(), "CSRF_INVALID") {
		t.Errorf("expected CSRF_INVALID in body, got %s", rec.Body.String())
	}
}

func TestCSRFRejectsMutatingRequestWithMismatchedHeaderAndCookie(t *testing.T) {
	r := gin.New()
	r.Use(middleware.CSRF(true))
	r.POST("/api/v1/anything", func(c *gin.Context) { c.String(200, "ok") })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/anything", nil)
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "cookie-value"})
	req.Header.Set("X-CSRF-Token", "different-value")
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", rec.Code)
	}
}

func TestCSRFAcceptsMutatingRequestWithMatchingHeaderAndCookie(t *testing.T) {
	r := gin.New()
	r.Use(middleware.CSRF(true))
	r.POST("/api/v1/anything", func(c *gin.Context) { c.String(200, "ok") })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/anything", nil)
	token := "matching-hex-value-1234"
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: token})
	req.Header.Set("X-CSRF-Token", token)
	r.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("status = %d, want 200 (CSRF matched), body=%s", rec.Code, rec.Body.String())
	}
}

func TestCSRFAllowsNonMutatingRequestsWithoutToken(t *testing.T) {
	r := gin.New()
	r.Use(middleware.CSRF(true))
	r.GET("/api/v1/items", func(c *gin.Context) { c.String(200, "ok") })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/items", nil)
	r.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("status = %d, want 200 (GET should bypass CSRF validation)", rec.Code)
	}
}

// --- Helpers --------------------------------------------------------------

func containsAll(s string, subs ...string) bool {
	for _, sub := range subs {
		if !containsString(s, sub) {
			return false
		}
	}
	return true
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if containsString(s, sub) {
			return true
		}
	}
	return false
}

func containsString(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
