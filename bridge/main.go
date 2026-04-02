//go:build cgo

package main

/*
#include "zp_bridge.h"
*/
import "C"

import "unsafe"

//export ZPFree
func ZPFree(ptr *C.char) {
	if ptr == nil {
		return
	}
	C.free(unsafe.Pointer(ptr))
}

//export ZPFreeResult
func ZPFreeResult(r C.ZPResult) {
	if r.data != nil {
		C.free(unsafe.Pointer(r.data))
	}
	if r.error != nil {
		C.free(unsafe.Pointer(r.error))
	}
}

//export ZPFreeResultPtr
func ZPFreeResultPtr(r *C.ZPResult) {
	if r == nil {
		return
	}
	if r.data != nil {
		C.free(unsafe.Pointer(r.data))
		r.data = nil
	}
	if r.error != nil {
		C.free(unsafe.Pointer(r.error))
		r.error = nil
	}
	r.code = 0
}

func main() {}
