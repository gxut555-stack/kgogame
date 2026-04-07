package logs

import (
	"fmt"
	"os"
	"sync"
	"testing"
)

func TestBuffer(t *testing.T) {
	if err := os.Mkdir("./test_log_dir", 0755); err != nil {
		t.Fatalf("create temp dir error %s", err)
	}
	GetInstance().Initlog("./test_log_dir/test.log", LOG_DEBUG, 5, 1024*1024*5)
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			Debug("this test run debug")
			Info("this test run info")
			Fatal("this test run fatal")
			//Stack("this test run stack")
			Strace("this test run strace")
			wg.Done()
		}()
	}
	wg.Wait()
	temp := pLogQueue.list
	hasBufferNum := 0
	for temp != nil {
		fmt.Printf("list p=%p \n", temp)
		temp = temp.next
		hasBufferNum++
	}
	fmt.Printf("buffer-len=%d\n", hasBufferNum)
	//删除测试文件
	if err := os.RemoveAll("./test_log_dir"); err != nil {
		t.Fatalf("delete temp dir error %s", err)
	}
}
