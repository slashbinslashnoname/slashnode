// Package system reports lightweight host indicators (disk, memory, load) so the
// UI can surface node health — in particular a warning when storage runs low,
// since Docker images/volumes fill the disk.
package system

import (
	"os"
	"strconv"
	"strings"
	"syscall"
)

// Disk describes filesystem usage for a path.
type Disk struct {
	Path    string `json:"path"`
	Total   uint64 `json:"total"`
	Used    uint64 `json:"used"`
	Free    uint64 `json:"free"`
	Percent int    `json:"percent"`
}

// Stats is the collected host health snapshot. Fields that can't be read on a
// given platform are left zero.
type Stats struct {
	Disk       Disk    `json:"disk"`
	MemTotal   uint64  `json:"mem_total"`
	MemUsed    uint64  `json:"mem_used"`
	MemPercent int     `json:"mem_percent"`
	Load1      float64 `json:"load1"`
	// DiskWarn/DiskCritical are convenience flags for the UI.
	DiskWarn     bool `json:"disk_warn"`
	DiskCritical bool `json:"disk_critical"`
	// GPUAvailable is true when an NVIDIA GPU + driver is present on the host.
	GPUAvailable bool `json:"gpu_available"`
}

const (
	warnPct = 85
	critPct = 95
)

// Collect gathers the current host stats (best-effort; never errors).
func Collect() Stats {
	s := Stats{Disk: diskUsage("/")}
	s.MemTotal, s.MemUsed, s.MemPercent = memUsage()
	s.Load1 = loadAvg1()
	s.DiskWarn = s.Disk.Percent >= warnPct
	s.DiskCritical = s.Disk.Percent >= critPct
	s.GPUAvailable = gpuAvailable()
	return s
}

// gpuAvailable reports whether an NVIDIA GPU + driver is present (cheap: just
// checks the device/proc nodes — no exec on the hot polling path).
func gpuAvailable() bool {
	for _, p := range []string{"/dev/nvidia0", "/proc/driver/nvidia/version"} {
		if _, err := os.Stat(p); err == nil {
			return true
		}
	}
	return false
}

func diskUsage(path string) Disk {
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return Disk{Path: path}
	}
	bsize := uint64(st.Bsize)
	total := st.Blocks * bsize
	avail := st.Bavail * bsize
	free := st.Bfree * bsize
	used := total - free
	pct := 0
	if used+avail > 0 {
		// df-style capacity: used / (used + available).
		pct = int((used*100 + (used + avail) - 1) / (used + avail))
	}
	return Disk{Path: path, Total: total, Used: used, Free: avail, Percent: pct}
}

// memUsage reads /proc/meminfo (Linux). Returns zeros where unavailable.
func memUsage() (total, used uint64, percent int) {
	b, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, 0, 0
	}
	var memTotal, memAvail uint64
	for _, line := range strings.Split(string(b), "\n") {
		f := strings.Fields(line)
		if len(f) < 2 {
			continue
		}
		v, _ := strconv.ParseUint(f[1], 10, 64) // kB
		switch f[0] {
		case "MemTotal:":
			memTotal = v * 1024
		case "MemAvailable:":
			memAvail = v * 1024
		}
	}
	if memTotal == 0 {
		return 0, 0, 0
	}
	used = memTotal - memAvail
	return memTotal, used, int(used * 100 / memTotal)
}

// loadAvg1 reads the 1-minute load average from /proc/loadavg (Linux).
func loadAvg1() float64 {
	b, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return 0
	}
	f := strings.Fields(string(b))
	if len(f) == 0 {
		return 0
	}
	v, _ := strconv.ParseFloat(f[0], 64)
	return v
}
