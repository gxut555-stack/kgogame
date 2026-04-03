package logs

import (
	"encoding/json"
	"log"
	"os"
)

func NewLogConfig(cfgFilePath string) *LogConfig {
	f, err := os.OpenFile(cfgFilePath, os.O_RDONLY, 0644)
	if err != nil {
		log.Printf("open file '%s' error: %s\n", cfgFilePath, err)
		return nil
	}
	defer f.Close()

	data, dec := LogConfigContainer{}, json.NewDecoder(f)
	if err = dec.Decode(&data); err != nil {
		log.Printf("parse file '%s' error: %s\n", cfgFilePath, err)
		return nil
	}
	return data.LogCfg
}
