/**
 * @Author: Chen Bin
 * @File: mask.go
 * @Date: 2024-11-06 10:48:07
 * @Description:
 */

package ws

import "unsafe"

const wordSize = int(unsafe.Sizeof(uintptr(0)))

// copy github.com/gorilla/websocket/mask.go
func MaskBytes(key [4]byte, pos int, b []byte) int {
	// Mask one byte at a time for small buffers.
	if len(b) < 2*wordSize {
		for i := range b {
			b[i] ^= key[pos&3]
			pos++
		}
		return pos & 3
	}

	// Mask one byte at a time to word boundary.
	if n := int(uintptr(unsafe.Pointer(&b[0]))) % wordSize; n != 0 {
		n = wordSize - n
		for i := range b[:n] {
			b[i] ^= key[pos&3]
			pos++
		}
		b = b[n:]
	}

	// Create aligned word size key.
	var k [wordSize]byte
	for i := range k {
		k[i] = key[(pos+i)&3]
	}
	kw := *(*uintptr)(unsafe.Pointer(&k))

	// Mask one word at a time.
	n := (len(b) / wordSize) * wordSize
	for i := 0; i < n; i += wordSize {
		*(*uintptr)(unsafe.Pointer(uintptr(unsafe.Pointer(&b[0])) + uintptr(i))) ^= kw
	}

	// Mask one byte at a time for remaining bytes.
	b = b[n:]
	for i := range b {
		b[i] ^= key[pos&3]
		pos++
	}

	return pos & 3
}

// 使用for循环 效率较低,不建议生产环境中使用
func maskByFor(key [4]byte, b []byte) {
	length := len(b)
	for i := 0; i < length; i++ {
		b[i] ^= key[i%4]
	}
}
