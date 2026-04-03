package plog

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"kgogame/util/plog/lumberjack"
	"os"
	"path/filepath"
)

// 初始化本地文件日志
func InitLocalFileLog() {
	Init()

	var (
		topic    = ini_parser.GetString("", "topic")
		filePath = ini_parser.GetString("", "file_path")
		maxSize  = int(ini_parser.GetInt32("", "max_size"))
		lvl      = ini_parser.GetString("", "level")
	)

	if topic == "" { // 沿用旧配置
		fmt.Println("topic empty")
		return
	}
	if filePath == "" { // 默认日志目录
		filePath = "/data/logs/phgames/plog/"
	}
	if maxSize == 0 { // 单文件大小
		maxSize = MaxSize
	}

	filename := fmt.Sprintf("%s_%d", topic, os.Getpid())
	lg := &lumberjack.Logger{
		Filename:  filepath.Join(filePath, filename),
		MaxSize:   maxSize,
		MaxAge:    7, // days
		LocalTime: true,
	}

	//用Zap格式文件库
	zapLevel = zap.NewAtomicLevelAt(getLogLevel(lvl).(zapcore.Level))
	w := zapcore.AddSync(lg)
	ec := NewEncoderConfig()
	ec.EncodeTime = TimeEncoderMs
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(ec),
		w,
		zapLevel,
	)

	Lzap = zap.New(core)
}
