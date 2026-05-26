package system

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gatewarden/agent/pkg/proto"
)

// Gather collects CPU, memory, and disk usage from /proc and df.
func Gather() (*proto.SystemHealth, error) {
	cpu, err := cpuPercent()
	if err != nil {
		cpu = 0
	}

	mem, err := memoryPercent()
	if err != nil {
		mem = 0
	}

	disk, err := diskPercent("/")
	if err != nil {
		disk = 0
	}

	return &proto.SystemHealth{
		CPUPercent:    cpu,
		MemoryPercent: mem,
		DiskPercent:   disk,
	}, nil
}

// --- CPU ---

func cpuPercent() (float64, error) {
	ct, err := readProcStat()
	if err != nil {
		return 0, err
	}
	total := ct.user + ct.nice + ct.system + ct.idle + ct.iowait + ct.irq + ct.softirq + ct.steal
	idle := ct.idle + ct.iowait

	if total == 0 {
		return 0, nil
	}

	return float64(total-idle) / float64(total) * 100.0, nil
}

type cpuTimes struct {
	user, nice, system, idle, iowait, irq, softirq, steal uint64
}

func readProcStat() (*cpuTimes, error) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return nil, err
	}

	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "cpu ") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 9 {
			return nil, fmt.Errorf("unexpected cpu line: %s", line)
		}

		ct := &cpuTimes{}
		vals := make([]uint64, 8)
		for i := 0; i < 8; i++ {
			v, err := strconv.ParseUint(fields[i+1], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("parse cpu field %d: %w", i, err)
			}
			vals[i] = v
		}
		ct.user = vals[0]
		ct.nice = vals[1]
		ct.system = vals[2]
		ct.idle = vals[3]
		ct.iowait = vals[4]
		ct.irq = vals[5]
		ct.softirq = vals[6]
		ct.steal = vals[7]
		return ct, nil
	}

	return nil, fmt.Errorf("cpu line not found in /proc/stat")
}

// --- Memory ---

func memoryPercent() (float64, error) {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, err
	}

	var total, available uint64

	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "MemTotal:") {
			total = parseMemValue(line)
		}
		if strings.HasPrefix(line, "MemAvailable:") {
			available = parseMemValue(line)
		}
	}

	if total == 0 {
		return 0, fmt.Errorf("could not parse MemTotal from /proc/meminfo")
	}

	used := total - available
	return float64(used) / float64(total) * 100.0, nil
}

func parseMemValue(line string) uint64 {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return 0
	}
	v, _ := strconv.ParseUint(fields[1], 10, 64)
	return v
}

// --- Disk ---

func diskPercent(path string) (float64, error) {
	// Use df --sync to get accurate root filesystem usage
	out, err := exec.Command("df", "--sync", "-B1", "--output=size,used", path).Output()
	if err != nil {
		// Fallback without --sync
		out, err = exec.Command("df", "-B1", "--output=size,used", path).Output()
		if err != nil {
			return 0, fmt.Errorf("df failed: %w", err)
		}
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) < 2 {
		return 0, fmt.Errorf("unexpected df output: %q", string(out))
	}

	fields := strings.Fields(lines[1])
	if len(fields) < 2 {
		return 0, fmt.Errorf("unexpected df data line: %s", lines[1])
	}

	total, _ := strconv.ParseUint(fields[0], 10, 64)
	used, _ := strconv.ParseUint(fields[1], 10, 64)

	if total == 0 {
		return 0, nil
	}

	return float64(used) / float64(total) * 100.0, nil
}
