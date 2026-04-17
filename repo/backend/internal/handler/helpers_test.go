package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/localinsights/portal/internal/errs"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// --- respondError ---------------------------------------------------------

func TestRespondErrorWritesStandardJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	respondError(c, http.StatusBadRequest, "ANY_LABEL", "bad input")

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body is not JSON: %v", err)
	}
	// Per the helper contract, the numeric HTTP status is used as "code".
	if code, _ := body["code"].(float64); int(code) != http.StatusBadRequest {
		t.Errorf("code = %v, want 400", body["code"])
	}
	if body["msg"] != "bad input" {
		t.Errorf("msg = %v, want 'bad input'", body["msg"])
	}
}

// --- respondAppError ------------------------------------------------------

func TestRespondAppErrorUsesAppErrorFields(t *testing.T) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	respondAppError(c, errs.ErrForbidden)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", rec.Code)
	}
	var body map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if int(body["code"].(float64)) != http.StatusForbidden {
		t.Errorf("code = %v, want 403", body["code"])
	}
	if body["msg"] != errs.ErrForbidden.Message {
		t.Errorf("msg = %v, want %q", body["msg"], errs.ErrForbidden.Message)
	}
}

func TestRespondAppErrorFallsBackToInternalErrorForPlainError(t *testing.T) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	respondAppError(c, errors.New("not an AppError"))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
	var body map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if int(body["code"].(float64)) != http.StatusInternalServerError {
		t.Errorf("code = %v, want 500", body["code"])
	}
	if body["msg"] == nil || body["msg"] == "" {
		t.Errorf("msg should be non-empty for internal errors")
	}
}

func TestRespondAppErrorUnwrapsWrappedAppError(t *testing.T) {
	wrapped := &wrapErr{inner: errs.ErrValidation}
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	respondAppError(c, wrapped)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", rec.Code)
	}
}

type wrapErr struct{ inner error }

func (w *wrapErr) Error() string { return "wrapped: " + w.inner.Error() }
func (w *wrapErr) Unwrap() error { return w.inner }

// --- getPagination --------------------------------------------------------

func makeCtxWithQuery(query string) (*gin.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/path?"+query, nil)
	return c, rec
}

func TestGetPaginationDefaults(t *testing.T) {
	c, _ := makeCtxWithQuery("")
	p := getPagination(c)
	if p.Page != 1 {
		t.Errorf("Page = %d, want 1", p.Page)
	}
	if p.PerPage != 20 {
		t.Errorf("PerPage = %d, want 20", p.PerPage)
	}
}

func TestGetPaginationHonorsQueryParams(t *testing.T) {
	c, _ := makeCtxWithQuery("page=3&per_page=50")
	p := getPagination(c)
	if p.Page != 3 || p.PerPage != 50 {
		t.Errorf("got page=%d per_page=%d, want page=3 per_page=50", p.Page, p.PerPage)
	}
}

func TestGetPaginationClampsInvalidValues(t *testing.T) {
	cases := []struct {
		query      string
		wantPage   int
		wantPer    int
	}{
		{"page=0&per_page=0", 1, 20},        // zero values → defaults
		{"page=-5&per_page=-10", 1, 20},     // negatives → defaults
		{"page=1&per_page=500", 1, 20},      // per_page > 100 → fallback
		{"page=abc&per_page=xyz", 1, 20},    // non-numeric → defaults
		{"page=1&per_page=100", 1, 100},     // exactly 100 allowed
	}
	for _, tc := range cases {
		t.Run(tc.query, func(t *testing.T) {
			c, _ := makeCtxWithQuery(tc.query)
			p := getPagination(c)
			if p.Page != tc.wantPage || p.PerPage != tc.wantPer {
				t.Errorf("got page=%d per_page=%d, want page=%d per_page=%d",
					p.Page, p.PerPage, tc.wantPage, tc.wantPer)
			}
		})
	}
}

// --- paginatedResponse ----------------------------------------------------

func TestPaginatedResponseShape(t *testing.T) {
	data := []string{"a", "b", "c"}
	p := getPagination(func() *gin.Context { c, _ := makeCtxWithQuery("page=2&per_page=10"); return c }())
	resp := paginatedResponse(data, p, 25)

	if resp["data"] == nil {
		t.Error("data should not be nil")
	}
	if resp["page"] != 2 {
		t.Errorf("page = %v, want 2", resp["page"])
	}
	if resp["per_page"] != 10 {
		t.Errorf("per_page = %v, want 10", resp["per_page"])
	}
	if resp["total"] != int64(25) {
		t.Errorf("total = %v, want 25", resp["total"])
	}
	// 25 items, 10 per page → 3 pages.
	if resp["total_pages"] != int64(3) {
		t.Errorf("total_pages = %v, want 3", resp["total_pages"])
	}
}

func TestPaginatedResponseExactMultiple(t *testing.T) {
	p := getPagination(func() *gin.Context { c, _ := makeCtxWithQuery("page=1&per_page=10"); return c }())
	resp := paginatedResponse([]int{}, p, 20)
	if resp["total_pages"] != int64(2) {
		t.Errorf("total_pages = %v, want 2 (20/10 exact)", resp["total_pages"])
	}
}

// --- parseUintParam -------------------------------------------------------

func TestParseUintParamAcceptsValidInt(t *testing.T) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Params = gin.Params{{Key: "id", Value: "42"}}
	id, ok := parseUintParam(c, "id")
	if !ok {
		t.Fatal("parseUintParam returned ok=false for valid input")
	}
	if id != 42 {
		t.Errorf("id = %d, want 42", id)
	}
	if rec.Code != 0 && rec.Code != 200 {
		t.Errorf("should not write error response when input is valid, status=%d", rec.Code)
	}
}

func TestParseUintParamRejectsNonNumeric(t *testing.T) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Params = gin.Params{{Key: "id", Value: "abc"}}
	_, ok := parseUintParam(c, "id")
	if ok {
		t.Fatal("parseUintParam returned ok=true for non-numeric input")
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
	var body map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	msg, _ := body["msg"].(string)
	if msg == "" || !contains(msg, "id") {
		t.Errorf("msg = %q, should mention parameter name", msg)
	}
}

func TestParseUintParamRejectsNegative(t *testing.T) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Params = gin.Params{{Key: "item_id", Value: "-5"}}
	_, ok := parseUintParam(c, "item_id")
	if ok {
		t.Fatal("parseUintParam accepted a negative number for uint")
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

// contains is a tiny substring check so we don't drag in strings imports
// where tests are otherwise self-contained.
func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
