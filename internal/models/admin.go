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
