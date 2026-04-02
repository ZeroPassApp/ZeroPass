//go:build !windows

package store

import "os"

func atomicReplaceFile(from, to string) error {
	return os.Rename(from, to)
}
