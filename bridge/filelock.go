//go:build cgo

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

const vaultLockFileName = "vault.lock"

func acquireVaultLock(vaultPath string, createDir bool) (*os.File, error) {
	if vaultPath == "" {
		return nil, errors.New("vault path must not be empty")
	}
	if createDir {
		if err := os.MkdirAll(vaultPath, 0700); err != nil {
			return nil, fmt.Errorf("create vault dir: %w", err)
		}
	} else {
		st, err := os.Stat(vaultPath)
		if err != nil {
			return nil, err
		}
		if !st.IsDir() {
			return nil, fmt.Errorf("vault path is not a directory")
		}
	}

	fp := filepath.Join(vaultPath, vaultLockFileName)
	f, err := os.OpenFile(fp, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, fmt.Errorf("open lock file: %w", err)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = f.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN) {
			return nil, fmt.Errorf("%w", errVaultBusy)
		}
		return nil, fmt.Errorf("acquire vault lock: %w", err)
	}
	return f, nil
}

func releaseVaultLock(f *os.File) {
	if f == nil {
		return
	}
	_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	_ = f.Close()
}
