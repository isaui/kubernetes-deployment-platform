package utils

import (
	"fmt"
)

// FormatCPUCores formats CPU cores value to a consistent string format
// Input is in milliCores, output is in cores with 2 decimal precision
func FormatCPUCores(milliCores int64) string {
	cores := float64(milliCores) / 1000.0
	return fmt.Sprintf("%.2f", cores)
}

// FormatBytesToHumanReadable formats bytes to human-readable format (Ki, Mi, Gi)
func FormatBytesToHumanReadable(bytes int64) string {
	const (
		KiB int64 = 1024
		MiB = KiB * 1024
		GiB = MiB * 1024
	)

	switch {
	case bytes >= GiB:
		return fmt.Sprintf("%.2fGi", float64(bytes)/float64(GiB))
	case bytes >= MiB:
		return fmt.Sprintf("%.2fMi", float64(bytes)/float64(MiB))
	case bytes >= KiB:
		return fmt.Sprintf("%.2fKi", float64(bytes)/float64(KiB))
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}

// CalculatePercentage calculates usage percentage
// Returns 0 if denominator is 0 or negative
func CalculatePercentage(used, total int64) float64 {
	if total <= 0 {
		return 0
	}
	return (float64(used) / float64(total)) * 100
}
