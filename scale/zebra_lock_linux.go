package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

const zebraGlobalLockPath = "/tmp/gscale-zebra/zebra.lock"

func withZebraGlobalLock(timeout time.Duration, fn func() error) error {
	if timeout <= 0 {
		timeout = 2 * time.Second
	}

	if err := os.MkdirAll(filepath.Dir(zebraGlobalLockPath), 0o755); err != nil {
		return fmt.Errorf("zebra: lock dir ochilmadi: %w", err)
	}

	f, err := os.OpenFile(zebraGlobalLockPath, os.O_CREATE|os.O_RDWR, 0o666)
	if err != nil {
		return fmt.Errorf("zebra: lock file ochilmadi: %w", err)
	}
	defer f.Close()

	deadline := time.Now().Add(timeout)
	for {
		err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			break
		}
		if !errors.Is(err, syscall.EWOULDBLOCK) && !errors.Is(err, syscall.EAGAIN) {
			return fmt.Errorf("zebra: lock xato: %w", err)
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("zebra: lock timeout")
		}
		time.Sleep(25 * time.Millisecond)
	}

	defer func() {
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	}()

	return fn()
}
