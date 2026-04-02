//go:build cgo

package main

/*
#include "zp_bridge.h"
*/
import "C"

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/zeropass/zeropass/core/crypto"
	"github.com/zeropass/zeropass/core/vault/importexport"
	"github.com/zeropass/zeropass/core/vault/types"
)

func ensureParentDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0700)
}

func importFromFile(s *vaultSession, path string, fn func(io.Reader) ([]types.Item, error)) (int, error) {
	if err := ensureManagersUnlocked(s); err != nil {
		return 0, err
	}
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	items, err := fn(f)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, it := range items {
		itm := it
		itm.Fields = copyMap(itm.Fields)
		itm.CustomFields = copyMap(itm.CustomFields)
		if err := s.mgr.AddItem(&itm); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func exportToFile(s *vaultSession, path string, fn func([]types.Item, io.Writer) error) error {
	if err := ensureManagersUnlocked(s); err != nil {
		return err
	}
	itemsPtr, err := s.mgr.AllItems()
	if err != nil {
		return err
	}
	items := make([]types.Item, 0, len(itemsPtr))
	for _, p := range itemsPtr {
		if p == nil {
			continue
		}
		items = append(items, *p)
	}

	if err := ensureParentDir(path); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return fn(items, f)
}

func copyMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

//export ZPImportCSV
func ZPImportCSV(handle C.long, path *C.char) (res C.ZPResult) {
	defer recoverToResult(&res)

	h := int64(handle)
	fp := goString(path)
	if fp == "" {
		return errorResult(fmt.Errorf("path must not be empty"))
	}
	
	s, err := getSession(h)
	if err != nil {
		return errorResult(err)
	}
	
	s.mu.Lock()
	defer s.mu.Unlock()
	
	mapping := importexport.CSVMapping{ // compatible with ExportCSV
		Name:     0,
		Type:     1,
		URL:      2,
		Username: 3,
		Password: 4,
		Notes:    5,
		Skip:     1,
	}
	count, err := importFromFile(s, fp, func(r io.Reader) ([]types.Item, error) {
		return importexport.ImportCSV(r, mapping)
	})
	if err != nil {
		return errorResult(err)
	}
	return okJSON(map[string]any{"imported": count})
}

//export ZPImportChrome
func ZPImportChrome(handle C.long, path *C.char) (res C.ZPResult) {
	defer recoverToResult(&res)
	return importSimple(handle, path, importexport.ImportChrome)
}

//export ZPImportFirefox
func ZPImportFirefox(handle C.long, path *C.char) (res C.ZPResult) {
	defer recoverToResult(&res)
	return importSimple(handle, path, importexport.ImportFirefox)
}

//export ZPImport1Password
func ZPImport1Password(handle C.long, path *C.char) (res C.ZPResult) {
	defer recoverToResult(&res)
	return importSimple(handle, path, importexport.Import1Password)
}

//export ZPImportBitwarden
func ZPImportBitwarden(handle C.long, path *C.char) (res C.ZPResult) {
	defer recoverToResult(&res)
	return importSimple(handle, path, importexport.ImportBitwarden)
}

//export ZPImportSafari
func ZPImportSafari(handle C.long, path *C.char) (res C.ZPResult) {
	defer recoverToResult(&res)
	return importSimple(handle, path, importexport.ImportSafari)
}

//export ZPImportLastPass
func ZPImportLastPass(handle C.long, path *C.char) (res C.ZPResult) {
	defer recoverToResult(&res)
	return importSimple(handle, path, importexport.ImportLastPass)
}

//export ZPImportKeePass
func ZPImportKeePass(handle C.long, path *C.char) (res C.ZPResult) {
	defer recoverToResult(&res)
	return importSimple(handle, path, importexport.ImportKeePass)
}

//export ZPImport1PUX
func ZPImport1PUX(handle C.long, path *C.char) (res C.ZPResult) {
	defer recoverToResult(&res)
	return importSimple(handle, path, importexport.Import1PUX)
}

func importSimple(handle C.long, path *C.char, fn func(io.Reader) ([]types.Item, error)) C.ZPResult {
	h := int64(handle)
	fp := goString(path)
	if fp == "" {
		return errorResult(fmt.Errorf("path must not be empty"))
	}

	s, err := getSession(h)
	if err != nil {
		return errorResult(err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	count, err := importFromFile(s, fp, fn)
	if err != nil {
		return errorResult(err)
	}
	return okJSON(map[string]any{"imported": count})
}

//export ZPExportJSON
func ZPExportJSON(handle C.long, path *C.char) (res C.ZPResult) {
	defer recoverToResult(&res)
	return exportSimple(handle, path, importexport.ExportJSON)
}

//export ZPExportCSV
func ZPExportCSV(handle C.long, path *C.char) (res C.ZPResult) {
	defer recoverToResult(&res)
	return exportSimple(handle, path, importexport.ExportCSV)
}

//export ZPExportEncrypted
func ZPExportEncrypted(handle C.long, path *C.char) (res C.ZPResult) {
	defer recoverToResult(&res)

	h := int64(handle)
	fp := goString(path)
	if fp == "" {
		return errorResult(fmt.Errorf("path must not be empty"))
	}

	s, err := getSession(h)
	if err != nil {
		return errorResult(err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := ensureManagersUnlocked(s); err != nil {
		return errorResult(err)
	}

	vk, err := s.vault.VaultKey()
	if err != nil {
		return errorResult(err)
	}
	keyCopy := make([]byte, len(vk))
	copy(keyCopy, vk)
	defer crypto.ZeroBytes(keyCopy)

	err = exportEncryptedToFile(s, fp, keyCopy)
	if err != nil {
		return errorResult(err)
	}
	return okNoData()
}

func exportSimple(handle C.long, path *C.char, fn func([]types.Item, io.Writer) error) C.ZPResult {
	h := int64(handle)
	fp := goString(path)
	if fp == "" {
		return errorResult(fmt.Errorf("path must not be empty"))
	}
	
	s, err := getSession(h)
	if err != nil {
		return errorResult(err)
	}
	
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if err := exportToFile(s, fp, fn); err != nil {
		return errorResult(err)
	}
	return okNoData()
}

func exportEncryptedToFile(s *vaultSession, path string, key []byte) error {
	itemsPtr, err := s.mgr.AllItems()
	if err != nil {
		return err
	}
	items := make([]types.Item, 0, len(itemsPtr))
	for _, p := range itemsPtr {
		if p != nil {
			items = append(items, *p)
		}
	}
	if err := ensureParentDir(path); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return importexport.ExportEncrypted(items, key, f)
}
