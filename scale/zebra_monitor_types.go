package main

import (
	"regexp"
	"sync"
	"time"
)

var zebraHexOnlyRegex = regexp.MustCompile(`^[0-9A-F]+$`)
var zebraIOMutex sync.Mutex

type ZebraStatus struct {
	Connected   bool
	DevicePath  string
	Name        string
	DeviceState string
	MediaState  string
	ReadLine1   string
	ReadLine2   string
	LastEPC     string
	Verify      string
	Action      string
	Error       string
	Attempts    int
	AutoTuned   bool
	Note        string
	UpdatedAt   time.Time
}
