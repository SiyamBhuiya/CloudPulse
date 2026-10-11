package collector

import "github.com/shirou/gopsutil/v4/cpu"

// CPUPercent is the average load across all cores since the previous call.
func CPUPercent() (float64, error) {
	p, err := cpu.Percent(0, false)
	if err != nil || len(p) == 0 {
		return 0, err
	}
	return p[0], nil
}
