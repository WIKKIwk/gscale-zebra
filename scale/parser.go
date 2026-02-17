package main

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	weightRegex   = regexp.MustCompile(`([-+]?\d+(?:[.,]\d+)?)\s*(kg|g|lb|lbs|oz)?`)
	stableRegex   = regexp.MustCompile(`(?i)\bST\b|\bSTABLE\b`)
	unstableRegex = regexp.MustCompile(`(?i)\bUS\b|\bUNSTABLE\b`)
)

func parseWeight(raw, defaultUnit string) (float64, string, *bool, bool) {
	matches := weightRegex.FindAllStringSubmatch(raw, -1)
	if len(matches) == 0 {
		return 0, "", nil, false
	}

	m := matches[len(matches)-1]
	v := strings.ReplaceAll(strings.TrimSpace(m[1]), ",", ".")
	weight, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0, "", nil, false
	}
	if weight < -1_000_000 || weight > 1_000_000 {
		return 0, "", nil, false
	}

	unit := strings.ToLower(strings.TrimSpace(m[2]))
	if unit == "" {
		unit = strings.ToLower(strings.TrimSpace(defaultUnit))
	}

	var stable *bool
	if unstableRegex.MatchString(raw) {
		v := false
		stable = &v
	} else if stableRegex.MatchString(raw) {
		v := true
		stable = &v
	}

	return weight, unit, stable, true
}

func appendRaw(existing, chunk string, max int) string {
	combined := existing + chunk
	if len(combined) <= max {
		return combined
	}
	return combined[len(combined)-max:]
}

func stableText(stable *bool) string {
	if stable == nil {
		return "unknown"
	}
	if *stable {
		return "stable"
	}
	return "unstable"
}

func sanitizeInline(raw string, max int) string {
	raw = strings.ReplaceAll(raw, "\r", "\\r")
	raw = strings.ReplaceAll(raw, "\n", "\\n")
	raw = strings.TrimSpace(raw)
	if len(raw) <= max {
		return raw
	}
	return raw[len(raw)-max:]
}
