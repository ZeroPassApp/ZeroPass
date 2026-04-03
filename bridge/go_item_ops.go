//go:build cgo

package main

import (
	"encoding/json"
	"fmt"

	"github.com/zeropass/zeropass/core/vault/types"
)

func zpListItems(handle int64, filterJSON string) goResult {
	s, err := getSession(handle)
	if err != nil {
		return errorResultGo(err)
	}

	filter := types.ItemFilter{}
	if filterJSON != "" {
		if err := json.Unmarshal([]byte(filterJSON), &filter); err != nil {
			return errorResultGo(fmt.Errorf("parse filter JSON: %w", err))
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ensureManagersUnlocked(s); err != nil {
		return errorResultGo(err)
	}

	items, err := s.mgr.ListItems(filter)
	if err != nil {
		return errorResultGo(err)
	}
	return okJSONGo(items)
}

func zpGetItem(handle int64, id string) goResult {
	if id == "" {
		return errorResultGo(fmt.Errorf("item ID must not be empty"))
	}
	s, err := getSession(handle)
	if err != nil {
		return errorResultGo(err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ensureManagersUnlocked(s); err != nil {
		return errorResultGo(err)
	}

	itm, err := s.mgr.GetItem(id)
	if err != nil {
		return errorResultGo(err)
	}
	return okJSONGo(itm)
}

func zpCreateItem(handle int64, itemJSON string) goResult {
	if itemJSON == "" {
		return errorResultGo(fmt.Errorf("item JSON must not be empty"))
	}
	s, err := getSession(handle)
	if err != nil {
		return errorResultGo(err)
	}

	var itm types.Item
	if err := json.Unmarshal([]byte(itemJSON), &itm); err != nil {
		return errorResultGo(fmt.Errorf("parse item JSON: %w", err))
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ensureManagersUnlocked(s); err != nil {
		return errorResultGo(err)
	}

	if err := s.mgr.AddItem(&itm); err != nil {
		return errorResultGo(err)
	}
	return okJSONGo(&itm)
}

func zpUpdateItem(handle int64, id string, itemJSON string) goResult {
	if id == "" {
		return errorResultGo(fmt.Errorf("item ID must not be empty"))
	}
	if itemJSON == "" {
		return errorResultGo(fmt.Errorf("item JSON must not be empty"))
	}
	s, err := getSession(handle)
	if err != nil {
		return errorResultGo(err)
	}

	var updated types.Item
	if err := json.Unmarshal([]byte(itemJSON), &updated); err != nil {
		return errorResultGo(fmt.Errorf("parse item JSON: %w", err))
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ensureManagersUnlocked(s); err != nil {
		return errorResultGo(err)
	}

	cur, err := s.mgr.GetItem(id)
	if err != nil {
		return errorResultGo(err)
	}
	if s.verMgr != nil {
		_ = s.verMgr.SaveVersion(cur)
	}

	updated.ID = id
	updated.CreatedAt = cur.CreatedAt
	updated.Version = cur.Version

	if err := s.mgr.UpdateItem(id, &updated); err != nil {
		return errorResultGo(err)
	}
	return okNoDataGo()
}

func zpDeleteItem(handle int64, id string) goResult {
	if id == "" {
		return errorResultGo(fmt.Errorf("item ID must not be empty"))
	}
	s, err := getSession(handle)
	if err != nil {
		return errorResultGo(err)
	}
	
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ensureManagersUnlocked(s); err != nil {
		return errorResultGo(err)
	}

	if err := s.mgr.DeleteItem(id); err != nil {
		return errorResultGo(err)
	}
	return okNoDataGo()
}
