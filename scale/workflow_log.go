package main

import (
	"log"
	"os"

	"core/workflowlog"
)

var scaleWorkflowLogs *workflowlog.Manager

func initWorkflowLogs() error {
	m, err := workflowlog.New("scale")
	if err != nil {
		return err
	}
	scaleWorkflowLogs = m
	return nil
}

func closeWorkflowLogs() {
	if scaleWorkflowLogs != nil {
		scaleWorkflowLogs.Close()
	}
}

func workerLog(name string) *log.Logger {
	if scaleWorkflowLogs != nil {
		return scaleWorkflowLogs.Logger(name)
	}
	return log.New(os.Stdout, "["+name+"] ", log.LstdFlags|log.Lmicroseconds|log.LUTC)
}
