package main

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	weightRegex   = regexp.MustCompile(`(?i)([-+N]?)\s*(\d+(?:[.,]\d+)?)\s*(kg|g|lb|lbs|oz)?\s*([-+]?)`)
	stableRegex   = regexp.MustCompile(`(?i)\bST\b|\bSTABLE\b`)
	unstableRegex = regexp.MustCompile(`(?i)\bUS\b|\bUNSTABLE\b`)
)

type weightCandidate struct {
	weight          float64
	unit            string
	index           int
	hasUnit         bool
	hasExplicitSign bool
	hasNegativeSign bool
	hasPositiveSign bool
}

func parseWeight(raw, defaultUnit string) (float64, string, *bool, bool) {
	normalized := normalizeMinus(raw)
	matches := weightRegex.FindAllStringSubmatchIndex(normalized, -1)
	if len(matches) == 0 {
		return 0, "", nil, false
	}

	candidates := make([]weightCandidate, 0, len(matches))
	for _, idx := range matches {
		if len(idx) < 10 {
			continue
		}

		prefix := submatchByIndex(normalized, idx[2], idx[3])
		numberPart := submatchByIndex(normalized, idx[4], idx[5])
		unitPart := strings.ToLower(strings.TrimSpace(submatchByIndex(normalized, idx[6], idx[7])))
		suffix := submatchByIndex(normalized, idx[8], idx[9])

		sign, hasSign, negative, positive := resolveSign(prefix, suffix)
		numberPart = strings.ReplaceAll(strings.TrimSpace(numberPart), ",", ".")
		if sign != "" {
			numberPart = sign + numberPart
		}

		w, err := strconv.ParseFloat(numberPart, 64)
		if err != nil {
			continue
		}
		if w < -1_000_000 || w > 1_000_000 {
			continue
		}

		unit := unitPart
		if unit == "" {
			unit = strings.ToLower(strings.TrimSpace(defaultUnit))
		}

		candidates = append(candidates, weightCandidate{
			weight:          w,
			unit:            unit,
			index:           idx[0],
			hasUnit:         unitPart != "",
			hasExplicitSign: hasSign,
			hasNegativeSign: negative,
			hasPositiveSign: positive,
		})
	}
	if len(candidates) == 0 {
		return 0, "", nil, false
	}

	best := candidates[0]
	bestScore := scoreCandidate(best)
	for _, c := range candidates[1:] {
		s := scoreCandidate(c)
		if s > bestScore || (s == bestScore && c.index > best.index) {
			best = c
			bestScore = s
		}
	}

	var stable *bool
	if unstableRegex.MatchString(normalized) {
		v := false
		stable = &v
	} else if stableRegex.MatchString(normalized) {
		v := true
		stable = &v
	}

	return best.weight, best.unit, stable, true
}

func scoreCandidate(c weightCandidate) int {
	score := 0
	if c.hasUnit {
		score += 80
	}
	if c.hasExplicitSign {
		score += 40
	}
	if c.hasNegativeSign {
		score += 120
	}
	if c.hasPositiveSign {
		score += 10
	}
	return score
}

func submatchByIndex(raw string, start, end int) string {
	if start < 0 || end < 0 || start > end || end > len(raw) {
		return ""
	}
	return raw[start:end]
}

func resolveSign(prefix, suffix string) (sign string, hasSign bool, negative bool, positive bool) {
	p := strings.ToUpper(strings.TrimSpace(prefix))
	s := strings.TrimSpace(suffix)

	switch {
	case p == "-" || s == "-":
		return "-", true, true, false
	case p == "N":
		return "-", true, true, false
	case p == "+" || s == "+":
		return "+", true, false, true
	default:
		return "", false, false, false
	}
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
