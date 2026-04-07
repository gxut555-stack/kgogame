package common

import (
	"unsafe"
)

// 不安全的将字符串转换成字节流
// 注意这是一个go 1.20版本以上可以使用，低版本使用下面的注释方法,低于1.20版本下面方法会报错
// 同时注意这个一个不安全的方法
func UnsafeStringToBytes(s string) []byte {
	if s == "" {
		return nil
	}
	return unsafe.Slice(unsafe.StringData(s), len(s))
	//说明 go 1.20以下使用以下方法
	///* #nosec G103 */
	//var b []byte
	//bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	///* #nosec G103 */
	//sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	//bh.Data = sh.Data
	//bh.Cap = sh.Len
	//bh.Len = sh.Len
	//return b
}

// 不安全的将字节流转换成字符串
// 注意这是一个go 1.20版本以上可以使用，低版本使用下面的注释方法,低于1.20版本下面方法会报错
// 同时注意这个一个不安全的方法
func UnsafeBytesToString(b []byte) string {
	if b == nil || len(b) == 0 {
		return ""
	}
	return unsafe.String(&b[0], len(b))
	// 低版本使用下面方法
	//return *(*string)(unsafe.Pointer(&b))
}
