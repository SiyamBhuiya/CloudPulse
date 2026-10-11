package collector

import "log/slog"

// Sample is one reading. The JSON names match what the API expects.
type Sample struct {
	CPU     float64 `json:"cpu"`
	Memory  float64 `json:"memory"`
	Disk    float64 `json:"disk"`
	Network float64 `json:"network"`
}

type Collector struct {
	net *NetworkMeter
}

func New() *Collector {
	_, _ = CPUPercent() // the first CPU reading is always 0, so take it now
	return &Collector{net: NewNetworkMeter()}
}

// Collect reads everything. One failing metric is logged and left at 0
// instead of stopping the whole report.
func (c *Collector) Collect() Sample {
	var s Sample
	var err error

	if s.CPU, err = CPUPercent(); err != nil {
		slog.Warn("cpu", "err", err)
	}
	if s.Memory, err = MemoryPercent(); err != nil {
		slog.Warn("memory", "err", err)
	}
	if s.Disk, err = DiskPercent(); err != nil {
		slog.Warn("disk", "err", err)
	}
	if s.Network, err = c.net.MBps(); err != nil {
		slog.Warn("network", "err", err)
	}
	return s
}
