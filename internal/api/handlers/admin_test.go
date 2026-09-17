package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rexec/rexec/internal/models"
)

func TestParseAdminListQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name string
		raw  string
		want models.AdminUserListParams
	}{
		{
			name: "defaults",
			raw:  "",
			want: models.AdminUserListParams{Page: 1, PerPage: 25},
		},
		{
			name: "page and per_page",
			raw:  "page=3&per_page=50",
			want: models.AdminUserListParams{Page: 3, PerPage: 50},
		},
		{
			name: "perPage camelCase alias",
			raw:  "page=2&perPage=10",
			want: models.AdminUserListParams{Page: 2, PerPage: 10},
		},
		{
			name: "caps per_page",
			raw:  "per_page=999",
			want: models.AdminUserListParams{Page: 1, PerPage: 100},
		},
		{
			name: "caps page",
			raw:  "page=99999",
			want: models.AdminUserListParams{Page: 10000, PerPage: 25},
		},
		{
			name: "negative page and per_page",
			raw:  "page=-2&per_page=0",
			want: models.AdminUserListParams{Page: 1, PerPage: 25},
		},
		{
			name: "search trimmed",
			raw:  "search=%20alice%20",
			want: models.AdminUserListParams{Page: 1, PerPage: 25, Search: "alice"},
		},
		{
			name: "search alias q",
			raw:  "q=bob",
			want: models.AdminUserListParams{Page: 1, PerPage: 25, Search: "bob"},
		},
		{
			name: "subscribers true",
			raw:  "subscribers=true",
			want: models.AdminUserListParams{Page: 1, PerPage: 25, Subscribers: true},
		},
		{
			name: "subscribers 1",
			raw:  "subscribers=1",
			want: models.AdminUserListParams{Page: 1, PerPage: 25, Subscribers: true},
		},
		{
			name: "subscribers false",
			raw:  "subscribers=false",
			want: models.AdminUserListParams{Page: 1, PerPage: 25, Subscribers: false},
		},
		{
			name: "search truncated at 100",
			raw:  "search=" + strings.Repeat("a", 120),
			want: models.AdminUserListParams{Page: 1, PerPage: 25, Search: strings.Repeat("a", 100)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req := httptest.NewRequest(http.MethodGet, "/api/admin/users?"+tt.raw, nil)
			c.Request = req

			got := parseAdminListQuery(c)
			if got != tt.want {
				t.Fatalf("parseAdminListQuery(%q) = %+v, want %+v", tt.raw, got, tt.want)
			}
		})
	}
}

func TestAdminUserListParamsOffsetAndPages(t *testing.T) {
	t.Parallel()

	params := models.AdminUserListParams{Page: 3, PerPage: 25}
	if got := params.Offset(); got != 50 {
		t.Fatalf("Offset() = %d, want 50", got)
	}
	if got := params.TotalPages(51); got != 3 {
		t.Fatalf("TotalPages(51) = %d, want 3", got)
	}
	if got := params.TotalPages(0); got != 0 {
		t.Fatalf("TotalPages(0) = %d, want 0", got)
	}

	zeroPage := models.AdminUserListParams{Page: 0, PerPage: 25}
	if got := zeroPage.Offset(); got != 0 {
		t.Fatalf("Offset() for page 0 = %d, want 0", got)
	}
}

func TestResolveAdminUsageRange(t *testing.T) {
	t.Parallel()

	tests := []struct {
		key      string
		interval string
		ok       bool
	}{
		{"24h", "hour", true},
		{"7d", "day", true},
		{"30d", "day", true},
		{"90d", "week", true},
		{"12m", "month", true},
		{"", "", false},
		{"year", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			from, to, interval, ok := resolveAdminUsageRange(tt.key)
			if ok != tt.ok {
				t.Fatalf("ok = %v, want %v", ok, tt.ok)
			}
			if !tt.ok {
				return
			}
			if interval != tt.interval {
				t.Fatalf("interval = %q, want %q", interval, tt.interval)
			}
			if !from.Before(to) {
				t.Fatalf("from %v should be before to %v", from, to)
			}
		})
	}
}
