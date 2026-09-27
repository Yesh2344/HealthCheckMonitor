package types

import "time"

// ServiceStatus holds the runtime health information of a monitored service.
type ServiceStatus struct {
	Name        string    `json:"name"`
	URL         string    `json:"url"`
	Healthy     bool      `json:"healthy"`
	LastChecked time.Time `json:"last_checked"`
	Error       string    `json:"error,omitempty"`
}