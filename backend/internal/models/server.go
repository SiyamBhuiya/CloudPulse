package models

import "time"

type Server struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Region   string  `json:"region"`
	Status   string  `json:"status"` // online | warning | down
	CPU      float64 `json:"cpu"`
	Memory   float64 `json:"memory"`
	Uptime   float64 `json:"uptime"`
	LastSeen int64   `json:"lastSeen"` // unix ms

	// Database values the browser does not need
	LastSeenAt *time.Time `json:"-"`
	CreatedAt  time.Time  `json:"-"`
}