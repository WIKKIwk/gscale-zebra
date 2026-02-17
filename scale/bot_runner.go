package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func startBotProcess(ctx context.Context, botDir string) error {
	dir, err := resolveBotDir(botDir)
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, "go", "run", "./cmd/bot")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("bot start xato: %w", err)
	}

	go func() {
		err := cmd.Wait()
		if ctx.Err() == nil && err != nil {
			fmt.Fprintf(os.Stderr, "warning: bot process to'xtadi: %v\n", err)
		}
	}()

	return nil
}

func resolveBotDir(botDir string) (string, error) {
	cands := make([]string, 0, 3)
	if strings.TrimSpace(botDir) != "" {
		cands = append(cands, strings.TrimSpace(botDir))
	}
	cands = append(cands, "../bot", "bot")

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
