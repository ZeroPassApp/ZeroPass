//go:build cgo

package main

import (
	"io"
	"os"
	"path/filepath"
)

func writeFileAtomic(path string, perm os.FileMode, fn func(w io.Writer) error) error {
	if err := ensureParentDir(path); err != nil {
		return err
	}

	dir := filepath.Dir(path)
	base := filepath.Base(path)

	f, err := os.CreateTemp(dir, base+".*.tmp")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer os.Remove(tmp)

	if err := f.Chmod(perm); err != nil {
		_ = f.Close()
		return err
	}
	if err := fn(f); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	if err := os.Rename(tmp, path); err != nil {
		return err
	}

	if d, err := os.Open(dir); err == nil {
		_ = d.Sync()
		_ = d.Close()
	}

	return nil
}

func writeFileAtomicBytes(path string, perm os.FileMode, b []byte) error {
	return writeFileAtomic(path, perm, func(w io.Writer) error {
		_, err := w.Write(b)
		return err
	})
}
