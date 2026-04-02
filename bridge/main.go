//go:build cgo

package main

/*
#include "zp_bridge.h"
#include <string.h>
*/
import "C"

import "unsafe"

//export ZPFree
func ZPFree(ptr *C.char) {
	if ptr == nil {
		return
	}
	// These strings are created via C.CString (malloc + NUL-terminated). Wipe
	// them before free to reduce secret remanence in the C heap.
	n := C.strlen(ptr)
	if n > 0 {
		C.memset(unsafe.Pointer(ptr), 0, n)
	}
	C.free(unsafe.Pointer(ptr))
}

//export ZPFreeResult
func ZPFreeResult(r C.ZPResult) {
	if r.data != nil {
		ZPFree(r.data)
	}
	if r.error != nil {
		ZPFree(r.error)
	}
}

//export ZPFreeResultPtr
func ZPFreeResultPtr(r *C.ZPResult) {
	if r == nil {
		return
	}
	if r.data != nil {
		ZPFree(r.data)
		r.data = nil
	}
	if r.error != nil {
		ZPFree(r.error)
		r.error = nil
	}
	r.code = 0
}

func main() {}
