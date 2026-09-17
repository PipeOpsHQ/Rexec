package models

import "time"

// AdminContainer represents a container with owner info
type AdminContainer struct {
	Container
	Username  string `json:"username"`
	UserEmail string `json:"user_email"`
}

// AdminTerminal represents an active terminal session
type AdminTerminal struct {
	ID          string    `json:"id"`
	ContainerID string    `json:"containerId"`
	Name        string    `json:"name"`
	Status      string    `json:"status"`
	UserID      string    `json:"userId"`
	Username    string    `json:"username"`
	ConnectedAt time.Time `json:"connectedAt"`
}

// AdminUser is the admin-dashboard user row, including container count.
// JSON tags match the frontend admin store (mixed camelCase / snake_case).
type AdminUser struct {
	ID                 string    `json:"id"`
	Email              string    `json:"email"`
	Username           string    `json:"username"`
	Tier               string    `json:"tier"`
	IsAdmin            bool      `json:"isAdmin"`
	SubscriptionActive bool      `json:"subscriptionActive"`
	PipeOpsID          string    `json:"pipeops_id,omitempty"`
	ContainerCount     int       `json:"containerCount"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// AdminUserListParams are pagination/filter options for the admin user list.
type AdminUserListParams struct {
	Page        int
	PerPage     int
	Search      string
	Subscribers bool
}

// Offset returns the SQL OFFSET for the current page.
func (p AdminUserListParams) Offset() int {
	if p.Page < 1 {
		return 0
	}
	return (p.Page - 1) * p.PerPage
}

// TotalPages returns the number of pages needed to cover total rows.
func (p AdminUserListParams) TotalPages(total int) int {
	if p.PerPage <= 0 || total <= 0 {
		return 0
	}
	return (total + p.PerPage - 1) / p.PerPage
}

// AdminUserList is a paginated admin user response.
type AdminUserList struct {
	Users      []*AdminUser `json:"users"`
	Page       int          `json:"page"`
	PerPage    int          `json:"perPage"`
	Total      int          `json:"total"`
	TotalPages int          `json:"totalPages"`
}

// AdminUsageTotals represents all-time and current usage totals for the admin dashboard.
type AdminUsageTotals struct {
	Users          int `json:"users"`
	Subscribers    int `json:"subscribers"`
	Containers     int `json:"containers"`
	ActiveSessions int `json:"activeSessions"`
	Logins         int `json:"logins"`
	Agents         int `json:"agents"`
	OnlineAgents   int `json:"onlineAgents"`
	Recordings     int `json:"recordings"`
	RecordingHours int `json:"recordingHours"`
}

// AdminUsageActivity represents activity within the selected time range.
type AdminUsageActivity struct {
	NewUsers       int `json:"newUsers"`
	NewContainers  int `json:"newContainers"`
	NewSessions    int `json:"newSessions"`
	NewLogins      int `json:"newLogins"`
	NewAgents      int `json:"newAgents"`
	NewRecordings  int `json:"newRecordings"`
	RecordingHours int `json:"recordingHours"`
}

// AdminUsagePoint represents one bucket in the usage chart.
type AdminUsagePoint struct {
	BucketStart   time.Time `json:"bucketStart"`
	BucketLabel   string    `json:"bucketLabel"`
	NewUsers      int       `json:"newUsers"`
	NewContainers int       `json:"newContainers"`
	NewSessions   int       `json:"newSessions"`
	NewLogins     int       `json:"newLogins"`
	NewAgents     int       `json:"newAgents"`
	NewRecordings int       `json:"newRecordings"`
}

// AdminUsageStats represents usage analytics for the admin dashboard.
type AdminUsageStats struct {
	Range    string             `json:"range"`
	Interval string             `json:"interval"`
	From     time.Time          `json:"from"`
	To       time.Time          `json:"to"`
	Totals   AdminUsageTotals   `json:"totals"`
	Activity AdminUsageActivity `json:"activity"`
	Timeline []AdminUsagePoint  `json:"timeline"`
}
