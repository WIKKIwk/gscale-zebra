package labelprint

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const defaultPrinterDevice = "/dev/usb/lp0"
const defaultLabelWidthDots = 560
const defaultLabelHeightDots = 320
const printerGlobalLockPath = "/tmp/gscale-zebra/zebra.lock"

type Service struct {
	devicePath  string
	labelWidth  int
	labelHeight int
}

func New(devicePath string, labelWidth, labelHeight int) *Service {
	devicePath = strings.TrimSpace(devicePath)
	if devicePath == "" {
		devicePath = defaultPrinterDevice
	}
	if labelWidth <= 0 {
		labelWidth = defaultLabelWidthDots
	}
	if labelHeight <= 0 {
		labelHeight = defaultLabelHeightDots
	}

	return &Service{
		devicePath:  devicePath,
		labelWidth:  labelWidth,
		labelHeight: labelHeight,
	}
}

func (s *Service) PrintImageBytes(ctx context.Context, imageBytes []byte) error {
	if len(imageBytes) == 0 {
		return fmt.Errorf("rasm bo'sh")
	}

	img, _, err := decodeImage(imageBytes)
	if err != nil {
		return fmt.Errorf("rasm decode qilinmadi: %w", err)
	}

	zpl, err := BuildImageLabelZPL(img, s.labelWidth, s.labelHeight)
	if err != nil {
		return err
	}

	if err := sendRawWithRetry(ctx, s.devicePath, zpl, 12, 120*time.Millisecond); err != nil {
		return fmt.Errorf("printerga yuborilmadi: %w", err)
	}
	return nil
}

func sendRawWithRetry(ctx context.Context, device string, payload []byte, attempts int, wait time.Duration) error {
	if attempts < 1 {
		attempts = 1
	}

	var lastErr error
	for i := 0; i < attempts; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		err := writeRaw(device, payload)
		if err == nil {
			return nil
		}
		lastErr = err
		if !isRetryablePrinterErr(err) || i == attempts-1 {
			break
		}

		t := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			t.Stop()
			return ctx.Err()
		case <-t.C:
		}
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("noma'lum xato")
	}
	return lastErr
}

func writeRaw(device string, payload []byte) error {
	device = strings.TrimSpace(device)
	if device == "" {
		return fmt.Errorf("printer device bo'sh")
	}
	if len(payload) == 0 {
		return fmt.Errorf("print payload bo'sh")
	}

	return withPrinterGlobalLock(8*time.Second, func() error {
		f, err := os.OpenFile(device, os.O_WRONLY, 0)
		if err != nil {
			return fmt.Errorf("device ochilmadi: %w", err)
		}
		defer f.Close()

		written := 0
		for written < len(payload) {
			n, werr := f.Write(payload[written:])
			if n > 0 {
				written += n
			}
			if werr != nil {
				return fmt.Errorf("yozib bo'lmadi: %w", werr)
			}
			if n == 0 {
				return fmt.Errorf("yozib bo'lmadi: 0 byte")
			}
		}

		return nil
	})
}

func withPrinterGlobalLock(timeout time.Duration, fn func() error) error {
	if timeout <= 0 {
		timeout = 2 * time.Second
	}

	if err := os.MkdirAll(filepath.Dir(printerGlobalLockPath), 0o755); err != nil {
		return fmt.Errorf("lock dir ochilmadi: %w", err)
	}

	f, err := os.OpenFile(printerGlobalLockPath, os.O_CREATE|os.O_RDWR, 0o666)
	if err != nil {
		return fmt.Errorf("lock file ochilmadi: %w", err)
	}
	defer f.Close()

	deadline := time.Now().Add(timeout)
	for {
		err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			break
		}
		if !errors.Is(err, syscall.EWOULDBLOCK) && !errors.Is(err, syscall.EAGAIN) {
			return fmt.Errorf("lock xato: %w", err)
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("lock timeout")
		}
		time.Sleep(25 * time.Millisecond)
	}
	defer func() {
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	}()

	return fn()
}

func isRetryablePrinterErr(err error) bool {
	if err == nil {
		return false
	}

	var pathErr *os.PathError
	if errors.As(err, &pathErr) {
		err = pathErr.Err
	}

	return errors.Is(err, syscall.EBUSY) ||
		errors.Is(err, syscall.EAGAIN) ||
		errors.Is(err, syscall.EWOULDBLOCK)
}
