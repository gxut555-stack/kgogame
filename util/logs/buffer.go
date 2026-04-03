package logs

import (
	"bytes"
	"sync"
)

var pLogQueue *queue

func init() {
	pLogQueue = new(queue)
}

// buffer holds a byte Buffer for reuse. The zero value is ready for use.
type buffer struct {
	bytes.Buffer
	next *buffer
}

const MAX_QUEUE_BUFFER = 100 //最大100个buffer，超过则gc掉

type queue struct {
	// list is a list of byte buffers, maintained under mux.
	list    *buffer
	created int
	mux     sync.Mutex
}

// getBuffer returns a new, ready-to-use buffer.
func (q *queue) getBuffer() *buffer {
	q.mux.Lock()
	defer q.mux.Unlock()
	b := q.list
	if b != nil {
		q.list = b.next
	}
	if b == nil {
		b = new(buffer)
		q.created++
	} else {
		b.next = nil
		b.Reset()
	}
	return b
}

// putBuffer returns a buffer to the free list.
func (q *queue) putBuffer(b *buffer) {
	q.mux.Lock()
	defer q.mux.Unlock()
	//当创建超过最大数量时不再存放
	if b.Len() >= 256 || q.created > MAX_QUEUE_BUFFER {
		// Let big buffers die a natural death.
		q.created--
		q.created = max(0, q.created)
		return
	}
	b.next = q.list
	q.list = b
}

func (q *queue) debug() {
	q.mux.Lock()
	defer q.mux.Unlock()
	num := 0
	t := q.list
	for t != nil {
		num++
		t = t.next
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
