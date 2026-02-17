package main

import "strings"

func safeStr(v, fallback string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return fallback
	}
	return v
}

func inferVerify(line1, line2, expected string) string {
	line1 = strings.TrimSpace(strings.Trim(line1, "\""))
	line2 = strings.TrimSpace(strings.Trim(line2, "\""))
	all := strings.ToUpper(strings.ReplaceAll(line1+line2, " ", ""))
	if line1 == "" && line2 == "" {
		return "UNKNOWN"
	}
	if strings.Contains(strings.ToLower(line1+" "+line2), "no tag") {
		return "NO TAG"
	}
	expected = strings.ToUpper(strings.TrimSpace(expected))
	if expected != "" {
		if strings.Contains(all, expected) {
			return "MATCH"
		}
		return "MISMATCH"
	}
	return "OK"
}
