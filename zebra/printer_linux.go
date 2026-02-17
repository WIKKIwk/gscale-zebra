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

	fd, err := syscall.Open(device, syscall.O_WRONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		return fmt.Errorf("device ochilmadi: %w", err)
	}
	defer syscall.Close(fd)

	deadline := time.Now().Add(2 * time.Second)
	written := 0
	for written < len(payload) {
		n, werr := syscall.Write(fd, payload[written:])
		if n > 0 {
			written += n
		}
		if werr != nil {
			errNo, ok := werr.(syscall.Errno)
			if ok && (errNo == syscall.EAGAIN || errNo == syscall.EWOULDBLOCK) {
				if time.Now().After(deadline) {
					return fmt.Errorf("yozib bo'lmadi: timeout (%w)", werr)
				}
				time.Sleep(25 * time.Millisecond)
				continue
			}
			return fmt.Errorf("yozib bo'lmadi: %w", werr)
		}
		if n == 0 {
			if time.Now().After(deadline) {
				return fmt.Errorf("yozib bo'lmadi: timeout")
			}
			time.Sleep(25 * time.Millisecond)
		}
	}
	return nil
}

func QueryHostStatus(device string, timeout time.Duration) (string, error) {
	resp, err := transceiveRaw(device, []byte("~HS\n"), timeout)
	if err != nil {
		return "", err
	}
	return normalizeStatusResponse(resp), nil
}

func QuerySGDVar(device, key string, timeout time.Duration) (string, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return "", errors.New("key bo'sh")
	}
	cmd := fmt.Sprintf("! U1 getvar \"%s\"\r\n", key)
	resp, err := transceiveRaw(device, []byte(cmd), timeout)
	if err != nil {
		return "", err
	}
	text := normalizeStatusResponse(resp)
	text = strings.TrimSpace(strings.Trim(text, "\""))
	if text == "" {
		return "", errors.New("bo'sh javob")
	}
	return text, nil
}

func transceiveRaw(device string, payload []byte, timeout time.Duration) ([]byte, error) {
	if timeout <= 0 {
		timeout = 1200 * time.Millisecond
	}

	fd, err := syscall.Open(device, syscall.O_RDWR|syscall.O_NONBLOCK, 0)
	if err != nil {
		if err := SendRaw(device, payload); err != nil {
			return nil, err
		}
		return nil, errors.New("R/W open bo'lmadi; faqat query yuborildi")
	}
	defer syscall.Close(fd)

	if err := writeFDNonBlocking(fd, payload, time.Now().Add(timeout)); err != nil {
		return nil, fmt.Errorf("payload yuborilmadi: %w", err)
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
			return nil, rerr
		}

		if n == 0 {
			time.Sleep(40 * time.Millisecond)
		}
	}

	if len(resp) == 0 {
		return nil, errors.New("javob olinmadi")
	}
	return resp, nil
}

func writeFDNonBlocking(fd int, payload []byte, deadline time.Time) error {
	off := 0
	for off < len(payload) {
		n, err := syscall.Write(fd, payload[off:])
		if n > 0 {
			off += n
		}
		if err != nil {
			errNo, ok := err.(syscall.Errno)
			if ok && (errNo == syscall.EAGAIN || errNo == syscall.EWOULDBLOCK) {
				if time.Now().After(deadline) {
					return fmt.Errorf("timeout: %w", err)
				}
				time.Sleep(20 * time.Millisecond)
				continue
			}
			return err
		}
		if n == 0 {
			if time.Now().After(deadline) {
				return errors.New("timeout")
			}
			time.Sleep(20 * time.Millisecond)
		}
	}
	return nil
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
