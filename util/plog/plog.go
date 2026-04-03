package plog

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"kgogame/util/gconf"
	"kgogame/util/gcontext"
	"kgogame/util/getConfig"
	"kgogame/util/plog/lumberjack"
	"log"
	"net"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var nullLogger = log.New(ioutil.Discard, "", log.LstdFlags)
var (
	Lzap             *zap.Logger
	LumberjackLogger *lumberjack.Logger //加一个包级别的Lumberjack.Logger对象，方便重置
	MaxSize          = 1
	zapLevel         zap.AtomicLevel
	ini_parser       getConfig.IniParser = getConfig.IniParser{}
	fileEncoder      zapcore.Encoder     //编码
	once             sync.Once
)

/*
type NsqLogger struct {
	pd     *nsq.Producer
	ch_log chan []byte
	//mu   sync.Mutex
	lulogger *lumberjack.Logger
}

//NsqLogger 实现 io.Writer interface
func (l *NsqLogger) Write(p []byte) (int, error) {
	//l.mu.Lock()
	//defer l.mu.Unlock()
	//produce something
	var buffer bytes.Buffer
	buffer.Write(p)
	l.ch_log <- buffer.Bytes()
	return len(p), nil

}
*/

func NewEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		// Keys can be anything except the empty string.
		TimeKey:        "T",
		LevelKey:       "L",
		NameKey:        "N",
		CallerKey:      "C",
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

func TimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006/01/02 15:04:05"))
}

func TimeEncoderMs(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006/01/02 15:04:05.000"))
}

// 日志文件名称和存放地址
func getCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	dir = strings.Replace(dir, "\\", "/", -1)

	//本地文件的文件名称
	var name string
	if gconf.CmdConf.ServerID > 0 {
		name = filepath.Base(os.Args[0]) + strconv.Itoa(int(gconf.CmdConf.ServerID)) + "-lumberjack.log"
	} else {
		name = filepath.Base(os.Args[0]) + "-lumberjack.log"
	}

	return filepath.Join(dir, name)

}

// 检查ip端口是否正常
func hostAddrCheck(addr string) bool {
	items := strings.Split(addr, ":")
	if items == nil || len(items) != 2 {
		return false
	}

	a := net.ParseIP(items[0])
	if a == nil {
		return false
	}

	match, err := regexp.MatchString("^[0-9]*$", items[1])
	if err != nil {
		return false
	}

	//i,err:=StringToInt64(items[1])
	i, err := strconv.ParseInt(items[1], 10, 64)
	if err != nil {
		return false
	}
	if i < 0 || i > 65535 {
		return false
	}

	if match == false {
		return false
	}

	return true
}

// 获取日志级别
func getLogLevel(lvl string) zapcore.LevelEnabler {
	switch lvl {
	case "debug":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	case "warn":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	default:
		return zap.DebugLevel
	}
}

// 日志库的初始化
func Init() {
	once.Do(func() {
		//读取logConfig.ini配置文件

		var cfgFile string = "logConf.ini"

		//Nsq IP
		var addr string = "127.0.0.1:4150"
		//Log topic
		var topic string
		//log level
		var lvl string = "debug"

		//log dir
		var log_local_file string

		//log size
		var log_max_size int

		var writer io.Writer

		//默认的topic值
		if gconf.CmdConf.ServerName == "" {
			topic = path.Base(os.Args[0]) + "_plog"
		} else {
			topic = gconf.CmdConf.ServerName + "_plog"
		}
		//使用配置文件中的值
		if gconf.CmdConf.PLogConf != "" {
			cfgFile = gconf.CmdConf.PLogConf
		}
		if err := ini_parser.Load(cfgFile); err != nil {
			fmt.Printf("try load config error[%s]\n", err.Error())
		} else {
			addr = ini_parser.GetString("", "alog_addr")
			topic = ini_parser.GetString("", "topic")
			log_local_file = ini_parser.GetString("", "local_file")
			log_max_size = int(ini_parser.GetInt32("", "max_size"))

			// 做个兼容，大鱼森林不想搞那么多配置文件
			if topic[0:3] == "bf_" {
				topic = strings.Replace(topic, "{$SRV_NAME}", gconf.CmdConf.ServerName, -1)
				topic = strings.Replace(topic, "{$SRV_ID}", fmt.Sprintf("%d", gconf.CmdConf.ServerID), -1)
			}
			//如果最终的日志文件能够按照svrId区分出来，那么就要把svrid拼接到topic中
			//topic = fmt.Sprint("%s%d",topic,gconf.CmdConf.ServerID)
			lvl = ini_parser.GetString("", "level")
		}

		//fmt.Println("ip, topic", ip, topic)
		//fmt.Println("getCurrentDirectory", getCurrentDirectory())

		if addr == "" {
			//ip = "127.0.0.1:4150"
		} else if !hostAddrCheck(addr) {
			fmt.Println("incorrect IP:PORT addr")
			addr = ""
		}
		if topic == "" {
			topic = "test"
		}

		//初始化Nsq配置
		//config := nsq.NewConfig()
		//p, err := nsq.NewProducer(addr, config)
		//if err != nil {
		//	fmt.Println(err)
		//}
		//p.SetLogger(nullLogger, 0)

		var (
			conf_file_name string
			conf_max_size  int
		)

		//设置LumberjackLogger 参数
		if log_local_file != "" {
			conf_file_name = log_local_file
		} else {
			conf_file_name = getCurrentDirectory()
		}

		if log_max_size != 0 {
			conf_max_size = log_max_size
		} else {
			conf_max_size = MaxSize
		}

		//fmt.Println("conf info:", conf_file_name, conf_max_size)

		LumberjackLogger = &lumberjack.Logger{
			Filename: conf_file_name,
			MaxSize:  conf_max_size,
		}
		writer = LumberjackLogger

		//nsqlog := &NsqLogger{
		//	pd:       p,
		//	ch_log:   make(chan []byte, 100),
		//	lulogger: LumberjackLogger,
		//}
		if addr != "" {
			if udplogger, err := CreateLoggerFromAddr(addr, topic); err != nil {
				fmt.Printf("udp addr error: %s\n", err.Error())
			} else {
				writer = udplogger
				//fmt.Printf("udp logger used\n")
			}
		} else {
			//fmt.Printf("udp addr is empty\n")
		}

		//用Zap格式文件库
		zapLevel = zap.NewAtomicLevelAt(getLogLevel(lvl).(zapcore.Level))
		w := zapcore.AddSync(writer)
		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(NewEncoderConfig()),
			//zapcore.NewConsoleEncoder(NewEncoderConfig()),
			w,
			//zap.DebugLevel,
			zapLevel,
		)

		//Lzap = zap.New(core, zap.AddCaller())
		Lzap = zap.New(core)

		//	ch_logfile := make(chan []byte, 100)

		//启动5个goroutine publish log到nsq
		/*
			for i := 0; i < 1; i++ {
				go func(l *NsqLogger, topicName string, goi int) {
					for {
						select {
						case s := <-l.ch_log:
							//如果没有设置Nsq的ip，则写本地文件
							if ip == "" {
								//lumberjack 是协程安全的，不用加锁
								_, err = l.lulogger.Write(s)
								if err != nil {
									fmt.Println("error %s", err)
								}
							} else {
								//如果设置了Nsq的ip，则写远程消息队列
								//Publish 是协程安全，不用加锁
								err := l.pd.Publish(topicName, s)
								if err != nil {
									//t.Fatalf("error %s", err)
									fmt.Println("error %s", err)
									l.lulogger.Write([]byte(err.Error()))
									l.lulogger.Write(s)

									//close(stopChan)
									//Publish 失败了就写文件
									//ch_logfile <- s
									//break
								}
							}
						}
					}
				}(nsqlog, topic, i)
			}
		*/

		//文件编码
		fileEncoder = zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	})
}

// lumberjack logger重置一次，让不同同类型的svr按照svrid打到不同的日志中
func ReInit() {
	Init()
	LumberjackLogger.Filename = getCurrentDirectory()
}

// 获取日志调用点的文件和行数信息
func getCaller(deep int) (string, int, string) {
	pcs := make([]uintptr, 1)
	if n := runtime.Callers(deep+1, pcs); n == 0 {
		return "???", 0, ""
	} else {
		frames := runtime.CallersFrames(pcs)
		frame, _ := frames.Next()
		return frame.File, frame.Line, frame.Function
	}
}

// Deprecated: Info 改为使用logs 库打日志
func Info(v ...interface{}) {
	Init()
	file, line, _ := getCaller(2)
	Lzap.Info(fmt.Sprint(v...), zap.String("FI", file), zap.Int("LN", line))
}

// Deprecated: VInfo 改为使用logs 库打日志
func VInfo(format string, t ...interface{}) {
	Init()
	file, line, _ := getCaller(2)
	Lzap.Info(fmt.Sprintf(format, t...), zap.String("FI", file), zap.Int("LN", line))
}

// Deprecated: Debug 改为使用logs 库打日志
// 说明：添加\x1b[34m 和 \x1b[0m\n 是为了方便颜色的展示！
func Debug(v ...interface{}) {
	Init()
	file, line, _ := getCaller(2)
	Lzap.Debug(fmt.Sprint(v...), zap.String("FI", file), zap.Int("LN", line))
}

// Deprecated: VDebug 改为使用logs 库打日志
func VDebug(format string, t ...interface{}) {
	Init()
	file, line, _ := getCaller(2)
	Lzap.Debug(fmt.Sprintf(format, t...), zap.String("FI", file), zap.Int("LN", line))
}

// Deprecated: Warn 改为使用logs 库打日志
func Warn(v ...interface{}) {
	Init()
	file, line, _ := getCaller(2)
	Lzap.Warn(fmt.Sprint(v...), zap.String("FI", file), zap.Int("LN", line))
}

// Deprecated: VWarn 改为使用logs 库打日志
func VWarn(format string, t ...interface{}) {
	Init()
	file, line, _ := getCaller(2)
	Lzap.Warn(fmt.Sprintf(format, t...), zap.String("FI", file), zap.Int("LN", line))
}

// Deprecated: Fatal 改为使用logs 库打日志
func Fatal(v ...interface{}) {
	Init()
	file, line, _ := getCaller(2)
	Lzap.Error(fmt.Sprint(v...), zap.String("FI", file), zap.Int("LN", line))
}

// Deprecated: VFatal 改为使用logs 库打日志
func VFatal(format string, t ...interface{}) {
	Init()
	file, line, _ := getCaller(2)
	Lzap.Error(fmt.Sprintf(format, t...), zap.String("FI", file), zap.Int("LN", line))
}

// Game专用
func VInfoGame(format string, deep int, t ...interface{}) {
	Init()
	file, line, _ := getCaller(deep)
	Lzap.Info(fmt.Sprintf(format, t...), zap.String("FI", file), zap.Int("LN", line))
}
func VDebugGame(format string, deep int, t ...interface{}) {
	Init()
	file, line, _ := getCaller(deep)
	Lzap.Debug(fmt.Sprintf(format, t...), zap.String("FI", file), zap.Int("LN", line))
}
func VWarnGame(format string, deep int, t ...interface{}) {
	Init()
	file, line, _ := getCaller(deep)
	Lzap.Warn(fmt.Sprintf(format, t...), zap.String("FI", file), zap.Int("LN", line))
}
func VFatalGame(format string, deep int, t ...interface{}) {
	Init()
	file, line, _ := getCaller(deep)
	Lzap.Error(fmt.Sprintf(format, t...), zap.String("FI", file), zap.Int("LN", line))
}

// Deprecated: CatchPanic 改为使用logs.CatchPanic
func CatchPanic() {
	Init()
	//panic异常处理
	if r := recover(); r != nil {
		var s string = string(debug.Stack())
		fmt.Printf("panic:%#v stack=%s \n\r", r, s)
		Fatal(r, s)
		//TODO:告警！
	}
}

// 重新加载配置
func ReloadIniConfig() {
	Init()
	if err := ini_parser.Reload(); err == nil { //重载成功
		if lvl := ini_parser.GetString("", "level"); lvl != "" { //值不为空
			SetLogLevel(lvl)
		}
	}
}

// 设置日志的级别
func SetLogLevel(lvl string) {
	Init()
	if lvlNew := getLogLevel(lvl).(zapcore.Level); lvlNew != zapLevel.Level() {
		zapLevel.SetLevel(lvlNew)
	}
}

//---------------

// Deprecated: SInfo 改为使用logs 库打日志
// 带调用函数名的Info
func SInfo(format string, t ...interface{}) {
	Init()
	file, line, fn := getCaller(2)
	Lzap.Info(fmt.Sprintf(fn+"|"+format, t...), zap.String("FI", file), zap.Int("LN", line))
}

// Deprecated: SDebug 改为使用logs 库打日志
// 带调用函数名的Debug
func SDebug(format string, t ...interface{}) {
	Init()
	file, line, fn := getCaller(2)
	Lzap.Debug(fmt.Sprintf(fn+"|"+format, t...), zap.String("FI", file), zap.Int("LN", line))
}

// Deprecated: SWarn 改为使用logs 库打日志
// 带调用函数名的Warn
func SWarn(format string, t ...interface{}) {
	Init()
	file, line, fn := getCaller(2)
	Lzap.Warn(fmt.Sprintf(fn+"|"+format, t...), zap.String("FI", file), zap.Int("LN", line))
}

// Deprecated: SFatal 改为使用logs 库打日志
// 带调用函数名的Fatal
func SFatal(format string, t ...interface{}) {
	Init()
	file, line, fn := getCaller(2)
	Lzap.Error(fmt.Sprintf(fn+"|"+format, t...), zap.String("FI", file), zap.Int("LN", line))
}

// 文件型日志结构定义
type FileLogger struct {
	stripFilePrefix string
	stripFuncPrefix string
	zapLogger       *zap.Logger
	*zap.SugaredLogger
}

// 新实例
func NewFileLogger(w *lumberjack.Logger) *FileLogger {
	Init()
	logger := zap.New(zapcore.NewCore(fileEncoder, zapcore.AddSync(w), zapLevel))
	return &FileLogger{
		zapLogger:     logger,
		SugaredLogger: logger.Sugar(),
	}
}

// 设置日志级别
func SetFileLoggerLevel(lvl string) {
	zapLevel = zap.NewAtomicLevelAt(getLogLevel(lvl).(zapcore.Level))
}

// 设置文件名需要去掉的前缀部分
func (f *FileLogger) SetStripFilePrefix(prefix string) {
	f.stripFilePrefix = prefix
}

// 设置函数名需要去掉的前缀部分
func (f *FileLogger) SetStripFuncPrefix(prefix string) {
	f.stripFuncPrefix = prefix
}

// 带调用函数名的Info
func (f *FileLogger) SInfo(format string, t ...interface{}) {
	file, line, fn := getCaller(2)
	if f.stripFilePrefix != "" {
		file = strings.TrimPrefix(file, f.stripFilePrefix)
	}
	if f.stripFuncPrefix != "" {
		fn = strings.TrimPrefix(fn, f.stripFuncPrefix)
	}
	f.Infof(fmt.Sprintf("%s:%d\t%s\t|%s", file, line, fn, format), t...)
}

// 带调用函数名的Debug
func (f *FileLogger) SDebug(format string, t ...interface{}) {
	file, line, fn := getCaller(2)
	if f.stripFilePrefix != "" {
		file = strings.TrimPrefix(file, f.stripFilePrefix)
	}
	if f.stripFuncPrefix != "" {
		fn = strings.TrimPrefix(fn, f.stripFuncPrefix)
	}
	f.Debugf(fmt.Sprintf("%s:%d\t%s\t|%s", file, line, fn, format), t...)
}

// 带调用函数名的Warn
func (f *FileLogger) SWarn(format string, t ...interface{}) {
	file, line, fn := getCaller(2)
	if f.stripFilePrefix != "" {
		file = strings.TrimPrefix(file, f.stripFilePrefix)
	}
	if f.stripFuncPrefix != "" {
		fn = strings.TrimPrefix(fn, f.stripFuncPrefix)
	}
	f.Warnf(fmt.Sprintf("%s:%d\t%s\t|%s", file, line, fn, format), t...)
}

// 带调用函数名的Fatal
func (f *FileLogger) SFatal(format string, t ...interface{}) {
	file, line, fn := getCaller(2)
	if f.stripFilePrefix != "" {
		file = strings.TrimPrefix(file, f.stripFilePrefix)
	}
	if f.stripFuncPrefix != "" {
		fn = strings.TrimPrefix(fn, f.stripFuncPrefix)
	}
	f.Errorf(fmt.Sprintf("%s:%d\t%s\t|%s", file, line, fn, format), t...)
}

func (f *FileLogger) Info(v ...interface{}) {
	file, line, _ := getCaller(2)
	f.Infof("%s:%d\t|%s", file, line, fmt.Sprint(v...))
}

func (f *FileLogger) VInfo(format string, t ...interface{}) {
	file, line, _ := getCaller(2)
	f.Infof("%s:%d\t|%s", file, line, fmt.Sprintf(format, t...))
}

// 说明：添加\x1b[34m 和 \x1b[0m\n 是为了方便颜色的展示！
func (f *FileLogger) Debug(v ...interface{}) {
	file, line, _ := getCaller(2)
	f.Debugf("%s:%d\t|%s", file, line, fmt.Sprint(v...))
}

func (f *FileLogger) VDebug(format string, t ...interface{}) {
	file, line, _ := getCaller(2)
	f.Debugf("%s:%d\t|%s", file, line, fmt.Sprintf(format, t...))
}

func (f *FileLogger) Warn(v ...interface{}) {
	file, line, _ := getCaller(2)
	f.Warnf("%s:%d\t|%s", file, line, fmt.Sprint(v...))
}

func (f *FileLogger) VWarn(format string, t ...interface{}) {
	file, line, _ := getCaller(2)
	f.Warnf("%s:%d\t|%s", file, line, fmt.Sprintf(format, t...))
}

// 带context及调用函数名的Info
func (f *FileLogger) CInfo(ctx context.Context, format string, t ...interface{}) {
	file, line, fn := getCaller(2)
	if f.stripFilePrefix != "" {
		file = strings.TrimPrefix(file, f.stripFilePrefix)
	}
	if f.stripFuncPrefix != "" {
		fn = strings.TrimPrefix(fn, f.stripFuncPrefix)
	}
	if ctx != nil {
		traceId, ok := ctx.Value(gcontext.TraceId).(string)
		if ok {
			fn = traceId + "_" + fn
		}
	}
	f.Infof(fmt.Sprintf("%s:%d\t%s\t|%s", file, line, fn, format), t...)
}

// 带context及调用函数名的Debug
func (f *FileLogger) CDebug(ctx context.Context, format string, t ...interface{}) {
	file, line, fn := getCaller(2)
	if f.stripFilePrefix != "" {
		file = strings.TrimPrefix(file, f.stripFilePrefix)
	}
	if f.stripFuncPrefix != "" {
		fn = strings.TrimPrefix(fn, f.stripFuncPrefix)
	}
	if ctx != nil {
		traceId, ok := ctx.Value(gcontext.TraceId).(string)
		if ok {
			fn = traceId + "_" + fn
		}
	}
	f.Debugf(fmt.Sprintf("%s:%d\t%s\t|%s", file, line, fn, format), t...)
}

// 带context及调用函数名的Warn
func (f *FileLogger) CWarn(ctx context.Context, format string, t ...interface{}) {
	file, line, fn := getCaller(2)
	if f.stripFilePrefix != "" {
		file = strings.TrimPrefix(file, f.stripFilePrefix)
	}
	if f.stripFuncPrefix != "" {
		fn = strings.TrimPrefix(fn, f.stripFuncPrefix)
	}
	if ctx != nil {
		traceId, ok := ctx.Value(gcontext.TraceId).(string)
		if ok {
			fn = traceId + "_" + fn
		}
	}
	f.Warnf(fmt.Sprintf("%s:%d\t%s\t|%s", file, line, fn, format), t...)
}

// 带context及调用函数名的Fatal
func (f *FileLogger) CFatal(ctx context.Context, format string, t ...interface{}) {
	file, line, fn := getCaller(2)
	if f.stripFilePrefix != "" {
		file = strings.TrimPrefix(file, f.stripFilePrefix)
	}
	if f.stripFuncPrefix != "" {
		fn = strings.TrimPrefix(fn, f.stripFuncPrefix)
	}
	if ctx != nil {
		traceId, ok := ctx.Value(gcontext.TraceId).(string)
		if ok {
			fn = traceId + "_" + fn
		}
	}
	f.Errorf(fmt.Sprintf("%s:%d\t%s\t|%s", file, line, fn, format), t...)
}

// 本地输出 console
func InitLocalLog() {
	Init()
	encoderCfg := zapcore.EncoderConfig{
		TimeKey:        "T",
		LevelKey:       "L",
		NameKey:        "N",
		CallerKey:      "C",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	core := zapcore.NewCore(zapcore.NewJSONEncoder(encoderCfg), os.Stdout, zap.DebugLevel)
	Lzap = zap.New(core).WithOptions()
}

// 开发测试
func DebugFileLogger(fl *FileLogger) {
	InitLocalLog()
	fl.zapLogger = Lzap
}
