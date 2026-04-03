package joy

import (
	"errors"
	"kgogame/util/logs"
	"log"
	"sync"
	"sync/atomic"
)

type Task[T comparable] struct {
	id  int32
	arg T //参数
	fn  func(arg T)
}

type CoroutinePool[T comparable] struct {
	id       int32
	cnt      int //协程数量
	stopChan chan struct{}
	taskChan chan *Task[T]
	workerWg sync.WaitGroup
	remainWg sync.WaitGroup
	isStop   int32 //是否停止
}

func NewCoroutinePool[T comparable](n int) *CoroutinePool[T] {
	if n > 1024 || n <= 0 {
		n = 1024
	}
	one := &CoroutinePool[T]{
		id:       1,
		cnt:      n,
		stopChan: make(chan struct{}),
		taskChan: make(chan *Task[T], 1000),
		isStop:   0,
	}
	for i := 0; i < n; i++ {
		one.workerWg.Add(1)
		go one.co()
	}
	return one
}

func (this *CoroutinePool[T]) co() {
	defer this.workerWg.Done()
	for {
		select {
		case <-this.stopChan:
			return
		case task := <-this.taskChan:
			log.Printf("co=====> taskId:%d arg:%+v", task.id, task.arg)
			task.fn(task.arg)
		}
	}
}

func (this *CoroutinePool[T]) getTaskId() int32 {
	id := atomic.AddInt32(&this.id, 1)
	if id > 0x7FFFFFFF {
		atomic.StoreInt32(&this.id, 1)
	}
	return id
}

// 保证先Stop后不能执行put
func (this *CoroutinePool[T]) Put(arg T, fn func(arg T)) error {
	if atomic.LoadInt32(&this.isStop) == 1 {
		return errors.New("CoroutinePool has been stopped")
	}
	defer logs.CatchPanic()
	task := &Task[T]{
		id:  this.getTaskId(),
		arg: arg,
		fn:  fn,
	}
	this.taskChan <- task
	return nil
}

func (this *CoroutinePool[T]) Stop() {
	atomic.CompareAndSwapInt32(&this.isStop, 0, 1) //设置停止标志
	close(this.stopChan)                           //关闭所有工作协程
	this.workerWg.Wait()                           //等待所有工作协程退出
	this.remainWg.Add(1)                           //交给等待任务处理剩余任务
	go this.doRemainTask()                         //处理剩余任务
	close(this.taskChan)                           //关闭任务通道
	this.remainWg.Wait()                           //等待任务完成
}

func (this *CoroutinePool[T]) doRemainTask() {
	defer this.remainWg.Done()
	for {
		task, ok := <-this.taskChan
		if !ok {
			break
		}
		task.fn(task.arg)
		log.Printf("remain taskId:%d arg:%+v", task.id, task.arg)
	}
}
