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

// AdminUsageTotals represents all-time and current usage totals for the admin dashboard.
type AdminUsageTotals struct {
	Users          int `json:"users"`
	Containers     int `json:"containers"`
	ActiveSessions int `json:"activeSessions"`
	Agents         int `json:"agents"`
	OnlineAgents   int `json:"onlineAgents"`
}

// AdminUsageActivity represents activity within the selected time range.
type AdminUsageActivity struct {
	NewUsers      int `json:"newUsers"`
	NewContainers int `json:"newContainers"`
	NewSessions   int `json:"newSessions"`
	NewAgents     int `json:"newAgents"`
}

// AdminUsagePoint represents one bucket in the usage chart.
type AdminUsagePoint struct {
	BucketStart   time.Time `json:"bucketStart"`
	BucketLabel   string    `json:"bucketLabel"`
	NewUsers      int       `json:"newUsers"`
	NewContainers int       `json:"newContainers"`
	NewSessions   int       `json:"newSessions"`
	NewAgents     int       `json:"newAgents"`
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
