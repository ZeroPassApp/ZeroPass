//go:build cgo

package main

import (
	"encoding/json"
	"fmt"
)

type goResult struct {
	code int
	data string
	err  string
}

func okNoDataGo() goResult {
	return goResult{code: zpCodeOK}
}

func okJSONGo(v any) goResult {
	b, err := json.Marshal(v)
	if err != nil {
		return errorResultGo(fmt.Errorf("marshal JSON: %w", err))
	}
	return goResult{code: zpCodeOK, data: string(b)}
}

func errorResultGo(err error) goResult {
	if err == nil {
		return okNoDataGo()
	}
	return goResult{code: codeForError(err), err: err.Error()}
}
