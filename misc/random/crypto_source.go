// crypto/rand define
// secure random number generator.
//
// On Linux, FreeBSD, Dragonfly and Solaris, Reader uses getrandom(2) if
// available, /dev/urandom otherwise.
// On OpenBSD and macOS, Reader uses getentropy(2).
// On other Unix-like systems, Reader reads from /dev/urandom.
// On Windows systems, Reader uses the RtlGenRandom API.
// On Wasm, Reader uses the Web Crypto API.

package random

import (
	cRand "crypto/rand"
	"encoding/binary"
	"time"
)

type cryptoSource struct{}

func (s cryptoSource) Seed(seed int64) {}

func (s cryptoSource) Int63() int64 {
	return int64(s.Uint64() & ^uint64(1<<63))
}

func (s cryptoSource) Uint64() (v uint64) {
	err := binary.Read(cRand.Reader, binary.BigEndian, &v)
	if err != nil {
		time.Now().UnixNano()
	}
	return v
}
