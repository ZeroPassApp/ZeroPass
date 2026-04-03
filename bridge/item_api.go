//go:build cgo

package main

/*
#include "zp_bridge.h"
*/
import "C"

import (
	"encoding/json"
	"fmt"

	"github.com/zeropass/zeropass/core/vault/types"
)

//export ZPListItems
func ZPListItems(handle C.long, filterJSON *C.char) (res C.ZPResult) {
	defer recoverToResult(&res)

	h := int64(handle)
	s, err := getSession(h)
	if err != nil {
		return errorResult(err)
	}
	filterStr := goString(filterJSON)

	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ensureManagersUnlocked(s); err != nil {
		return errorResult(err)
	}

	filter := types.ItemFilter{}
	if filterStr != "" {
		if err := json.Unmarshal([]byte(filterStr), &filter); err != nil {
			return errorResult(fmt.Errorf("parse filter JSON: %w", err))
		}
	}

	items, err := s.mgr.ListItems(filter)
	if err != nil {
		return errorResult(err)
	}
	return okJSON(items)
}

//export ZPGetItem
func ZPGetItem(handle C.long, itemID *C.char) (res C.ZPResult) {
	defer recoverToResult(&res)

	h := int64(handle)
	id := goString(itemID)
	if id == "" {
		return errorResult(fmt.Errorf("item ID must not be empty"))
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

	itm, err := s.mgr.GetItem(id)
	if err != nil {
		return errorResult(err)
	}
	return okJSON(itm)
}

//export ZPCreateItem
func ZPCreateItem(handle C.long, itemJSON *C.char) (res C.ZPResult) {
	defer recoverToResult(&res)

	h := int64(handle)
	payload := goString(itemJSON)
	if payload == "" {
		return errorResult(fmt.Errorf("item JSON must not be empty"))
	}

	s, err := getSession(h)
	if err != nil {
		return errorResult(err)
	}

	var itm types.Item
	if err := json.Unmarshal([]byte(payload), &itm); err != nil {
		return errorResult(fmt.Errorf("parse item JSON: %w", err))
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ensureManagersUnlocked(s); err != nil {
		return errorResult(err)
	}

	if err := s.mgr.AddItem(&itm); err != nil {
		return errorResult(err)
	}
	return okJSON(&itm)
}

//export ZPUpdateItem
func ZPUpdateItem(handle C.long, itemID *C.char, itemJSON *C.char) (res C.ZPResult) {
	defer recoverToResult(&res)

	h := int64(handle)
	id := goString(itemID)
	payload := goString(itemJSON)
	if id == "" {
		return errorResult(fmt.Errorf("item ID must not be empty"))
	}
	if payload == "" {
		return errorResult(fmt.Errorf("item JSON must not be empty"))
	}

	s, err := getSession(h)
	if err != nil {
		return errorResult(err)
	}

	var updated types.Item
	if err := json.Unmarshal([]byte(payload), &updated); err != nil {
		return errorResult(fmt.Errorf("parse item JSON: %w", err))
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ensureManagersUnlocked(s); err != nil {
		return errorResult(err)
	}

	cur, err := s.mgr.GetItem(id)
	if err != nil {
		return errorResult(err)
	}
	if s.verMgr != nil {
		_ = s.verMgr.SaveVersion(cur)
	}

	updated.ID = id
	updated.CreatedAt = cur.CreatedAt
	updated.Version = cur.Version

	if err := s.mgr.UpdateItem(id, &updated); err != nil {
		return errorResult(err)
	}
	return okJSON(&updated)
}

//export ZPDeleteItem
func ZPDeleteItem(handle C.long, itemID *C.char) (res C.ZPResult) {
	defer recoverToResult(&res)

	h := int64(handle)
	id := goString(itemID)
	if id == "" {
		return errorResult(fmt.Errorf("item ID must not be empty"))
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

	if err := s.mgr.DeleteItem(id); err != nil {
		return errorResult(err)
	}
	return okNoData()
}
