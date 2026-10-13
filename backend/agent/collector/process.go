package collector

import (
	"runtime"
	"sort"

	"github.com/shirou/gopsutil/v4/process"
)

// ProcessInfo is one line in the "Top processes" table. The JSON names match the API.
type ProcessInfo struct {
	Name     string  `json:"name"`
	CPU      float64 `json:"cpu"`
	MemoryMB float64 `json:"memoryMb"`
}

// ProcessMeter keeps process handles between readings. CPU use is measured
// since the previous reading, so a handle has to live across calls.
type ProcessMeter struct {
	procs map[int32]*process.Process
	cores float64
}

func NewProcessMeter() *ProcessMeter {
	return &ProcessMeter{
		procs: make(map[int32]*process.Process),
		cores: float64(runtime.NumCPU()),
	}
}

// Top returns the n busiest programs. The first call only primes the meter
// and returns nothing.
func (m *ProcessMeter) Top(n int) []ProcessInfo {
	pids, err := process.Pids()
	if err != nil {
		return nil
	}

	first := len(m.procs) == 0
	alive := make(map[int32]struct{}, len(pids))
	byName := make(map[string]*ProcessInfo)

	for _, pid := range pids {
		if pid == 0 { // the "System Idle Process" on Windows
			continue
		}
		alive[pid] = struct{}{}

		p, known := m.procs[pid]
		if !known {
			np, err := process.NewProcess(pid)
			if err != nil {
				continue
			}
			p = np
			m.procs[pid] = p
		}

		cpu, err := p.Percent(0) // since the last call for this process
		if err != nil {
			continue
		}
		if !known {
			cpu = 0 // first sight has no previous reading to compare against
		}
		cpu /= m.cores // gopsutil counts 100% per core; Task Manager counts the whole machine

		name, err := p.Name()
		if err != nil || name == "" {
			continue
		}
		var mb float64
		if mi, err := p.MemoryInfo(); err == nil && mi != nil {
			mb = float64(mi.RSS) / 1024 / 1024
		}

		agg := byName[name]
		if agg == nil {
			agg = &ProcessInfo{Name: name}
			byName[name] = agg
		}
		agg.CPU += cpu
		agg.MemoryMB += mb
	}

	// forget processes that have exited
	for pid := range m.procs {
		if _, ok := alive[pid]; !ok {
			delete(m.procs, pid)
		}
	}

	if first {
		return nil
	}

	list := make([]ProcessInfo, 0, len(byName))
	for _, p := range byName {
		if p.CPU > 100 {
			p.CPU = 100
		}
		list = append(list, *p)
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].CPU != list[j].CPU {
			return list[i].CPU > list[j].CPU
		}
		return list[i].MemoryMB > list[j].MemoryMB
	})
	if len(list) > n {
		list = list[:n]
	}
	return list
}
