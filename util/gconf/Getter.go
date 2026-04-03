package gconf

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

func GetStringConf(name string) string {
	if item, ok := gConfCli.Item(name); ok {
		return item.Value
	} else {
		return ""
	}
}

// 直接从apollo获取配置
func GetApolloConf(name string) string {
	return GetStringConf(name)
}

// 检查某个接口是否需要进行并发安全的控制
func IsAPINeedConcurentCheck(name string) bool {
	return false
}

// 检查某个接口是否为黑名单
func IsAPIBlack(svc string, fun string) bool {
	return strings.HasPrefix(svc, "Model") || strings.HasSuffix(fun, "_")
}

func GetRedisPoolConf(name string) *RedisConf {
	apolloConfLock.RLock()
	ob, ok := redisPoolConf[name]
	apolloConfLock.RUnlock()
	if !ok {
		fmt.Println("the GetRedisPoolConf not exists!", name)
	}
	return ob
}

func GetDBConf(name string) *DBConf {
	//fmt.Printf("the DB name = %s\n",name)
	apolloConfLock.RLock()
	ob, ok := mysqlConf[name]
	apolloConfLock.RUnlock()
	if !ok {
		fmt.Println("the GetDBConf not exists!", name)
	}
	return ob
}

func GetZKSRC() string {
	return GetStringConf("ZK.src")
}

func GetZkSRCSlice() (theSlice []string) {
	//zk的ip端口以逗号分隔
	str := GetZKSRC()
	if len(str) > 0 {
		theSlice = strings.Split(str, ",")
	}
	//fmt.Printf("zk服务目标地址len=%d, info=%#v \n\r", len(theSlice), theSlice)
	if len(theSlice) <= 0 {
		return []string{""}
	}
	return theSlice
}

func GetZKBasePath() string {
	return GetStringConf("ZK.BasePath")
}

func GetRedisSrc() string {
	return GetStringConf("redis.src")
}

func GetNsqSrc() string {
	return GetStringConf("nsq.nsqd.n1")
}

func GetNSQLookupd() string {
	return GetStringConf("nsq.nslookupd.lk1")
}

// 获取go-fds文件上传地址 格式：http://172.16.0.172:8087/upload
func GetFDSUploadPath() string {
	return GetStringConf("image.fds.upload")
}

// 获取go-fds文件上传外网主机地址 格式：http://47.92.231.60:8077/
func GetFDSUploadOuterHost() string {
	return GetStringConf("image.fds.upload.outer")
}

// 获取项目标识
func GetProject() string {
	return GetStringConf("project")
}

// 是否巴西项目
func IsBRProject() bool {
	return GetProject() == "BR"
}

// 是否印度项目
func IsINProject() bool {
	return GetProject() == "IN"
}

// 是否Zoo项目
func IsZooProject() bool {
	return GetProject() == "Zoo"
}

// 是否菲律宾项目
func IsPhilippinesProject() bool {
	return GetProject() == "PH"
}

// 获取项目名称
func GetProjectName() string {
	name := GetStringConf("project_name")
	if name == "" {
		return "未配项目名称"
	}
	return name
}

// 获取mongodb 配置
func GetMongoDBConf(confName string) *MongoConfig {
	apolloConfLock.Lock()
	defer apolloConfLock.Unlock()
	if c, ok := mongoConf[confName]; ok {
		return c
	}
	configString := GetStringConf(confName)
	if configString == "" {
		return nil
	}
	var mc MongoConfig
	err := json.Unmarshal([]byte(configString), &mc)
	if err != nil {
		log.Printf("json unmarshal err:%v conf:%s", err, configString)
		return nil
	}
	mongoConf[confName] = &mc
	return &mc
}

// 获取kafka配置
func GetKafkaConf() string {
	return GetStringConf("kafka")
}
