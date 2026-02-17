package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"
)

type USBLPPrinter struct {
	DevicePath   string
	VendorID     string
	ProductID    string
	Manufacturer string
	Product      string
	Serial       string
	BusNum       string
	DevNum       string
}

func (p USBLPPrinter) IsZebra() bool {
	if strings.EqualFold(strings.TrimSpace(p.VendorID), "0a5f") {
		return true
	}
	text := strings.ToLower(strings.TrimSpace(p.Manufacturer + " " + p.Product))
	return strings.Contains(text, "zebra") || strings.Contains(text, "ztc")
}

func (p USBLPPrinter) DisplayName() string {
	name := strings.TrimSpace(p.Manufacturer + " " + p.Product)
	if name == "" {
		return "unknown"
	}
	return name
}

func FindUSBLPPrinters() ([]USBLPPrinter, error) {
	devices, err := filepath.Glob("/dev/usb/lp*")
	if err != nil {
		return nil, err
	}

	printers := make([]USBLPPrinter, 0, len(devices))
	for _, dev := range devices {
		p := USBLPPrinter{DevicePath: dev}
		fillPrinterSysfs(&p)
		printers = append(printers, p)
	}

	sort.Slice(printers, func(i, j int) bool {
		return printers[i].DevicePath < printers[j].DevicePath
	})
	return printers, nil
}

func SelectPrinter(preferred string) (USBLPPrinter, error) {
	printers, err := FindUSBLPPrinters()
	if err != nil {
		return USBLPPrinter{}, err
	}
	if len(printers) == 0 {
		return USBLPPrinter{}, errors.New("USB printer topilmadi")
	}

	if strings.TrimSpace(preferred) != "" {
		want := strings.TrimSpace(preferred)
		for _, p := range printers {
			if p.DevicePath == want {
				return p, nil
			}
		}
		return USBLPPrinter{}, fmt.Errorf("ko'rsatilgan device topilmadi: %s", want)
	}

	for _, p := range printers {
		if p.IsZebra() {
			return p, nil
		}
	}
	return printers[0], nil
}

func fillPrinterSysfs(p *USBLPPrinter) {
	base := filepath.Base(p.DevicePath)
	classPath := filepath.Join("/sys/class/usbmisc", base)
	ifacePath, err := filepath.EvalSymlinks(filepath.Join(classPath, "device"))
	if err != nil {
		return
	}

	parent := filepath.Dir(ifacePath)
	p.VendorID = readTrim(filepath.Join(parent, "idVendor"))
	p.ProductID = readTrim(filepath.Join(parent, "idProduct"))
	p.Manufacturer = readTrim(filepath.Join(parent, "manufacturer"))
	p.Product = readTrim(filepath.Join(parent, "product"))
	p.Serial = readTrim(filepath.Join(parent, "serial"))
	p.BusNum = readTrim(filepath.Join(parent, "busnum"))
	p.DevNum = readTrim(filepath.Join(parent, "devnum"))
}

func readTrim(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func SendRaw(device string, payload []byte) error {
	if strings.TrimSpace(device) == "" {
		return errors.New("device bo'sh")
	}
	if len(payload) == 0 {
		return errors.New("payload bo'sh")
	}

	f, err := os.OpenFile(device, os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("device ochilmadi: %w", err)
	}
	defer f.Close()

	n, err := f.Write(payload)
	if err != nil {
		return fmt.Errorf("yozib bo'lmadi: %w", err)
	}
	if n != len(payload) {
		return fmt.Errorf("to'liq yozilmadi: %d/%d", n, len(payload))
	}
	return nil
}

func QueryHostStatus(device string, timeout time.Duration) (string, error) {
	if timeout <= 0 {
		timeout = 1200 * time.Millisecond
	}

	fd, err := syscall.Open(device, syscall.O_RDWR|syscall.O_NONBLOCK, 0)
	if err != nil {
		if err := SendRaw(device, []byte("~HS\n")); err != nil {
			return "", err
		}
		return "", errors.New("R/W open bo'lmadi; faqat query yuborildi")
	}
	defer syscall.Close(fd)

	if _, err := syscall.Write(fd, []byte("~HS\n")); err != nil {
		return "", fmt.Errorf("~HS yuborilmadi: %w", err)
	}

	deadline := time.Now().Add(timeout)
	buf := make([]byte, 4096)
	resp := make([]byte, 0, 4096)

	for time.Now().Before(deadline) {
		n, rerr := syscall.Read(fd, buf)
		if n > 0 {
			resp = append(resp, buf[:n]...)
			if n < len(buf) {
				break
			}
		}

		if rerr != nil {
			errNo, ok := rerr.(syscall.Errno)
			if ok && (errNo == syscall.EAGAIN || errNo == syscall.EWOULDBLOCK) {
				time.Sleep(40 * time.Millisecond)
				continue
			}
			if len(resp) > 0 {
				break
			}
			return "", rerr
		}

		if n == 0 {
			time.Sleep(40 * time.Millisecond)
		}
	}

	if len(resp) == 0 {
		return "", errors.New("status javobi olinmadi")
	}

	return normalizeStatusResponse(resp), nil
}

func normalizeStatusResponse(raw []byte) string {
	text := string(raw)
	text = strings.ReplaceAll(text, "\x00", "")
	text = strings.ReplaceAll(text, "\r", "\n")
	rows := strings.Split(text, "\n")
	clean := make([]string, 0, len(rows))
	for _, row := range rows {
		row = strings.TrimSpace(row)
		if row == "" {
			continue
		}
		clean = append(clean, row)
	}
	return strings.Join(clean, "\n")
}
