package main

import (
	"fmt"
	"kgogame/misc/common"
	"kgogame/util/toml"
	"time"
)

type WebSocketConfig struct {
	Tls bool   `toml:"Tls"`
	Key string `toml:"Key"`
	Pem string `toml:"Pem"`
}

type Config struct {
	cfgPath    string          //配置路径
	WsLoadTime int64           //加载时间
	WebSocket  WebSocketConfig `toml:"WebSocket"`
}

var pConfig *Config

func init() {
	pConfig = &Config{
		cfgPath:    "",
		WsLoadTime: 0,
	}
}

func (pCfg *Config) InitConfig(cfgPath string) error {

	if cfgPath == "" {
		return fmt.Errorf("cfgPath is empty")
	}
	pCfg.cfgPath = cfgPath
	if ok, err := common.FileExists(cfgPath); err != nil {
		return fmt.Errorf("open file '%s' error: %s\n", cfgPath, err.Error())
	} else if !ok {
		return fmt.Errorf("open file '%s' error: not exist\n", cfgPath)
	} else {
		data := Config{}
		if _, err = toml.DecodeFile(pCfg.cfgPath, &data); err != nil {
			return fmt.Errorf("parse file '%s' error: %s", pCfg.cfgPath, err.Error())
		} else {
			pCfg.WsLoadTime = time.Now().Unix()
			pCfg.WebSocket = data.WebSocket
			fmt.Printf("config:%s", common.Data2Json(pCfg))
			return nil
		}
	}
}

func (pCfg *Config) ReloadConfig() error {
	data := Config{}
	if _, err := toml.DecodeFile(pCfg.cfgPath, &data); err != nil {
		return fmt.Errorf("parse file '%s' error: %s", pCfg.cfgPath, err.Error())
	} else {
		pCfg.WsLoadTime = time.Now().Unix()
		pCfg.WebSocket = data.WebSocket
		fmt.Printf("reload config:%s", common.Data2Json(pCfg))
		return nil
	}
}
