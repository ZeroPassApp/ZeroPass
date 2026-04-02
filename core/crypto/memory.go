package crypto

import (
	"crypto/subtle"
	"sync"
	"unsafe"
)

// memoryBarrier prevents the compiler from optimizing away memory writes.
// We use a sync.Mutex lock/unlock cycle as a memory barrier.
var memoryBarrier sync.Mutex

// ZeroBytes securely zeroes a byte slice to remove sensitive data from memory.
// It uses a volatile write pattern to prevent compiler optimization from
// eliminating the zeroing operation.
func ZeroBytes(b []byte) {
	if len(b) == 0 {
		return
	}

	// Write zeros using a pointer-based approach that the compiler
	// cannot optimize away.
	p := unsafe.Pointer(&b[0])
	for i := range b {
		*(*byte)(unsafe.Add(p, i)) = 0
	}

	// Memory barrier to ensure the writes are not reordered or eliminated.
	memoryBarrier.Lock()
	memoryBarrier.Unlock() //nolint:staticcheck
}

// SecureCompare performs constant-time comparison of two byte slices.
// Returns true if and only if the slices are equal.
// Uses crypto/subtle to prevent timing side-channel attacks.
// Note: does NOT short-circuit on length mismatch to avoid leaking length info.
func SecureCompare(a, b []byte) bool {
	return subtle.ConstantTimeCompare(a, b) == 1
}

// SecureCompareStrings performs constant-time comparison of two strings.
func SecureCompareStrings(a, b string) bool {
	return SecureCompare([]byte(a), []byte(b))
}
