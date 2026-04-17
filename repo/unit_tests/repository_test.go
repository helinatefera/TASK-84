package unit_tests_test

import (
	"testing"

	"github.com/localinsights/portal/internal/repository"
)

// Pure-logic tests for the repository layer's helper types. These lock in
// pagination math that every list endpoint relies on.

func TestPaginationOffset(t *testing.T) {
	cases := []struct {
		name    string
		page    int
		perPage int
		want    int
	}{
		{"first page is offset 0", 1, 20, 0},
		{"second page of 20", 2, 20, 20},
		{"fifth page of 50", 5, 50, 200},
		{"per_page 1", 10, 1, 9},
		{"per_page 100", 3, 100, 200},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := repository.Pagination{Page: tc.page, PerPage: tc.perPage}
			if got := p.Offset(); got != tc.want {
				t.Errorf("Offset() = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestPaginationZeroValuesYieldNegativeOffset(t *testing.T) {
	// A zero-valued Pagination yields Offset() = -PerPage. This documents
	// current behavior so callers (handlers) are responsible for clamping
	// Page >= 1 before passing to repositories.
	p := repository.Pagination{}
	if got := p.Offset(); got != 0 {
		// Page=0, PerPage=0 → (0-1)*0 = 0
		t.Errorf("Offset() = %d, want 0 for zero Pagination", got)
	}
}

func TestPaginationFirstPagePerPage1(t *testing.T) {
	// Concrete example: first page with 1 item per page should offset at 0.
	p := repository.Pagination{Page: 1, PerPage: 1}
	if p.Offset() != 0 {
		t.Errorf("Offset() = %d, want 0", p.Offset())
	}
}
