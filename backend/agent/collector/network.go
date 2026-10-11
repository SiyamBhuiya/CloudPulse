package collector

import (
	"time"

	"github.com/shirou/gopsutil/v4/net"
)

// NetworkMeter turns the ever-growing byte counters into a speed.
type NetworkMeter struct {
	last uint64
	at   time.Time
}

func NewNetworkMeter() *NetworkMeter {
	m := &NetworkMeter{}
	_, _ = m.MBps() // the first reading only sets the starting point
	return m
}

// MBps is traffic in plus out since the previous call, in megabytes per second.
func (m *NetworkMeter) MBps() (float64, error) {
	counters, err := net.IOCounters(false) // false = all interfaces combined
	if err != nil || len(counters) == 0 {
		return 0, err
	}
	total := counters[0].BytesSent + counters[0].BytesRecv
	now := time.Now()

	var rate float64
	if !m.at.IsZero() && total >= m.last {
		if dt := now.Sub(m.at).Seconds(); dt > 0 {
			rate = float64(total-m.last) / dt / 1e6
		}
	}
	m.last, m.at = total, now
	return rate, nil
}
