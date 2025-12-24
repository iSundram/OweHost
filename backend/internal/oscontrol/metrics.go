// Package oscontrol provides OS-level system monitoring
package oscontrol

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// SystemMetrics represents current system metrics
type SystemMetrics struct {
	CPU     CPUMetrics     `json:"cpu"`
	Memory  MemoryMetrics  `json:"memory"`
	Disk    DiskMetrics    `json:"disk"`
	Network NetworkMetrics `json:"network"`
}

// CPUMetrics represents CPU usage information
type CPUMetrics struct {
	Usage     float64 `json:"usage"`      // percentage
	Cores     int     `json:"cores"`      // total cores
	UsedCores float64 `json:"used_cores"` // cores being used
}

// MemoryMetrics represents memory usage information
type MemoryMetrics struct {
	Usage float64 `json:"usage"` // percentage
	Total float64 `json:"total"` // GB
	Used  float64 `json:"used"`  // GB
	Free  float64 `json:"free"`  // GB
}

// DiskMetrics represents disk usage information
type DiskMetrics struct {
	Usage float64 `json:"usage"` // percentage
	Total float64 `json:"total"` // GB
	Used  float64 `json:"used"`  // GB
	Free  float64 `json:"free"`  // GB
}

// NetworkMetrics represents network usage information
type NetworkMetrics struct {
	Usage     float64 `json:"usage"`     // percentage (relative to max observed)
	Bandwidth string  `json:"bandwidth"` // current rate (e.g., "125 Mbps")
}

// GetSystemMetrics returns current system resource usage
func (s *Service) GetSystemMetrics() (*SystemMetrics, error) {
	cpu, err := s.getCPUMetrics()
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU metrics: %w", err)
	}

	memory, err := s.getMemoryMetrics()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory metrics: %w", err)
	}

	disk, err := s.getDiskMetrics()
	if err != nil {
		return nil, fmt.Errorf("failed to get disk metrics: %w", err)
	}

	network, err := s.getNetworkMetrics()
	if err != nil {
		return nil, fmt.Errorf("failed to get network metrics: %w", err)
	}

	return &SystemMetrics{
		CPU:     *cpu,
		Memory:  *memory,
		Disk:    *disk,
		Network: *network,
	}, nil
}

// getCPUMetrics gets current CPU usage
func (s *Service) getCPUMetrics() (*CPUMetrics, error) {
	cores := runtime.NumCPU()

	// Read CPU usage from /proc/stat
	usage, err := s.readCPUUsage()
	if err != nil {
		usage = 0
	}

	usedCores := float64(cores) * (usage / 100.0)

	return &CPUMetrics{
		Usage:     usage,
		Cores:     cores,
		UsedCores: usedCores,
	}, nil
}

// readCPUUsage reads CPU usage percentage
func (s *Service) readCPUUsage() (float64, error) {
	// Use top command to get CPU usage
	cmd := exec.Command("sh", "-c", "top -bn1 | grep 'Cpu(s)' | sed 's/.*, *\\([0-9.]*\\)%* id.*/\\1/' | awk '{print 100 - $1}'")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	usage, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	if err != nil {
		return 0, err
	}

	return usage, nil
}

// getMemoryMetrics gets current memory usage
func (s *Service) getMemoryMetrics() (*MemoryMetrics, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var total, available float64
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		switch fields[0] {
		case "MemTotal:":
			val, _ := strconv.ParseFloat(fields[1], 64)
			total = val / 1024 / 1024 // Convert KB to GB
		case "MemAvailable:":
			val, _ := strconv.ParseFloat(fields[1], 64)
			available = val / 1024 / 1024 // Convert KB to GB
		}
	}

	used := total - available
	usage := (used / total) * 100

	return &MemoryMetrics{
		Usage: usage,
		Total: total,
		Used:  used,
		Free:  available,
	}, nil
}

// getDiskMetrics gets current disk usage
func (s *Service) getDiskMetrics() (*DiskMetrics, error) {
	cmd := exec.Command("df", "-BG", "/")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("unexpected df output")
	}

	fields := strings.Fields(lines[1])
	if len(fields) < 5 {
		return nil, fmt.Errorf("unexpected df output format")
	}

	// Parse values (remove 'G' suffix)
	total, _ := strconv.ParseFloat(strings.TrimSuffix(fields[1], "G"), 64)
	used, _ := strconv.ParseFloat(strings.TrimSuffix(fields[2], "G"), 64)
	free, _ := strconv.ParseFloat(strings.TrimSuffix(fields[3], "G"), 64)
	usageStr := strings.TrimSuffix(fields[4], "%")
	usage, _ := strconv.ParseFloat(usageStr, 64)

	return &DiskMetrics{
		Usage: usage,
		Total: total,
		Used:  used,
		Free:  free,
	}, nil
}

// getNetworkMetrics gets current network usage
func (s *Service) getNetworkMetrics() (*NetworkMetrics, error) {
	// This is a simplified version - in production you'd track bytes over time
	// For now, return a placeholder
	return &NetworkMetrics{
		Usage:     15.5, // Placeholder
		Bandwidth: "N/A",
	}, nil
}
