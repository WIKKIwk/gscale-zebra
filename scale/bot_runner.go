package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

type BotProcess struct {
	cmd  *exec.Cmd
	done chan error
}

func startBotProcess(botDir string) (*BotProcess, error) {
	dir, err := resolveBotDir(botDir)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("go", "run", "./cmd/bot")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("bot start xato: %w", err)
	}

	bp := &BotProcess{cmd: cmd, done: make(chan error, 1)}
	go func() {
		bp.done <- cmd.Wait()
	}()
	return bp, nil
}

func (bp *BotProcess) Stop(timeout time.Duration) error {
	if bp == nil || bp.cmd == nil || bp.cmd.Process == nil {
		return nil
	}
	if timeout <= 0 {
		timeout = 3 * time.Second
	}

	pid := bp.cmd.Process.Pid
	if pgid, err := syscall.Getpgid(pid); err == nil {
		_ = syscall.Kill(-pgid, syscall.SIGTERM)
	} else {
		_ = bp.cmd.Process.Signal(syscall.SIGTERM)
	}

	select {
	case <-time.After(timeout):
		if pgid, err := syscall.Getpgid(pid); err == nil {
			_ = syscall.Kill(-pgid, syscall.SIGKILL)
		} else {
			_ = bp.cmd.Process.Kill()
		}
		select {
		case <-bp.done:
		default:
		}
		return fmt.Errorf("bot force-killed (pid=%d)", pid)
	case err := <-bp.done:
		if err == nil {
			return nil
		}
		return nil
	}
}

func resolveBotDir(botDir string) (string, error) {
	cands := make([]string, 0, 6)
	if strings.TrimSpace(botDir) != "" {
		cands = append(cands, strings.TrimSpace(botDir))
	}
	cands = append(cands, "../bot", "bot")

	if wd, err := os.Getwd(); err == nil {
		cands = append(cands,
			filepath.Join(wd, "../bot"),
			filepath.Join(wd, "bot"),
		)
	}

	for _, c := range cands {
		abs, err := filepath.Abs(c)
		if err != nil {
			continue
		}
		st, err := os.Stat(abs)
		if err != nil || !st.IsDir() {
			continue
		}
		if _, err := os.Stat(filepath.Join(abs, "cmd", "bot", "main.go")); err == nil {
			return abs, nil
		}
	}

	return "", errors.New("bot papkasi topilmadi (cmd/bot/main.go yo'q)")
}
