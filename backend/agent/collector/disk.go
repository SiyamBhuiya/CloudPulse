package collector

import (
	"os"
	"runtime"

	"github.com/shirou/gopsutil/v4/disk"
)

// rootPath is the disk to measure: the system drive on Windows, "/" elsewhere.
func rootPath() string {
	if runtime.GOOS == "windows" {
		if d := os.Getenv("SystemDrive"); d != "" {
			return d + `\`
		}
		return `C:\`
	}
	return "/"
}

func DiskPercent() (float64, error) {
	u, err := disk.Usage(rootPath())
	if err != nil {
		return 0, err
	}
	return u.UsedPercent, nil
}
