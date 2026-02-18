package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSupportsColonAndAliasKeys(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, ".env")
	data := "url:https://erp.accord.uz\napi key:abc\napi secret:def\ntelegram bot token:123:XYZ\n"
	if err := os.WriteFile(p, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(p)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if cfg.ERPURL != "https://erp.accord.uz" {
		t.Fatalf("ERPURL mismatch: %q", cfg.ERPURL)
	}
	if cfg.ERPAPIKey != "abc" {
		t.Fatalf("ERPAPIKey mismatch: %q", cfg.ERPAPIKey)
	}
	if cfg.ERPAPISecret != "def" {
		t.Fatalf("ERPAPISecret mismatch: %q", cfg.ERPAPISecret)
	}
	if cfg.TelegramBotToken != "123:XYZ" {
		t.Fatalf("TelegramBotToken mismatch: %q", cfg.TelegramBotToken)
	}
	if cfg.ScaleQtyFile != defaultScaleQtyFile {
		t.Fatalf("ScaleQtyFile mismatch: %q", cfg.ScaleQtyFile)
	}
}
