package logs

import (
	"encoding/json"
	"fmt"
	"io"
	"kgogame/util/toml"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type LogConfigContainer struct {
	LogCfg *LogConfig `json:"Log" toml:"Log"`
}

var initOnce sync.Once

// 从JSON或者toml配置文件加载数据
func InitializeLog(serviceName string, serviceId int32, cfgFilePath string) error {
	if cfgFilePath == "" {
		return fmt.Errorf("config file path is empty")
	}
	var (
		err error
	)
	initOnce.Do(func() {
		data, lErr := Load(cfgFilePath) //兼容toml和json文件
		if lErr != nil {
			err = lErr
			return
		}
		if data.LogCfg == nil {
			err = fmt.Errorf("parse file '%s' data.Log is nil", cfgFilePath)
			return
		}
		InitLog(serviceName, serviceId, data.LogCfg)
	})
	return err
}

// 重载配置
func ReloadLogConfig(serviceName string, serviceId int32, cfgFilePath string) error {
	if cfgFilePath == "" {
		return fmt.Errorf("config file path is empty")
	}
	data, lErr := Load(cfgFilePath) //兼容toml和json文件
	if lErr != nil {
		return lErr
	}
	if data.LogCfg == nil {
		return fmt.Errorf("parse file '%s' data.Log is nil", cfgFilePath)
	}
	ReloadLog(serviceName, serviceId, data.LogCfg)
	return nil
}

func InitLogWithPid(serviceName string, serviceId int32, cfg *LogConfig) {
	initLogOnce.Do(func() {
		isInitLogOnce = true
		li := GetInstance()

		lc := &LogConfig{
			Level:   0, //debug
			FileNum: 30,
			FileMax: 30 * 1024 * 1024, // 30M
			LogPath: "/data/logs/phgames/services",
			UdpAddr: "127.0.0.1:4150",
			MaxAge:  0, // 天
		}

		lvlName := "debug"
		if cfg != nil {
			if _, ok := lvlMap[cfg.Level]; ok {
				lc.Level = cfg.Level
			}
			if cfg.FileNum > 0 {
				lc.FileNum = cfg.FileNum
			}
			if cfg.FileMax > 0 {
				lc.FileMax = cfg.FileMax
			}
			if cfg.LogPath != "" {
				lc.LogPath = cfg.LogPath
			}
			if cfg.UdpAddr != "" {
				lc.UdpAddr = cfg.UdpAddr
			}
			if cfg.MaxAge > 0 {
				lc.MaxAge = cfg.MaxAge
			}
		}

		lvlName = strings.ToLower(lvlMap[lc.Level])
		pid := os.Getpid()

		logPath := fmt.Sprintf("%s/%s_%d.%s.%d.log", lc.LogPath, serviceName, serviceId, lvlName, pid)
		li.Initlog(logPath, lc.Level, lc.FileNum, lc.FileMax)
		li.SetDayLog(lc.MaxAge)
		li.SetFullFnName(true)
		// 设置日志报送
		SetUdpReport(lc.UdpAddr, serviceName)
		fmt.Printf("log config path=%s level=%d filenum=%d filemax=%d\n", logPath, lc.Level, lc.FileNum, lc.FileMax)
	})
}

// 加载配置文件
func Load(configFile string) (*LogConfigContainer, error) {
	ext := filepath.Ext(configFile)
	if ext == ".toml" {
		return loadByToml(configFile)
	} else {
		return loadByJson(configFile)
	}
}

func loadByToml(configFile string) (*LogConfigContainer, error) {
	if ok, err := DirectoryExists(configFile); err != nil {
		return nil, fmt.Errorf("open file '%s' error: %s\n", configFile, err.Error())
	} else if !ok {
		return nil, fmt.Errorf("open file '%s' error: not exist\n", configFile)
	} else {
		data := LogConfigContainer{}
		if _, err = toml.DecodeFile(configFile, &data); err != nil {
			return nil, fmt.Errorf("parse file '%s' error: %s", configFile, err.Error())
		} else {
			return &data, nil
		}
	}
}

func loadByJson(configFile string) (*LogConfigContainer, error) {
	//打开文件
	if f, err := os.OpenFile(configFile, os.O_RDWR, 0644); err != nil {
		return nil, err
	} else {
		//操作完毕之后自动关闭句柄
		defer f.Close()
		//读文件
		if content, err := io.ReadAll(f); err != nil {
			return nil, err
		} else {
			//解析文件
			cfg := LogConfigContainer{}
			if err = json.Unmarshal(content, &cfg); err != nil {
				return nil, err
			} else {
				return &cfg, nil
			}
		}
	}
}

// 判断文件夹是否存在
func DirectoryExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	} else {
		return true, nil
	}
}
