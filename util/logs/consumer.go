package logs

import (
	"sync"
)

const (
	WRITE_TYPE_FILE = 1
	WRITE_TYPE_UDP  = 2
	WRITE_TYPE_TCP  = 3
)

// 写入消费者
type WriterConsumer struct {
	Type        int
	buf         *buffer
	backupBuf   *buffer
	notifyChan  chan struct{}
	doWriteFunc func(b *buffer) //执行写入
	mux         sync.Mutex
}

func CreateConsumer(tp int, f func(b *buffer)) *WriterConsumer {
	w := &WriterConsumer{
		Type:        tp,
		buf:         nil,
		backupBuf:   nil,
		notifyChan:  make(chan struct{}), //使用空结构是避免内存复制
		doWriteFunc: f,
		mux:         sync.Mutex{},
	}
	go w.channel() //写入渠道
	return w
}

func (w *WriterConsumer) channel() {
	for {
		_, ok := <-w.notifyChan
		if !ok {
			break
		}

		if w.buf != nil {
			w.doWriteFunc(w.buf)
			pLogQueue.putBuffer(w.buf)
			w.buf = nil
		} else {
			w.doWriteFunc(w.backupBuf)
			pLogQueue.putBuffer(w.backupBuf)
			w.backupBuf = nil
		}
	}
}

func (w *WriterConsumer) Send(b *buffer) {
	w.mux.Lock()
	defer w.mux.Unlock()
	if w.buf == nil {
		w.buf = b
	} else {
		w.backupBuf = b
	}
	w.notifyChan <- struct{}{} //通知消费
}
