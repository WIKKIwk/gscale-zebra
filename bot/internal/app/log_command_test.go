package app

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestCollectWorkflowLogFilesFromStart(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	logsDir := filepath.Join(root, "logs")
	botDir := filepath.Join(logsDir, "bot")
	scaleDir := filepath.Join(logsDir, "scale")
	startDir := filepath.Join(root, "bot")

	mustMkdirAll(t, botDir)
	mustMkdirAll(t, scaleDir)
	mustMkdirAll(t, startDir)

	mustWriteFile(t, filepath.Join(botDir, "main.log"), []byte("bot-main"))
	mustWriteFile(t, filepath.Join(botDir, "worker.run.log"), []byte("bot-run"))
	mustWriteFile(t, filepath.Join(scaleDir, "main.log"), []byte("scale-main"))

	files, err := collectWorkflowLogFilesFromStart(startDir)
	if err != nil {
		t.Fatalf("collectWorkflowLogFilesFromStart error: %v", err)
	}

	var got []string
	for _, f := range files {
		got = append(got, f.RelPath)
	}

	want := []string{
		"bot/main.log",
		"bot/worker.run.log",
		"scale/main.log",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("rel paths mismatch\ngot : %#v\nwant: %#v", got, want)
	}
}

func TestLogDocumentName(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"bot/main.log":     "bot_main.log",
		"scale\\main.log":  "scale_main.log",
		"  /a/b/c.log  ":   "a_b_c.log",
		"___":              "workflow.log",
		"":                 "workflow.log",
		"worker.batch.log": "worker.batch.log",
	}

	for in, want := range cases {
		got := logDocumentName(in)
		if got != want {
			t.Fatalf("logDocumentName(%q)=%q want=%q", in, got, want)
		}
	}
}

func mustMkdirAll(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll(%s): %v", dir, err)
	}
}

func mustWriteFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}
}
