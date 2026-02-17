package main

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	weightRegex   = regexp.MustCompile(`(?i)([-+]?)\s*(\d+(?:[.,]\d+)?)\s*(kg|g|lb|lbs|oz)?\s*([-+]?)`)
	stableRegex   = regexp.MustCompile(`(?i)\bST\b|\bSTABLE\b`)
	unstableRegex = regexp.MustCompile(`(?i)\bUS\b|\bUNSTABLE\b`)
)

func parseWeight(raw, defaultUnit string) (float64, string, *bool, bool) {
	normalized := normalizeMinus(raw)
	matches := weightRegex.FindAllStringSubmatch(normalized, -1)
	if len(matches) == 0 {
		return 0, "", nil, false
	}

	// 1-pass: unit bor matchlar, 2-pass: unit yo'q matchlar.
	for pass := 0; pass < 2; pass++ {
		requireUnit := pass == 0
		for i := len(matches) - 1; i >= 0; i-- {
			weight, unit, ok := parseWeightMatch(matches[i], defaultUnit, requireUnit)
			if !ok {
				continue
			}

			var stable *bool
			if unstableRegex.MatchString(normalized) {
				v := false
				stable = &v
			} else if stableRegex.MatchString(normalized) {
				v := true
				stable = &v
			}

			return weight, unit, stable, true
		}
	}

	return 0, "", nil, false
}

func parseWeightMatch(m []string, defaultUnit string, requireUnit bool) (float64, string, bool) {
	if len(m) < 5 {
		return 0, "", false
	}

	prefixSign := strings.TrimSpace(m[1])
	numberPart := strings.TrimSpace(m[2])
	unitPart := strings.ToLower(strings.TrimSpace(m[3]))
	suffixSign := strings.TrimSpace(m[4])

	if requireUnit && unitPart == "" {
		return 0, "", false
	}

	sign := prefixSign
	if sign == "" && (suffixSign == "-" || suffixSign == "+") {
		sign = suffixSign
	}

	numberPart = strings.ReplaceAll(numberPart, ",", ".")
	if sign == "-" || sign == "+" {
		numberPart = sign + numberPart
	}

	weight, err := strconv.ParseFloat(numberPart, 64)
	if err != nil {
		return 0, "", false
	}
	if weight < -1_000_000 || weight > 1_000_000 {
		return 0, "", false
	}

	if unitPart == "" {
		unitPart = strings.ToLower(strings.TrimSpace(defaultUnit))
	}

	return weight, unitPart, true
}

func normalizeMinus(raw string) string {
	replacer := strings.NewReplacer(
		"−", "-",
		"–", "-",
		"—", "-",
	)
	return replacer.Replace(raw)
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
