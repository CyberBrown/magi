package disk

import (
	"os"
	"strings"
	"syscall"

	"github.com/cyberbrown/magi/types"
)

// interestingMounts are mount points we want to report on.
var interestingMounts = []string{"/", "/home", "/data", "/mnt/data"}

// Collect reads filesystem usage for key mount points.
func Collect() (*types.DiskStatus, error) {
	mounts, err := readMounts()
	if err != nil {
		return nil, err
	}

	var result []types.DiskMount
	seen := make(map[string]bool)

	for _, m := range mounts {
		if seen[m.MountPoint] {
			continue
		}
		if !isInteresting(m.MountPoint, m.Filesystem) {
			continue
		}

		var stat syscall.Statfs_t
		if err := syscall.Statfs(m.MountPoint, &stat); err != nil {
			continue
		}

		total := stat.Blocks * uint64(stat.Bsize)
		free := stat.Bavail * uint64(stat.Bsize)
		used := total - (stat.Bfree * uint64(stat.Bsize))

		var pct float64
		if total > 0 {
			pct = float64(used) / float64(total) * 100
		}

		result = append(result, types.DiskMount{
			MountPoint:   m.MountPoint,
			Filesystem:   m.Filesystem,
			Total:        total,
			Used:         used,
			Free:         free,
			UsagePercent: pct,
		})
		seen[m.MountPoint] = true
	}

	return &types.DiskStatus{Mounts: result}, nil
}

type mount struct {
	Filesystem string
	MountPoint string
}

func readMounts() ([]mount, error) {
	data, err := os.ReadFile("/proc/mounts")
	if err != nil {
		return nil, err
	}

	var mounts []mount
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		mounts = append(mounts, mount{
			Filesystem: fields[0],
			MountPoint: fields[1],
		})
	}
	return mounts, nil
}

func isInteresting(mountPoint, fs string) bool {
	// Skip snap/loop mounts
	if strings.HasPrefix(mountPoint, "/snap/") || strings.Contains(fs, "loop") {
		return false
	}
	for _, m := range interestingMounts {
		if mountPoint == m {
			return true
		}
	}
	if strings.HasPrefix(fs, "/dev/") {
		return true
	}
	return false
}
