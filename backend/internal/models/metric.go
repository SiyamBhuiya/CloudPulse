package models

import "time"

type Sample struct {
	ServerID string
	At       time.Time
	CPU      float64 // percent
	Memory   float64 // percent
	Disk     float64 // percent
	Network  float64 // MB/s
}
