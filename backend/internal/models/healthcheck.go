package models

// HealthCheck matches the HealthCheck type in the React app.
type HealthCheck struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	URL            string   `json:"url"`
	Status         string   `json:"status"`     // pending | up | slow | down
	ResponseMs     *int     `json:"responseMs"` // null = no response
	Uptime         float64  `json:"uptime"`     // 30-day percent
	IntervalSec    int      `json:"intervalSec"`
	ExpectedStatus int      `json:"expectedStatus"`
	LastChecked    int64    `json:"lastChecked"` // unix ms
	Enabled        bool     `json:"enabled"`
	Days           []string `json:"days"` // 30 entries, oldest first: ok | degraded | down | none
}
