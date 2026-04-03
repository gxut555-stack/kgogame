package logs

import (
	"fmt"
	"strings"
	"sync"
)

type LogConfig struct {
	Level   int    `json:"Level" toml:"Level"`
	FileNum int    `json:"FileNum" toml:"FileNum"`
	FileMax int64  `json:"FileMax" toml:"FileMax"`
	LogPath string `json:"LogPath" toml:"LogPath"`
	UdpAddr string `json:"UdpAddr" toml:"UdpAddr"`
	MaxAge  int    `json:"MaxAge" toml:"MaxAge"`
}

var initLogOnce sync.Once
var isInitLogOnce bool

// 初始化日志 cfg==nil 时尝试使用InitializeLog
func InitLog(serviceName string, serviceId int32, cfg *LogConfig) {
	initLogOnce.Do(func() {
		isInitLogOnce = true
		li := GetInstance()
		lvlName := "debug"
		if cfg == nil {
			cfg = &LogConfig{
				Level:   0, //debug
				FileNum: 5,
				FileMax: 104857600,
				LogPath: ".",
				UdpAddr: "",
			}
		}
		if n, ok := lvlMap[cfg.Level]; ok {
			lvlName = strings.ToLower(n)
		}
		logPath := fmt.Sprintf("%s/%s.%d.%s.log", cfg.LogPath, serviceName, serviceId, lvlName)
		li.Initlog(logPath, cfg.Level, cfg.FileNum, cfg.FileMax)
		// 设置日志报送
		SetUdpReport(cfg.UdpAddr, serviceName)
		fmt.Printf("log config path=%s level=%d filenum=%d filemax=%d\n", logPath, cfg.Level, cfg.FileNum, cfg.FileMax)
	})
}

func ReloadLog(serviceName string, serviceId int32, cfg *LogConfig) {
	li := GetInstance()
	lvlName := "debug"
	if n, ok := lvlMap[cfg.Level]; ok {
		lvlName = strings.ToLower(n)
	}
	logPath := fmt.Sprintf("%s/%s.%d.%s.log", cfg.LogPath, serviceName, serviceId, lvlName)
	li.Reload(logPath, cfg.Level, cfg.FileNum, cfg.FileMax)
	fmt.Printf("reload log config path=%s level=%d filenum=%d filemax=%d\n", logPath, cfg.Level, cfg.FileNum, cfg.FileMax)
}

// 是否初始化过logs日志库
func IsInitLogsLib() bool {
	return isInitLogOnce
}
