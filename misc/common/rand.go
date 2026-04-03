package common

import (
	cRand "crypto/rand"
	"encoding/binary"
	"math/rand"
	"time"
)

func NewRand() *rand.Rand {
	seed := time.Now().UnixNano()
	_ = binary.Read(cRand.Reader, binary.BigEndian, &seed)
	return rand.New(rand.NewSource(seed))
}
