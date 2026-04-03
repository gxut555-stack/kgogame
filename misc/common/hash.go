package common

import (
	"hash/crc32"
)

func HashByCrc32(str string) uint32 {
	if str == "" {
		return 0
	}

	n := crc32.ChecksumIEEE([]byte(str))
	return n & 0x7fffffff //去掉最高位符号位
}
