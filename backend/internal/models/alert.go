package models

// AlertRule matches the AlertRule type in the React app.
type AlertRule struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	ServerID    string  `json:"serverId"` // a server id, or "all"
	Metric      string  `json:"metric"`   // cpu | memory | disk
	Threshold   float64 `json:"threshold"`
	DurationMin int     `json:"durationMin"`
	Channel     string  `json:"channel"` // email | slack | webhook
	Enabled     bool    `json:"enabled"`
}

// AlertEvent is one row of the Alerts "History" tab.
type AlertEvent struct {
	ID         string `json:"id"`
	RuleName   string `json:"ruleName"`
	ServerName string `json:"serverName"`
	At         int64  `json:"at"`     // unix ms
	Status     string `json:"status"` // firing | resolved
}

// ServerEvent is one row of "Recent events" on ServerDetail.
type ServerEvent struct {
	ID       string `json:"id"`
	Kind     string `json:"kind"` // alert | ok | error
	Text     string `json:"text"`
	At       int64  `json:"at"`
	Resolved bool   `json:"resolved"`
}
