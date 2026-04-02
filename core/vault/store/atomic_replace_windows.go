//go:build windows

package store

import "golang.org/x/sys/windows"

func atomicReplaceFile(from, to string) error {
	fromP, err := windows.UTF16PtrFromString(from)
	if err != nil {
		return err
	}
	toP, err := windows.UTF16PtrFromString(to)
	if err != nil {
		return err
	}
	return windows.MoveFileEx(fromP, toP, windows.MOVEFILE_REPLACE_EXISTING|windows.MOVEFILE_WRITE_THROUGH)
}
