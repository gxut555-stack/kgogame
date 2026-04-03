package logs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"kgogame/base/gDefine"
	"log"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type game_log struct {
	file          *os.File
	nWriteSize    int64 //当前写入大小
	consumer      *WriterConsumer
	filename      string
	level         int
	fileMaxSize   int64
	fileNum       int
	lastWriteTime time.Time //上次写入时间
	//是否基于日期切片
	isSliceDay bool
	dayMaxAge  int //保留几天数据
	mux        sync.Mutex
	fullFnName bool // 是否保留完整函数名
}

const (
	LOG_DEBUG = iota
	LOG_STACK
	LOG_INFO
	LOG_ERROR
	LOG_STRACE
	LOG_FATAL
	LOG_WARN
)

var (
	lvlMap = map[int]string{0: "DEBUG", 1: "STACK", 2: "INFO", 3: "ERROR", 4: "STRACE", 5: "FATAL", 6: "WARN"}
)

var lclog *game_log
var lcOnce sync.Once

func GetInstance() *game_log {
	if lclog == nil {
		lcOnce.Do(func() {
			lclog = new(game_log)
			lclog.consumer = CreateConsumer(WRITE_TYPE_FILE, lclog.write)
		})
	}
	return lclog
}

// 开启基于日期日志
func (l *game_log) SetDayLog(maxAge int) {
	l.mux.Lock()
	defer l.mux.Unlock()
	if maxAge > 0 {
		l.isSliceDay = true
		l.dayMaxAge = maxAge
	} else {
		l.isSliceDay = false
		l.dayMaxAge = 0
	}
}

// 设置保留完整函数名
func (l *game_log) SetFullFnName(val bool) {
	l.mux.Lock()
	defer l.mux.Unlock()
	l.fullFnName = val
}

func (l *game_log) Initlog(file string, level int, filenum int, max int64) {
	var err error
	l.filename = file
	l.file, err = os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Open File Fail : ", err)
	}
	stat, err := l.file.Stat()
	if err != nil {
		log.Fatal("Stat File Fail : ", err)
	}
	l.nWriteSize = stat.Size()
	l.level = level
	l.fileMaxSize = max
	l.lastWriteTime = time.Unix(stat.ModTime().Unix(), 0)
	if l.fileMaxSize == 0 {
		l.fileMaxSize = 1024 * 1024 * 50 //50MB
	}
	l.fileNum = filenum
	if l.fileNum == 0 {
		l.fileNum = 4
	}
}

func (l *game_log) Reload(file string, level int, filenum int, max int64) {
	l.mux.Lock()
	defer l.mux.Unlock()
	if l.filename != file { //文件名替换，先创建新文件，然后替换
		fd, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Printf("reload open file fail : %s", err)
			return
		}
		_ = l.file.Sync()
		_ = l.file.Close()
		l.filename = file
		l.file = fd
	} else {
		l.filename = file
	}
	l.level = level
	l.fileNum = filenum
	l.fileMaxSize = max
}

func (l *game_log) flush() {
	if l.file != nil {
		l.file.Sync()
	}
}

// 提供给接口层实现
func (l *game_log) Debug(format string, v ...interface{}) {
	if LOG_DEBUG < l.level {
		return
	}
	l.prepare(LOG_DEBUG, false, format, v...)
}
func (l *game_log) Info(format string, v ...interface{}) {
	if LOG_INFO < l.level {
		return
	}
	l.prepare(LOG_INFO, false, format, v...)
}
func (l *game_log) Error(format string, v ...interface{}) {
	if LOG_ERROR < l.level {
		return
	}
	l.prepare(LOG_ERROR, false, format, v...)
}

func (l *game_log) prepare(lvl int, isStack bool, format string, args ...interface{}) {
	buf := pLogQueue.getBuffer()
	t := time.Now()
	var file, filename string
	var line int
	var funcname string
	var pc uintptr
	var ok bool
	pc, file, line, ok = runtime.Caller(2)
	filename = file
	if !ok {
		file = "???"
		line = 0
		funcname = "UnKnow.UnKnow"
	} else {
		funcptr := runtime.FuncForPC(pc)
		funcname = funcptr.Name()
		n := strings.LastIndexAny(file, "/")
		file = file[n+1:]
		w := strings.LastIndexAny(funcname, "/")
		funcname = funcname[w+1:]
		//调整如下，保留主函数，即存储更多信息
		//w = strings.Index(funcname, ".")
		//funcname = funcname[w+1:]
	}

	timeStr := t.Format("2006-01-02 MST 15:04:05.000")
	buf.WriteString(fmt.Sprintf("[%-6s %s] [%s:%s:%d] ", lvlMap[lvl], timeStr,
		file, funcname, line)) // ignore err
	fmt.Fprintf(buf, format, args...)
	if buf.Bytes()[buf.Len()-1] != '\n' {
		buf.WriteByte('\n')
	}
	if isStack {
		buf.Write(stacks())
	}
	if lvl == LOG_FATAL {
		//report(buf.Bytes())
		reportLog(&LogInfo{
			Level:       int32(lvl),
			LogTime:     t.Unix(),
			Msg:         fmt.Sprintf(format, args...),
			Filename:    filename,
			Line:        int32(line),
			ServiceName: serviceName,
		})
	}
	//l.write(buf)
	l.consumer.Send(buf)
}

func (l *game_log) write(buf *buffer) {
	/*----------------此处锁只有用在reload，目前只有一个协程在单独写----------------*/
	l.mux.Lock()
	defer l.mux.Unlock()
	t := time.Now()
	if l.isSliceDay && !isSameDay(&l.lastWriteTime, &t) {
		if l.newfile(&l.lastWriteTime) == false { //新建文件不存在，可能由于磁盘已满，此时不写入,丢弃掉本次写入
			l.lastWriteTime = t //避免一次检测继续创建新文件
			return
		}
		go l.deleteMaxDayFile()
	}
	if l.nWriteSize > l.fileMaxSize {
		if l.newfile(&l.lastWriteTime) == false { //新建文件不存在，可能由于磁盘已满，此时不写入,丢弃掉本次写入
			return
		}
	}
	l.lastWriteTime = t

	//此处写入会存在写入乱序的情况，取决于内存写入问题，一般情况下在prepare情况下加锁操作后，
	//会导致后进的协程写入晚于之前协程写入，主要在于Lock其实是一个耗时过程，必然导致操作慢与当前协程
	//但问题还是有可能存在,问题在于锁--->写入这段时间不能明确执行速率的问题
	n, err := l.file.Write(buf.Bytes())
	if err != nil {
		log.Println("Open File Fail : ", err)
		return
	}
	//atomic.AddInt64(&l.nWriteSize,int64(n)) //可以使用此方法保证写入原子性，但大小准确性要求不高
	l.nWriteSize += int64(n)

}

func fileExist(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func (l *game_log) newfile(t *time.Time) bool {
	if l.file != nil {
		l.file.Close()
		year, month, day := t.Date()
		newYear, newMonth, newDay := time.Now().Date()
		newFileName := fmt.Sprintf("%s.%d", l.filename, 1)
		if l.isSliceDay {
			newFileName = fmt.Sprintf("%s.%d-%02d-%02d.%d", l.filename, newYear, newMonth, newDay, 1)
		}
		var oldFileName, oldNewFileName string
		for i := l.fileNum; i > 0; i-- {
			if l.isSliceDay {
				oldFileName = fmt.Sprintf("%s.%d-%02d-%02d.%d", l.filename, year, month, day, i)
			} else {
				oldFileName = fmt.Sprintf("%s.%d", l.filename, i)
			}
			if fileExist(oldFileName) && i == l.fileNum {
				os.Remove(oldFileName)
			} else if fileExist(oldFileName) {
				if l.isSliceDay {
					oldNewFileName = fmt.Sprintf("%s.%d-%02d-%02d.%d", l.filename, year, month, day, i+1)
				} else {
					oldNewFileName = fmt.Sprintf("%s.%d", l.filename, i+1)
				}
				os.Rename(oldFileName, oldNewFileName)
			}
		}
		os.Rename(l.filename, newFileName)
	}
	var err error
	l.file, err = os.OpenFile(l.filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Open File Fail : ", err) //创建文件失败后，不主动exit程序
		l.nWriteSize = 0
		l.file = nil
		return false
	}
	l.nWriteSize = 0
	return true
}

func getDayZero() int64 {
	timeStr := time.Now().Format("2006-01-02")
	t, _ := time.ParseInLocation("2006-01-02 15:04:05", timeStr+" 00:00:00", time.Local)
	return t.Unix()
}

func getDayUnix(date string) int64 {
	t, err := time.ParseInLocation("2006-01-02 15:04:05", date+" 00:00:00", time.Local)
	if err != nil {
		log.Printf("trans date[%s] err =%s", date, err)
		return 0
	}
	return t.Unix()
}

// 删除大于MaxAge天数的文件
func (l *game_log) deleteMaxDayFile() {
	now := getDayZero()
	delTime := now - 86400*int64(l.dayMaxAge)
	dir, err := filepath.Abs(filepath.Dir(l.filename))
	if err != nil {
		log.Printf("read dir error=%s", err)
		return
	}
	files, _ := ioutil.ReadDir(dir)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fileName := dir + "/" + file.Name()
		if fileName == l.filename {
			continue
		}
		date := getDate(fileName)
		if strings.LastIndex(date, "-") == -1 {
			continue
		}
		fileUnixTime := getDayUnix(date)
		if fileUnixTime == 0 {
			continue
		}
		if fileUnixTime < delTime {
			os.Remove(fileName)
		}
	}
}

func getDate(path string) string {
	index := 0
	for i := len(path) - 1; i >= 0 && path[i] != '/'; i-- {
		if path[i] == '.' {
			if index != 0 {
				return path[i+1 : index]
			} else {
				index = i
			}
		}
	}
	return ""
}

// 输出日志到文件中，DEBUG级别, format为格式化配置
func Debug(format string, v ...interface{}) {
	if !isInitLogOnce {
		InitLog("Uninitialized", 1, nil)
	}
	l := GetInstance()
	if LOG_DEBUG < l.level {
		return
	}
	l.prepare(LOG_DEBUG, false, format, v...)
}

func Info(format string, v ...interface{}) {
	if !isInitLogOnce {
		InitLog("Uninitialized", 1, nil)
	}
	l := GetInstance()
	if LOG_INFO < l.level {
		return
	}
	l.prepare(LOG_INFO, false, format, v...)
}

func Error(format string, v ...interface{}) {
	if !isInitLogOnce {
		InitLog("Uninitialized", 1, nil)
	}
	l := GetInstance()
	if LOG_ERROR < l.level {
		return
	}
	l.prepare(LOG_ERROR, false, format, v...)
}

func Strace(format string, v ...interface{}) {
	if !isInitLogOnce {
		InitLog("Uninitialized", 1, nil)
	}
	l := GetInstance()
	if LOG_STRACE < l.level {
		return
	}
	l.prepare(LOG_STRACE, false, format, v...)
}

func Fatal(format string, v ...interface{}) {
	if !isInitLogOnce {
		InitLog("Uninitialized", 1, nil)
	}
	l := GetInstance()
	if LOG_FATAL < l.level {
		return
	}
	l.prepare(LOG_FATAL, false, format, v...)
}

func Stack(format string, v ...interface{}) {
	if !isInitLogOnce {
		InitLog("Uninitialized", 1, nil)
	}
	l := GetInstance()
	if LOG_STACK < l.level {
		return
	}
	l.prepare(LOG_STACK, true, format, v...)
}

func Flush() {
	l := GetInstance()
	l.flush()
}

// stacks is a wrapper for runtime.Stack that attempts to recover the data for all goroutines.
func stacks() []byte {
	// We don't know how big the traces are, so grow a few times if they don't fit. Start large, though.
	n := 10000
	var trace []byte
	for i := 0; i < 5; i++ {
		trace = make([]byte, n)
		nbytes := runtime.Stack(trace, false)
		if nbytes < len(trace) {
			return trace[:nbytes]
		}
		n *= 2
	}
	return trace
}

// 在main的开始引入，其他goroutine中也需要引入！
func CatchPanic() {
	//panic异常处理
	if r := recover(); r != nil {
		var s = string(debug.Stack())
		fmt.Printf("panic:%#v stack=%s \n\r", r, s)
		Fatal("panic:%#v stack=%s", r, s)
		//TODO:告警！
	}
}

// 是否为同一天
func isSameDay(a, b *time.Time) bool {
	aYear, aMonth, aDay := a.Date()
	bYear, bMonth, bDay := b.Date()
	if aYear == bYear && aMonth == bMonth && aDay == bDay {
		return true
	}
	return false
}

var pUdpSocket *net.UDPConn
var serviceName string

func SetUdpReport(addr string, name string) {
	if addr == "" {
		addr = "127.0.0.1:4150"
	}

	init_udp(addr)
	serviceName = name
	if serviceName == "" {
		serviceName = "unknown"
	}
}

func init_udp(addr string) {
	var e error
	dstAddr, e := net.ResolveUDPAddr("udp", addr)
	if e != nil {
		log.Printf("net.ResolveUDPAddr(udp, %s) error: %s \n", addr, e.Error())
		return
	}
	pUdpSocket, e = net.DialUDP("udp", nil, dstAddr)
	if e != nil {
		log.Printf("init udp socket err %s\n", e)
		return
	}
	//设置缓存区buffer
	_ = pUdpSocket.SetWriteBuffer(1024 * 128) //128K，最大128Kbuffer，超过则写入导致粘包，对端解析不过来
}

const MAX_SYNC = 10
const (
	REPORT_UNLOCK = 0
	REPORT_LOCK   = 1
)

var reportSeq uint32
var reportLock [MAX_SYNC]uint32 //控制并发数量，确保并发量的情况下，拒绝发送消息

func report(b []byte) {
	if len(b) == 0 || pUdpSocket == nil {
		return
	}
	seq := atomic.AddUint32(&reportSeq, 1)
	idx := seq % MAX_SYNC
	//去掉尝试2次逻辑，因为此处本身是控制并发，在尝试2次情况下本身就提高了并发数量
	//故此处只遍历一次，控制并发量，保证远端服务能够正常运行，同时避免udp洪冲击导致丢包率提升
	//此处控制并发逻辑在于send_udp的速度，假设send_udp的耗时为1ms，则在同一毫秒情况下，最多只允许10次发送请求
	//理论上send_udp的发送时间小于1ms，因为发送udp只是从用户态数据拷贝到内核态去即返回成功，在没有缓冲区塞满的情况下是小于1ms的
	for i := 0; i < 1; i++ {
		if atomic.CompareAndSwapUint32(&reportLock[idx], REPORT_UNLOCK, REPORT_LOCK) {
			send_udp(b)
			atomic.CompareAndSwapUint32(&reportLock[idx], REPORT_LOCK, REPORT_UNLOCK) //unlock
			break
		} else {
			log.Printf("lock-idx=%d is lock cannot send msg,maybe fatal has flood dump,ignore this msg\n", idx)
		}
	}
	if seq > 0x0fffffff {
		atomic.CompareAndSwapUint32(&reportSeq, seq, 0) //此处不一定能成功，但预留了足够长的区间去尝试
	}
}

// 发送UDP消息
func send_udp(msg []byte) {
	//写操作会用到锁机制保护数据流的顺序，不会导致随机写入
	_, e := pUdpSocket.Write([]byte(gDefine.NSQ_TOPIC_SERVICE_ERROR_LOG + "|" + string(msg)))
	if e != nil {
		log.Printf("send udp msg error %s\n", e)
	}
}

// 日志信息
type LogInfo struct {
	Level       int32  `json:"level"`
	LogTime     int64  `json:"log_time"`
	Msg         string `json:"msg"`
	Filename    string `json:"filename"`
	Line        int32  `json:"line"`
	ServiceName string `json:"service_name"`
}

// udp报送
func reportLog(info *LogInfo) {
	b, err := json.Marshal(info)
	if err != nil {
		log.Printf("json.Marshal(%#v) err: %s", info, err)
		return
	}
	report(b)
}

// 去除掉udp上报
func HackUdpIsNil() {
	pUdpSocket = nil
}
