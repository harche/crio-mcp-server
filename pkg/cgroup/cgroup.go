package cgroup

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Stats holds basic cgroup v2 resource statistics.
type Stats struct {
	CPUUsageUSec     uint64
	MemoryUsageBytes uint64
}

// FindCgroup2Mountpoint returns the mount point of the cgroup v2 filesystem.
func FindCgroup2Mountpoint() (string, error) {
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, " ")
		if len(fields) < 10 {
			continue
		}
		sep := -1
		for i, v := range fields {
			if v == "-" {
				sep = i
				break
			}
		}
		if sep == -1 || sep+3 >= len(fields) {
			continue
		}
		fstype := fields[sep+1]
		if fstype == "cgroup2" {
			return fields[4], nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", errors.New("cgroup2 mountpoint not found")
}

// PIDFromInspect extracts the container PID from a crictl inspect JSON output.
func PIDFromInspect(data string) (int, error) {
	var obj interface{}
	if err := json.Unmarshal([]byte(data), &obj); err != nil {
		return 0, err
	}
	if pid, ok := findIntKey(obj, "pid"); ok {
		return pid, nil
	}
	return 0, errors.New("pid not found in inspect data")
}

func findIntKey(obj interface{}, key string) (int, bool) {
	switch v := obj.(type) {
	case map[string]interface{}:
		for k, val := range v {
			if k == key {
				switch t := val.(type) {
				case float64:
					return int(t), true
				case int:
					return t, true
				}
			}
			if pid, ok := findIntKey(val, key); ok {
				return pid, true
			}
		}
	case []interface{}:
		for _, val := range v {
			if pid, ok := findIntKey(val, key); ok {
				return pid, true
			}
		}
	}
	return 0, false
}

// CgroupPathForPID returns the cgroup v2 path for a given pid.
func CgroupPathForPID(pid int) (string, error) {
	f, err := os.Open(fmt.Sprintf("/proc/%d/cgroup", pid))
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 3)
		if len(parts) == 3 && parts[1] == "" {
			return parts[2], nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", errors.New("cgroup path not found")
}

// StatsFromPath reads cpu and memory statistics from a cgroup v2 path.
func StatsFromPath(root, cgPath string) (Stats, error) {
	cpuFile := filepath.Join(root, cgPath, "cpu.stat")
	memFile := filepath.Join(root, cgPath, "memory.current")

	var stats Stats

	if f, err := os.Open(cpuFile); err == nil {
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			fields := strings.Fields(scanner.Text())
			if len(fields) == 2 && fields[0] == "usage_usec" {
				fmt.Sscanf(fields[1], "%d", &stats.CPUUsageUSec)
				break
			}
		}
		f.Close()
	}

	if data, err := os.ReadFile(memFile); err == nil {
		fmt.Sscanf(strings.TrimSpace(string(data)), "%d", &stats.MemoryUsageBytes)
	}

	return stats, nil
}

// StatsFromInspect returns cgroup stats using a crictl inspect JSON payload.
func StatsFromInspect(data string) (Stats, error) {
	pid, err := PIDFromInspect(data)
	if err != nil {
		return Stats{}, err
	}
	cgPath, err := CgroupPathForPID(pid)
	if err != nil {
		return Stats{}, err
	}
	mount, err := FindCgroup2Mountpoint()
	if err != nil {
		return Stats{}, err
	}
	return StatsFromPath(mount, cgPath)
}
