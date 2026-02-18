package config

import (
	"bufio"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const defaultBridgeStateFile = "/tmp/gscale-zebra/bridge_state.json"
const defaultPrinterDevice = "/dev/usb/lp0"
const defaultLabelWidthDots = 560
const defaultLabelHeightDots = 320

type Config struct {
	TelegramBotToken string
	ERPURL           string
	ERPAPIKey        string
	ERPAPISecret     string
	BridgeStateFile  string
	PrinterDevice    string
	LabelWidthDots   int
	LabelHeightDots  int
}

func Load(envPath string) (Config, error) {
	if strings.TrimSpace(envPath) == "" {
		envPath = ".env"
	}

	fileVals, err := parseEnvFile(envPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, err
	}

	labelWidth, err := parsePositiveInt(
		"LABEL_WIDTH_DOTS",
		firstNonEmpty(
			os.Getenv("LABEL_WIDTH_DOTS"),
			fileVals["LABEL_WIDTH_DOTS"],
			fileVals["PRINTER_LABEL_WIDTH_DOTS"],
		),
		defaultLabelWidthDots,
	)
	if err != nil {
		return Config{}, err
	}

	labelHeight, err := parsePositiveInt(
		"LABEL_HEIGHT_DOTS",
		firstNonEmpty(
			os.Getenv("LABEL_HEIGHT_DOTS"),
			fileVals["LABEL_HEIGHT_DOTS"],
			fileVals["PRINTER_LABEL_HEIGHT_DOTS"],
		),
		defaultLabelHeightDots,
	)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		TelegramBotToken: firstNonEmpty(
			os.Getenv("TELEGRAM_BOT_TOKEN"),
			fileVals["TELEGRAM_BOT_TOKEN"],
			fileVals["BOT_TOKEN"],
			fileVals["TOKEN"],
		),
		ERPURL: firstNonEmpty(
			os.Getenv("ERP_URL"),
			fileVals["ERP_URL"],
			fileVals["URL"],
		),
		ERPAPIKey: firstNonEmpty(
			os.Getenv("ERP_API_KEY"),
			fileVals["ERP_API_KEY"],
			fileVals["API_KEY"],
		),
		ERPAPISecret: firstNonEmpty(
			os.Getenv("ERP_API_SECRET"),
			fileVals["ERP_API_SECRET"],
			fileVals["API_SECRET"],
		),
		BridgeStateFile: firstNonEmpty(
			os.Getenv("BRIDGE_STATE_FILE"),
			fileVals["BRIDGE_STATE_FILE"],
			defaultBridgeStateFile,
		),
		PrinterDevice: firstNonEmpty(
			os.Getenv("PRINTER_DEVICE"),
			fileVals["PRINTER_DEVICE"],
			os.Getenv("ZEBRA_DEVICE"),
			fileVals["ZEBRA_DEVICE"],
			defaultPrinterDevice,
		),
		LabelWidthDots:  labelWidth,
		LabelHeightDots: labelHeight,
	}

	if err := cfg.Validate(); err != nil {
		abs, _ := filepath.Abs(envPath)
		return Config{}, fmt.Errorf("config invalid (%s): %w", abs, err)
	}

	return cfg, nil
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.TelegramBotToken) == "" {
		return errors.New("TELEGRAM_BOT_TOKEN bo'sh")
	}
	if strings.TrimSpace(c.ERPURL) == "" {
		return errors.New("ERP_URL bo'sh")
	}
	if strings.TrimSpace(c.ERPAPIKey) == "" {
		return errors.New("ERP_API_KEY bo'sh")
	}
	if strings.TrimSpace(c.ERPAPISecret) == "" {
		return errors.New("ERP_API_SECRET bo'sh")
	}
	if strings.TrimSpace(c.BridgeStateFile) == "" {
		return errors.New("BRIDGE_STATE_FILE bo'sh")
	}
	if strings.TrimSpace(c.PrinterDevice) == "" {
		return errors.New("PRINTER_DEVICE bo'sh")
	}
	if c.LabelWidthDots <= 0 {
		return errors.New("LABEL_WIDTH_DOTS noto'g'ri")
	}
	if c.LabelHeightDots <= 0 {
		return errors.New("LABEL_HEIGHT_DOTS noto'g'ri")
	}

	u, err := url.Parse(strings.TrimSpace(c.ERPURL))
	if err != nil || u.Scheme == "" || u.Host == "" {
		return errors.New("ERP_URL noto'g'ri (example: https://erp.accord.uz)")
	}
	return nil
}

func parseEnvFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	out := make(map[string]string)
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		k, v, ok := parseLine(line)
		if !ok {
			continue
		}
		out[normalizeKey(k)] = strings.TrimSpace(trimQuotes(v))
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func parseLine(line string) (string, string, bool) {
	if idx := strings.Index(line, "="); idx > 0 {
		return strings.TrimSpace(line[:idx]), strings.TrimSpace(line[idx+1:]), true
	}
	if idx := strings.Index(line, ":"); idx > 0 {
		return strings.TrimSpace(line[:idx]), strings.TrimSpace(line[idx+1:]), true
	}
	return "", "", false
}

func normalizeKey(k string) string {
	k = strings.TrimSpace(strings.ToUpper(k))
	repl := strings.NewReplacer(" ", "_", ".", "_", "-", "_")
	return repl.Replace(k)
}

func trimQuotes(v string) string {
	if len(v) >= 2 {
		if (v[0] == '"' && v[len(v)-1] == '"') || (v[0] == '\'' && v[len(v)-1] == '\'') {
			return v[1 : len(v)-1]
		}
	}
	return v
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func parsePositiveInt(name, raw string, defaultValue int) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return defaultValue, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		return 0, fmt.Errorf("%s noto'g'ri: %q", name, raw)
	}
	return v, nil
}
