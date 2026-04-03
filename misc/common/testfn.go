package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"kgogame/util/gconf"
)

func InitTestFn(platform string) {
	err := gconf.InitApolloConf("tcp://192.168.6.16:8060/" + platform + "games")
	if err != nil {
		panic(err)
	}
}

func InitProdFn(platform string) {
	err := gconf.InitApolloConf("tcp://10.0.0.167:8060/" + platform + "games")
	if err != nil {
		panic(err)
	}
}

func InitTestFileLog() {
	//不在提供初始化日志文件
}

func Data2Json(data interface{}, pretty ...bool) string {
	if data == nil {
		return ""
	}
	if len(pretty) > 0 {
		byt, _ := json.MarshalIndent(data, "", "  ")
		return string(byt)
	}
	byt, _ := json.Marshal(data)
	return string(byt)
}

func Data2JsonWithCompress(data interface{}) (string, error) {
	if data == nil {
		return "", nil
	}

	byt, _ := json.Marshal(data)
	var buf bytes.Buffer
	if err := Compress(&buf, byt); err != nil {
		return "", fmt.Errorf("compress failed %s", err)
	}

	return buf.String(), nil
}
