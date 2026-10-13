package models

import "time"

// ProcessInfo is one line in the "Top processes" table.
type ProcessInfo struct {
	Name     string  `json:"name"`
	CPU      float64 `json:"cpu"`      // percent
	MemoryMB float64 `json:"memoryMb"` // megabytes
}

type Sample struct {
	ServerID  string
	At        time.Time
	CPU       float64       // percent
	Memory    float64       // percent
	Disk      float64       // percent
	Network   float64       // MB/s
	Processes []ProcessInfo // only sent every few seconds, usually empty
}

// HistoryPoint is one averaged point on the history chart.
type HistoryPoint struct {
	T      int64   `json:"t"` // unix ms
	CPU    float64 `json:"cpu"`
	Memory float64 `json:"memory"`
	Disk   float64 `json:"disk"`
}
