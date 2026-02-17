package main

import "time"

type Reading struct {
	Source    string
	Port      string
	Baud      int
	Weight    *float64
	Unit      string
	Stable    *bool
	Raw       string
	Error     string
	UpdatedAt time.Time
}

type scaleAPIResponse struct {
	OK     bool     `json:"ok"`
	Weight *float64 `json:"weight"`
	Unit   string   `json:"unit"`
	Stable *bool    `json:"stable"`
	Port   string   `json:"port"`
	Raw    string   `json:"raw"`
	Error  string   `json:"error"`
}
