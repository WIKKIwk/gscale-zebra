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
	if cfg.BridgeStateFile != defaultBridgeStateFile {
		t.Fatalf("BridgeStateFile mismatch: %q", cfg.BridgeStateFile)
	}
	if cfg.PrinterDevice != defaultPrinterDevice {
		t.Fatalf("PrinterDevice mismatch: %q", cfg.PrinterDevice)
	}
	if cfg.LabelWidthDots != defaultLabelWidthDots {
		t.Fatalf("LabelWidthDots mismatch: %d", cfg.LabelWidthDots)
	}
	if cfg.LabelHeightDots != defaultLabelHeightDots {
		t.Fatalf("LabelHeightDots mismatch: %d", cfg.LabelHeightDots)
	}
}

func TestLoadSupportsPrinterAndLabelOverrides(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, ".env")
	data := "\n" +
		"TELEGRAM_BOT_TOKEN=123:XYZ\n" +
		"ERP_URL=https://erp.accord.uz\n" +
		"ERP_API_KEY=abc\n" +
		"ERP_API_SECRET=def\n" +
		"PRINTER_DEVICE=/dev/usb/lp1\n" +
		"LABEL_WIDTH_DOTS=640\n" +
		"LABEL_HEIGHT_DOTS=400\n"
	if err := os.WriteFile(p, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(p)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if cfg.PrinterDevice != "/dev/usb/lp1" {
		t.Fatalf("PrinterDevice mismatch: %q", cfg.PrinterDevice)
	}
	if cfg.LabelWidthDots != 640 {
		t.Fatalf("LabelWidthDots mismatch: %d", cfg.LabelWidthDots)
	}
	if cfg.LabelHeightDots != 400 {
		t.Fatalf("LabelHeightDots mismatch: %d", cfg.LabelHeightDots)
	}
}
