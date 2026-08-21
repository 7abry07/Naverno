package utils

import (
	"fmt"
	"strings"
)

func Clamp(text string, limit int) string {
	if len(text) == limit {
		return text
	} else if len(text) > limit-3 {
		text = text[:limit-3] + "..."
	} else {
		bd := &strings.Builder{}
		bd.WriteString(text)
		for range limit - len(text) {
			bd.WriteString(" ")
		}
		text = bd.String()
	}
	return text
}

func FormatRate(rate uint64) string {
	if rate > 1000000000000 {
		return fmt.Sprintf("%.2f Tbps", float64(rate)/1000000000000.0)
	}
	if rate > 1000000000 {
		return fmt.Sprintf("%.2f Gbps", float64(rate)/1000000000.0)
	}
	if rate > 1000000 {
		return fmt.Sprintf("%.2f Mbps", float64(rate)/1000000.0)
	}
	if rate > 1000 {
		return fmt.Sprintf("%.2f Kbps", float64(rate)/1000.0)
	}
	return fmt.Sprintf("%v bps", rate)
}

func FormatLength(length uint64) string {
	if length > 1000000000000 {
		return fmt.Sprintf("%.2f TiB", float64(length)/1000000000000.0)
	}
	if length > 1000000000 {
		return fmt.Sprintf("%.2f GiB", float64(length)/1000000000.0)
	}
	if length > 1000000 {
		return fmt.Sprintf("%.2f MiB", float64(length)/1000000.0)
	}
	if length > 1000 {
		return fmt.Sprintf("%.2f KiB", float64(length)/1000.0)
	}
	return fmt.Sprintf("%v B", length)
}
