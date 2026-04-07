package gowarp

import "sync/atomic"

var CntOfGoroutine *int32

// 初始化 -- 一个进程只允许存在一个实例
func Init(cnt int32) {
	if CntOfGoroutine != nil {
		panic("CntOfGoroutine already initialized")
	}
	cpy := cnt
	CntOfGoroutine = &cpy
}

// 开启协程时加个计数，以便停止时等待处理
func GoWrap(fn func()) {
	if CntOfGoroutine == nil {
		panic("CntOfGoroutine not initialized")
	}
	atomic.AddInt32(CntOfGoroutine, 1)
	defer atomic.AddInt32(CntOfGoroutine, -1)
	fn()
}

// 获取运行的协程数
func GetRunningGoroutine() int32 {
	return atomic.LoadInt32(CntOfGoroutine)
}
