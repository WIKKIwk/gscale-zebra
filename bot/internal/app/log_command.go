package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const maxTelegramDocumentBytes = 45 * 1024 * 1024

type workflowLogFile struct {
	AbsPath string
	RelPath string
	Size    int64
}

func (a *App) handleLogCommand(ctx context.Context, chatID int64) error {
	files, err := discoverWorkflowLogFiles()
	if err != nil {
		return a.tg.SendMessage(ctx, chatID, "Log fayllar topilmadi: "+err.Error())
	}

	if err := a.tg.SendMessage(ctx, chatID, fmt.Sprintf("Log yuborish boshlandi (%d ta fayl).", len(files))); err != nil {
		return err
	}

	sent := 0
	skipped := 0
	for _, f := range files {
		if f.Size > maxTelegramDocumentBytes {
			a.logCallback.Printf("log file skipped (too big): %s size=%d", f.AbsPath, f.Size)
			skipped++
			continue
		}

		data, err := os.ReadFile(f.AbsPath)
		if err != nil {
			a.logCallback.Printf("log file read error: %s err=%v", f.AbsPath, err)
			skipped++
			continue
		}

		filename := logDocumentName(f.RelPath)
		if err := a.tg.SendDocument(ctx, chatID, filename, data, f.RelPath); err != nil {
			a.logCallback.Printf("log send error: %s err=%v", f.AbsPath, err)
			skipped++
			continue
		}
		sent++
	}

	return a.tg.SendMessage(ctx, chatID, fmt.Sprintf("Log yuborish tugadi. Yuborildi: %d, o'tkazildi: %d.", sent, skipped))
}

func discoverWorkflowLogFiles() ([]workflowLogFile, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("cwd topilmadi: %w", err)
	}
	return collectWorkflowLogFilesFromStart(wd)
}

func collectWorkflowLogFilesFromStart(start string) ([]workflowLogFile, error) {
	logsRoot, err := findLogsRoot(start)
	if err != nil {
		return nil, err
	}

	dirs := []string{
		filepath.Join(logsRoot, "bot"),
		filepath.Join(logsRoot, "scale"),
	}

	files := make([]workflowLogFile, 0, 16)
	for _, dir := range dirs {
		entries, derr := os.ReadDir(dir)
		if derr != nil {
			if os.IsNotExist(derr) {
				continue
			}
			return nil, fmt.Errorf("log papka o'qilmadi (%s): %w", dir, derr)
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			info, ierr := e.Info()
			if ierr != nil {
				continue
			}
			abs := filepath.Join(dir, e.Name())
			rel, rerr := filepath.Rel(logsRoot, abs)
			if rerr != nil {
				rel = e.Name()
			}
			files = append(files, workflowLogFile{
				AbsPath: abs,
				RelPath: filepath.ToSlash(rel),
				Size:    info.Size(),
			})
		}
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].RelPath < files[j].RelPath
	})

	if len(files) == 0 {
		return nil, fmt.Errorf("logs papkada fayl yo'q")
	}
	return files, nil
}

func findLogsRoot(start string) (string, error) {
	cur, err := filepath.Abs(strings.TrimSpace(start))
	if err != nil {
		return "", fmt.Errorf("path xato: %w", err)
	}

	for {
		candidate := filepath.Join(cur, "logs")
		if st, err := os.Stat(candidate); err == nil && st.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			break
		}
		cur = parent
	}

	return "", fmt.Errorf("logs papkasi topilmadi")
}

func logDocumentName(rel string) string {
	name := strings.TrimSpace(rel)
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.Trim(name, "._- ")
	if name == "" {
		return "workflow.log"
	}
	return name
}
