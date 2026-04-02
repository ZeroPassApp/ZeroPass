//go:build cgo

package main

/*
#include "zp_bridge.h"
*/
import "C"

import (
	"encoding/json"
	"fmt"

	"github.com/zeropass/zeropass/core/crypto/password"
	"github.com/zeropass/zeropass/core/vault/health"
)

var pwGen = password.NewGenerator()

//export ZPGeneratePassword
func ZPGeneratePassword(length C.int, optionsJSON *C.char) (res C.ZPResult) {
	defer recoverToResult(&res)

	opts := password.GeneratorOptions{}
	if s := goString(optionsJSON); s != "" {
		if err := json.Unmarshal([]byte(s), &opts); err != nil {
			return errorResult(fmt.Errorf("parse options JSON: %w", err))
		}
	}

	pw, err := pwGen.GenerateRandom(int(length), opts)
	if err != nil {
		return errorResult(err)
	}
	return okJSON(map[string]any{"password": pw})
}

//export ZPGeneratePassphrase
func ZPGeneratePassphrase(words C.int, separator *C.char) (res C.ZPResult) {
	defer recoverToResult(&res)

	sep := goString(separator)
	if sep == "" {
		sep = "-"
	}
	pp, err := pwGen.GeneratePassphrase(int(words), sep)
	if err != nil {
		return errorResult(err)
	}
	return okJSON(map[string]any{"passphrase": pp})
}

//export ZPScorePassword
func ZPScorePassword(passwordStr *C.char) (res C.ZPResult) {
	defer recoverToResult(&res)

	pw := goString(passwordStr)
	result, err := pwGen.ScoreStrength(pw)
	if err != nil {
		return errorResult(err)
	}
	return okJSON(result)
}

//export ZPCheckBreach
func ZPCheckBreach(passwordStr *C.char) (res C.ZPResult) {
	defer recoverToResult(&res)

	pw := goString(passwordStr)
	c := health.NewHIBPClient()
	breached, count, err := c.CheckPassword(pw)
	if err != nil {
		return errorResult(err)
	}
	return okJSON(map[string]any{"breached": breached, "count": count})
}
