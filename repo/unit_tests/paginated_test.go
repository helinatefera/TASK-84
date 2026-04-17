package unit_tests_test

import (
	"testing"

	"github.com/localinsights/portal/internal/dto/response"
)

func TestNewPaginated(t *testing.T) {
	cases := []struct {
		name           string
		page           int
		perPage        int
		total          int64
		wantTotalPages int64
	}{
		{"empty result set", 1, 20, 0, 0},
		{"exact multiple of page size", 1, 20, 100, 5},
		{"partial last page", 1, 20, 101, 6},
		{"single item", 1, 20, 1, 1},
		{"per_page larger than total", 1, 50, 3, 1},
		{"perPage=1 per-item paging", 1, 1, 10, 10},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := []string{"a", "b"}
			got := response.NewPaginated(data, tc.page, tc.perPage, tc.total)

			if got.Page != tc.page {
				t.Errorf("Page = %d, want %d", got.Page, tc.page)
			}
			if got.PerPage != tc.perPage {
				t.Errorf("PerPage = %d, want %d", got.PerPage, tc.perPage)
			}
			if got.Total != tc.total {
				t.Errorf("Total = %d, want %d", got.Total, tc.total)
			}
			if got.TotalPages != tc.wantTotalPages {
				t.Errorf("TotalPages = %d, want %d", got.TotalPages, tc.wantTotalPages)
			}
			if got.Data == nil {
				t.Errorf("Data should not be nil")
			}
		})
	}
}
